package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/remotescope/notification_service/internal/dispatcher"
	"go.uber.org/zap"
)

const (
	queueNotify     = "q:notify"
	workerConcurrency = 5
)

// NotificationPayload is the message format published by the AI worker.
type NotificationPayload struct {
	Type     string   `json:"type"`
	JobID    string   `json:"job_id"`
	Title    string   `json:"title"`
	Company  string   `json:"company"`
	Category string   `json:"category"`
	Skills   []string `json:"skills"`
}

func main() {
	// ── Logger ───────────────────────────────────────────────────
	log, _ := zap.NewProduction()
	defer log.Sync()
	sugar := log.Sugar()

	// ── Config ───────────────────────────────────────────────────
	redisURL := getEnv("REDIS_URL", "redis://localhost:6379/0")
	port := getEnv("NOTIFICATION_SERVICE_PORT", "8003")

	// ── Redis ────────────────────────────────────────────────────
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		sugar.Fatalf("Invalid REDIS_URL: %v", err)
	}
	rdb := redis.NewClient(opt)
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		sugar.Fatalf("Redis connection failed: %v", err)
	}
	sugar.Info("✅ Connected to Redis")

	// ── Dispatcher ──────────────────────────────────────────────
	disp := dispatcher.New(dispatcher.Config{
		ResendAPIKey:    getEnv("RESEND_API_KEY", ""),
		EmailFrom:       getEnv("EMAIL_FROM", "notifications@remotescope.app"),
		EmailFromName:   getEnv("EMAIL_FROM_NAME", "RemoteScope"),
		VAPIDPublicKey:  getEnv("VAPID_PUBLIC_KEY", ""),
		VAPIDPrivateKey: getEnv("VAPID_PRIVATE_KEY", ""),
		VAPIDSubject:    getEnv("VAPID_SUBJECT", "mailto:contact@remotescope.app"),
		TelegramToken:   getEnv("TELEGRAM_BOT_TOKEN", ""),
		DiscordWebhook:  getEnv("DISCORD_WEBHOOK_URL", ""),
		SlackWebhook:    getEnv("SLACK_WEBHOOK_URL", ""),
		EnableEmail:     getEnv("ENABLE_EMAIL_NOTIFICATIONS", "true") == "true",
		EnablePush:      getEnv("ENABLE_PUSH_NOTIFICATIONS", "true") == "true",
		EnableTelegram:  getEnv("ENABLE_TELEGRAM", "false") == "true",
		EnableDiscord:   getEnv("ENABLE_DISCORD", "false") == "true",
		EnableSlack:     getEnv("ENABLE_SLACK", "false") == "true",
	}, sugar)

	// ── Health HTTP server ────────────────────────────────────────
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok","service":"remotescope-notifications"}`)
	})
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	go func() {
		sugar.Infof("🔔 Notification service listening on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			sugar.Errorf("HTTP server error: %v", err)
		}
	}()

	// ── Worker goroutines ─────────────────────────────────────────
	sem := make(chan struct{}, workerConcurrency)
	go func() {
		sugar.Info("Worker started — consuming q:notify")
		for {
			result, err := rdb.BLPop(ctx, 5*time.Second, queueNotify).Result()
			if err != nil {
				if err == redis.Nil {
					continue // timeout, loop
				}
				if ctx.Err() != nil {
					return // shutting down
				}
				sugar.Errorf("BLPop error: %v", err)
				time.Sleep(time.Second)
				continue
			}

			data := result[1]
			sem <- struct{}{}
			go func(payload string) {
				defer func() { <-sem }()
				var msg NotificationPayload
				if err := json.Unmarshal([]byte(payload), &msg); err != nil {
					sugar.Warnf("Failed to parse notification payload: %v", err)
					return
				}
				disp.Dispatch(ctx, msg.Type, msg.JobID, msg.Title, msg.Company, msg.Category, msg.Skills)
			}(data)
		}
	}()

	// ── Graceful shutdown ─────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	sugar.Info("Shutting down notification service...")

	shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(shutCtx)
	rdb.Close()
	sugar.Info("Notification service stopped")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
