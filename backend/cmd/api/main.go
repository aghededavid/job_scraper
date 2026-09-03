package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/remotescope/backend/internal/api"
	"github.com/remotescope/backend/internal/auth"
	"github.com/remotescope/backend/internal/repository"
	"github.com/remotescope/backend/internal/websocket"
	"github.com/remotescope/backend/pkg/config"
	limiter "github.com/ulule/limiter/v3"
	ginlimiter "github.com/ulule/limiter/v3/drivers/middleware/gin"
	redisstore "github.com/ulule/limiter/v3/drivers/store/redis"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// @title RemoteScope API
// @version 1.0
// @description Production-ready remote job aggregation platform API
// @termsOfService https://remotescope.app/terms
// @contact.name API Support
// @contact.email api@remotescope.app
// @license.name MIT
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// ── CLI subcommands ──────────────────────────────────────────
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "migrate":
			runMigrate()
			return
		case "seed":
			runSeed()
			return
		}
	}

	// ── Config ──────────────────────────────────────────────────
	cfg := config.Load()

	// ── Logger ──────────────────────────────────────────────────
	var log *zap.Logger
	if cfg.App.Env == "production" {
		log, _ = zap.NewProduction()
	} else {
		log, _ = zap.NewDevelopment()
	}
	defer log.Sync()
	sugar := log.Sugar()

	// ── Database ─────────────────────────────────────────────────
	gormLog := logger.Default.LogMode(logger.Silent)
	if cfg.App.Env == "development" {
		gormLog = logger.Default.LogMode(logger.Info)
	}

	db, err := gorm.Open(postgres.Open(cfg.Database.URL), &gorm.Config{
		Logger: gormLog,
	})
	if err != nil {
		sugar.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)
	sugar.Info("✅ Connected to PostgreSQL")

	// ── Redis ─────────────────────────────────────────────────────
	rdb, redisErr := connectRedis(cfg, sugar)
	if redisErr != nil {
		sugar.Warnf("Redis unavailable (%v) — rate limiting disabled", redisErr)
	} else {
		sugar.Info("✅ Connected to Redis")
	}

	// ── Services ─────────────────────────────────────────────────
	jwtSvc := auth.NewJWTService(cfg)
	oauthSvc := auth.NewOAuthService(cfg)

	// ── Repositories ─────────────────────────────────────────────
	userRepo := repository.NewUserRepository(db)
	jobRepo := repository.NewJobRepository(db)
	savedRepo := repository.NewSavedJobRepository(db)
	companyRepo := repository.NewCompanyRepository(db)
	notifRepo := repository.NewNotificationRepository(db)

	// ── Handlers ─────────────────────────────────────────────────
	authHandler := api.NewAuthHandler(userRepo, jwtSvc, oauthSvc, cfg.Security.BcryptCost)
	jobsHandler := api.NewJobsHandler(jobRepo)
	searchHandler := api.NewSearchHandler(jobRepo, cfg)
	usersHandler := api.NewUsersHandler(userRepo)
	companiesHandler := api.NewCompaniesHandler(companyRepo)
	notifsHandler := api.NewNotificationsHandler(notifRepo)
	adminHandler := api.NewAdminHandler(db, jobRepo, userRepo, rdb)

	// ── WebSocket Hub ─────────────────────────────────────────────
	hub := websocket.NewHub()
	go hub.Run()

	// ── Router ───────────────────────────────────────────────────
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(ginLogger(sugar))

	// CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.App.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"X-Total-Count"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// ── Rate Limiting ─────────────────────────────────────────────
	var globalLimiter gin.HandlerFunc
	var searchLimiter gin.HandlerFunc
	if rdb != nil {
		store, storeErr := redisstore.NewStoreWithOptions(rdb, limiter.StoreOptions{
			Prefix:   "rs_ratelimit",
			MaxRetry: 3,
		})
		if storeErr == nil {
			globalRate := limiter.Rate{
				Period: time.Minute,
				Limit:  int64(cfg.RateLimit.RequestsPerMinute),
			}
			searchRate := limiter.Rate{
				Period: time.Minute,
				Limit:  int64(cfg.RateLimit.SearchPerMinute),
			}
			globalLimiter = ginlimiter.NewMiddleware(limiter.New(store, globalRate))
			searchLimiter = ginlimiter.NewMiddleware(limiter.New(store, searchRate))
			sugar.Infof("✅ Rate limiting enabled (%d req/min global, %d req/min search)",
				cfg.RateLimit.RequestsPerMinute, cfg.RateLimit.SearchPerMinute)
		}
	}

	// Apply global rate limit
	if globalLimiter != nil {
		r.Use(globalLimiter)
	}

	// Health
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "remotescope-api"})
	})

	// WebSocket
	r.GET("/ws", func(c *gin.Context) {
		websocket.ServeWS(hub, jwtSvc, c.Writer, c.Request)
	})

	// ── API v1 ───────────────────────────────────────────────────
	v1 := r.Group("/api/v1")

	// Auth routes (public)
	authGroup := v1.Group("/auth")
	{
		authGroup.POST("/signup", authHandler.Signup)
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/refresh", authHandler.Refresh)
		authGroup.POST("/logout", authHandler.Logout)
		authGroup.GET("/oauth/:provider", authHandler.OAuthRedirect)
		authGroup.GET("/oauth/:provider/callback", authHandler.OAuthCallback)
	}

	// Job routes (public + optional auth)
	optAuth := auth.OptionalAuth(jwtSvc)
	jobsGroup := v1.Group("/jobs", optAuth)
	{
		jobsGroup.GET("", jobsHandler.List)
		jobsGroup.GET("/latest", jobsHandler.Latest)
		jobsGroup.GET("/trending", jobsHandler.Trending)
		jobsGroup.GET("/:id", jobsHandler.Get)
	}

	// Recommendations (optional auth)
	v1.GET("/recommendations", optAuth, jobsHandler.Recommendations)

	// Search (public, stricter rate limit)
	searchGroup := v1.Group("")
	if searchLimiter != nil {
		searchGroup.Use(searchLimiter)
	}
	searchGroup.POST("/search", optAuth, searchHandler.Search)
	searchGroup.GET("/search/suggest", searchHandler.Suggest)

	// Companies (public)
	v1.GET("/companies", companiesHandler.List)
	v1.GET("/companies/:slug", companiesHandler.Get)

	// Categories (public)
	v1.GET("/categories", func(c *gin.Context) {
		var cats []struct {
			Name  string `json:"name"`
			Slug  string `json:"slug"`
			Icon  string `json:"icon"`
			Color string `json:"color"`
		}
		db.Raw("SELECT name, slug, icon, color FROM categories ORDER BY sort_order").Scan(&cats)
		c.JSON(http.StatusOK, gin.H{"success": true, "data": cats})
	})

	// Protected routes
	protected := v1.Group("", auth.RequireAuth(jwtSvc))
	{
		// User profile
		protected.GET("/users/me", usersHandler.GetMe)
		protected.PUT("/users/me", usersHandler.UpdateMe)
		protected.PUT("/users/me/preferences", usersHandler.UpdatePreferences)
		protected.POST("/users/me/complete-onboarding", usersHandler.CompleteOnboarding)

		// Saved jobs
		protected.GET("/saved-jobs", func(c *gin.Context) {
			userID, _ := auth.GetUserID(c)
			page := 1
			perPage := 20
			saved, total, _ := savedRepo.ListSaved(userID, page, perPage)
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data":    gin.H{"items": saved, "total": total},
			})
		})
		protected.POST("/jobs/:id/save", jobsHandler.SaveJob(jobRepo, savedRepo))
		protected.DELETE("/jobs/:id/save", jobsHandler.UnsaveJob(savedRepo))
		protected.POST("/jobs/:id/hide", jobsHandler.HideJob(savedRepo))
		protected.POST("/jobs/:id/apply", jobsHandler.ApplyJob(savedRepo))

		// Applied jobs
		protected.GET("/applied-jobs", func(c *gin.Context) {
			userID, _ := auth.GetUserID(c)
			applied, total, _ := savedRepo.ListApplied(userID, 1, 20)
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data":    gin.H{"items": applied, "total": total},
			})
		})

		// Notifications
		protected.GET("/notifications", notifsHandler.List)
		protected.PUT("/notifications/:id/read", notifsHandler.MarkRead)
		protected.PUT("/notifications/read-all", notifsHandler.MarkAllRead)

		// Saved searches
		protected.GET("/saved-searches", notifsHandler.ListSavedSearches)
		protected.POST("/saved-searches", notifsHandler.CreateSavedSearch)
		protected.DELETE("/saved-searches/:id", notifsHandler.DeleteSavedSearch)
	}

	// Admin routes
	admin := v1.Group("/admin", auth.RequireAdmin(jwtSvc))
	{
		admin.GET("/stats", adminHandler.Stats)
		admin.GET("/scrapers", adminHandler.ListScrapers)
		admin.POST("/scrapers/:id/run", adminHandler.TriggerScraper)
		admin.GET("/scraper-runs", adminHandler.ScraperRuns)
		admin.GET("/users", adminHandler.ListUsers)
		admin.PUT("/users/:id/role", adminHandler.UpdateUserRole)
		admin.POST("/companies/:id/blacklist", adminHandler.BlacklistCompany)
		admin.DELETE("/jobs/:id", adminHandler.DeleteJob)
		admin.POST("/jobs/:id/flag-fake", adminHandler.FlagFakeJob)
	}

	// ── Start Server ─────────────────────────────
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.App.Port),
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		sugar.Infof("🚀 RemoteScope API running on port %d (env: %s)", cfg.App.Port, cfg.App.Env)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			sugar.Fatalf("Server failed: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	sugar.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		sugar.Errorf("Server shutdown error: %v", err)
	}
	if rdb != nil {
		rdb.Close()
	}
	sugar.Info("Server stopped")
}

// ── CLI commands ──────────────────────────────────────────────────

func runMigrate() {
	fmt.Println("Running database migrations...")
	cfg := config.Load()
	db, err := gorm.Open(postgres.Open(cfg.Database.URL), &gorm.Config{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "DB connection failed: %v\n", err)
		os.Exit(1)
	}
	// Migrations are handled by docker-entrypoint-initdb.d SQL files in dev.
	// In production, use golang-migrate or goose. This command validates connectivity.
	sqlDB, _ := db.DB()
	if err := sqlDB.Ping(); err != nil {
		fmt.Fprintf(os.Stderr, "Ping failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Database is reachable. SQL migrations are applied via migrations/001_initial_schema.sql on first start.")
}

func runSeed() {
	fmt.Println("Seeding demo data...")
	cfg := config.Load()
	db, err := gorm.Open(postgres.Open(cfg.Database.URL), &gorm.Config{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "DB connection failed: %v\n", err)
		os.Exit(1)
	}

	// Seed categories
	categories := []map[string]interface{}{
		{"name": "Frontend", "slug": "frontend", "icon": "🖥️", "color": "#3b82f6", "sort_order": 1},
		{"name": "Backend", "slug": "backend", "icon": "⚙️", "color": "#8b5cf6", "sort_order": 2},
		{"name": "Full Stack", "slug": "full-stack", "icon": "🔗", "color": "#06b6d4", "sort_order": 3},
		{"name": "DevOps & Cloud", "slug": "devops-cloud", "icon": "☁️", "color": "#10b981", "sort_order": 4},
		{"name": "AI & Machine Learning", "slug": "ai-ml", "icon": "🤖", "color": "#f59e0b", "sort_order": 5},
		{"name": "Data Science", "slug": "data-science", "icon": "📊", "color": "#ef4444", "sort_order": 6},
		{"name": "Mobile", "slug": "mobile", "icon": "📱", "color": "#ec4899", "sort_order": 7},
		{"name": "UI/UX Design", "slug": "ui-ux-design", "icon": "🎨", "color": "#f97316", "sort_order": 8},
		{"name": "Cybersecurity", "slug": "cybersecurity", "icon": "🔒", "color": "#dc2626", "sort_order": 9},
		{"name": "Product Management", "slug": "product-management", "icon": "📋", "color": "#7c3aed", "sort_order": 10},
	}
	for _, cat := range categories {
		db.Exec(`
			INSERT INTO categories (id, name, slug, icon, color, sort_order)
			VALUES (uuid_generate_v4(), ?, ?, ?, ?, ?)
			ON CONFLICT (slug) DO NOTHING
		`, cat["name"], cat["slug"], cat["icon"], cat["color"], cat["sort_order"])
	}

	// Seed job sources
	sources := []map[string]interface{}{
		{"name": "Remote OK", "slug": "remoteok", "adapter_class": "RemoteOKScraper", "base_url": "https://remoteok.com", "fetch_method": "api", "schedule_cron": "0 */6 * * *"},
		{"name": "Remotive", "slug": "remotive", "adapter_class": "RemotiveScraper", "base_url": "https://remotive.com", "fetch_method": "api", "schedule_cron": "30 */6 * * *"},
		{"name": "Jobicy", "slug": "jobicy", "adapter_class": "JobicyScraper", "base_url": "https://jobicy.com", "fetch_method": "api", "schedule_cron": "0 */8 * * *"},
		{"name": "We Work Remotely", "slug": "weworkremotely", "adapter_class": "WWRScraper", "base_url": "https://weworkremotely.com", "fetch_method": "rss", "schedule_cron": "15 */6 * * *"},
		{"name": "Greenhouse", "slug": "greenhouse", "adapter_class": "GreenhouseScraper", "base_url": "https://boards.greenhouse.io", "fetch_method": "api", "schedule_cron": "45 */12 * * *"},
	}
	for _, src := range sources {
		db.Exec(`
			INSERT INTO job_sources (id, name, slug, adapter_class, base_url, fetch_method, schedule_cron)
			VALUES (uuid_generate_v4(), ?, ?, ?, ?, ?, ?)
			ON CONFLICT (slug) DO NOTHING
		`, src["name"], src["slug"], src["adapter_class"], src["base_url"], src["fetch_method"], src["schedule_cron"])
	}

	fmt.Println("✅ Demo data seeded: categories and job sources")
}

// ── Helpers ───────────────────────────────────────────────────────

func connectRedis(cfg *config.Config, sugar *zap.SugaredLogger) (*redis.Client, error) {
	if cfg.Redis.URL == "" {
		return nil, fmt.Errorf("REDIS_URL not configured")
	}
	opt, err := redis.ParseURL(cfg.Redis.URL)
	if err != nil {
		return nil, err
	}
	rdb := redis.NewClient(opt)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return rdb, nil
}

func ginLogger(sugar *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		sugar.Infow("request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency", time.Since(start),
			"ip", c.ClientIP(),
		)
	}
}
