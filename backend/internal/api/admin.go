package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/remotescope/backend/internal/models"
	"github.com/remotescope/backend/internal/repository"
	"gorm.io/gorm"
)

type AdminHandler struct {
	db       *gorm.DB
	jobRepo  *repository.JobRepository
	userRepo *repository.UserRepository
	rdb      *redis.Client
}

func NewAdminHandler(db *gorm.DB, jobRepo *repository.JobRepository, userRepo *repository.UserRepository, rdb *redis.Client) *AdminHandler {
	return &AdminHandler{db: db, jobRepo: jobRepo, userRepo: userRepo, rdb: rdb}
}

// GET /api/v1/admin/stats
func (h *AdminHandler) Stats(c *gin.Context) {
	var stats struct {
		TotalUsers     int64 `json:"total_users"`
		TotalJobs      int64 `json:"total_jobs"`
		ActiveJobs     int64 `json:"active_jobs"`
		TotalCompanies int64 `json:"total_companies"`
		JobsToday      int64 `json:"jobs_today"`
		JobsThisHour   int64 `json:"jobs_this_hour"`
		TotalSources   int64 `json:"total_sources"`
		ActiveSources  int64 `json:"active_sources"`
	}

	h.db.Model(&models.User{}).Count(&stats.TotalUsers)
	h.db.Model(&models.Job{}).Count(&stats.TotalJobs)
	h.db.Model(&models.Job{}).Where("is_active = true").Count(&stats.ActiveJobs)
	h.db.Model(&models.Company{}).Where("is_blacklisted = false").Count(&stats.TotalCompanies)
	h.db.Model(&models.Job{}).Where("first_seen_at >= ?", time.Now().Truncate(24*time.Hour)).Count(&stats.JobsToday)
	h.db.Model(&models.Job{}).Where("first_seen_at >= ?", time.Now().Add(-time.Hour)).Count(&stats.JobsThisHour)
	h.db.Model(&models.JobSource{}).Count(&stats.TotalSources)
	h.db.Model(&models.JobSource{}).Where("is_active = true").Count(&stats.ActiveSources)

	c.JSON(http.StatusOK, models.Success(stats))
}

// GET /api/v1/admin/scrapers
func (h *AdminHandler) ListScrapers(c *gin.Context) {
	var sources []models.JobSource
	h.db.Order("name ASC").Find(&sources)
	c.JSON(http.StatusOK, models.Success(sources))
}

// POST /api/v1/admin/scrapers/:id/run
func (h *AdminHandler) TriggerScraper(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResp("invalid scraper ID"))
		return
	}

	var source models.JobSource
	if err := h.db.First(&source, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResp("scraper not found"))
		return
	}

	// Publish trigger message to Redis scrape queue
	if h.rdb != nil {
		payload, _ := json.Marshal(map[string]string{
			"source_slug": source.Slug,
			"source_id":   source.ID.String(),
			"triggered_by": "admin",
		})
		if pubErr := h.rdb.RPush(context.Background(), "q:scrape", payload).Err(); pubErr != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResp("failed to enqueue scraper"))
			return
		}
	}

	c.JSON(http.StatusOK, models.SuccessMsg("scraper triggered: "+source.Name))
}

// GET /api/v1/admin/scraper-runs
func (h *AdminHandler) ScraperRuns(c *gin.Context) {
	var runs []models.ScraperRun
	h.db.Preload("Source").
		Order("started_at DESC").
		Limit(100).
		Find(&runs)
	c.JSON(http.StatusOK, models.Success(runs))
}

// GET /api/v1/admin/users
func (h *AdminHandler) ListUsers(c *gin.Context) {
	page := 1
	perPage := 50
	users, total, err := h.userRepo.List(page, perPage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResp("failed to fetch users"))
		return
	}
	c.JSON(http.StatusOK, models.Success(gin.H{"data": users, "total": total}))
}

// PUT /api/v1/admin/users/:id/role
func (h *AdminHandler) UpdateUserRole(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResp("invalid user ID"))
		return
	}
	var body struct {
		Role models.UserRole `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResp(err.Error()))
		return
	}
	h.db.Model(&models.User{}).Where("id = ?", id).Update("role", body.Role)
	c.JSON(http.StatusOK, models.SuccessMsg("user role updated"))
}

// POST /api/v1/admin/companies/:id/blacklist
func (h *AdminHandler) BlacklistCompany(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResp("invalid company ID"))
		return
	}
	var body struct {
		Reason string `json:"reason"`
	}
	c.ShouldBindJSON(&body)
	h.db.Model(&models.Company{}).Where("id = ?", id).Updates(map[string]interface{}{
		"is_blacklisted":   true,
		"blacklist_reason": body.Reason,
	})
	c.JSON(http.StatusOK, models.SuccessMsg("company blacklisted"))
}

// DELETE /api/v1/admin/jobs/:id
func (h *AdminHandler) DeleteJob(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResp("invalid job ID"))
		return
	}
	h.db.Model(&models.Job{}).Where("id = ?", id).Update("is_active", false)
	c.JSON(http.StatusOK, models.SuccessMsg("job deactivated"))
}

// POST /api/v1/admin/jobs/:id/flag-fake
func (h *AdminHandler) FlagFakeJob(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResp("invalid job ID"))
		return
	}
	h.db.Model(&models.Job{}).Where("id = ?", id).Updates(map[string]interface{}{
		"is_fake_suspected": true,
		"is_active":         false,
	})
	c.JSON(http.StatusOK, models.SuccessMsg("job flagged as fake"))
}
