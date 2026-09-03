package dispatcher

import (
	"context"

	"github.com/remotescope/notification_service/internal/discord"
	"github.com/remotescope/notification_service/internal/email"
	"github.com/remotescope/notification_service/internal/push"
	"github.com/remotescope/notification_service/internal/slack"
	"github.com/remotescope/notification_service/internal/telegram"
	"go.uber.org/zap"
)

// Config holds all channel configuration flags and credentials.
type Config struct {
	// Email (Resend)
	ResendAPIKey  string
	EmailFrom     string
	EmailFromName string
	// Web Push (VAPID)
	VAPIDPublicKey  string
	VAPIDPrivateKey string
	VAPIDSubject    string
	// Telegram
	TelegramToken string
	// Discord
	DiscordWebhook string
	// Slack
	SlackWebhook string
	// Feature flags
	EnableEmail    bool
	EnablePush     bool
	EnableTelegram bool
	EnableDiscord  bool
	EnableSlack    bool
}

// Dispatcher routes a notification to all enabled channels.
type Dispatcher struct {
	cfg      Config
	log      *zap.SugaredLogger
	email    *email.Sender
	push     *push.Sender
	telegram *telegram.Sender
	discord  *discord.Sender
	slack    *slack.Sender
}

func New(cfg Config, log *zap.SugaredLogger) *Dispatcher {
	d := &Dispatcher{cfg: cfg, log: log}

	if cfg.EnableEmail && cfg.ResendAPIKey != "" {
		d.email = email.NewSender(cfg.ResendAPIKey, cfg.EmailFrom, cfg.EmailFromName)
	}
	if cfg.EnablePush && cfg.VAPIDPublicKey != "" {
		d.push = push.NewSender(cfg.VAPIDPublicKey, cfg.VAPIDPrivateKey, cfg.VAPIDSubject)
	}
	if cfg.EnableTelegram && cfg.TelegramToken != "" {
		d.telegram = telegram.NewSender(cfg.TelegramToken)
	}
	if cfg.EnableDiscord && cfg.DiscordWebhook != "" {
		d.discord = discord.NewSender(cfg.DiscordWebhook)
	}
	if cfg.EnableSlack && cfg.SlackWebhook != "" {
		d.slack = slack.NewSender(cfg.SlackWebhook)
	}
	return d
}

// Dispatch sends a notification to all enabled channels.
func (d *Dispatcher) Dispatch(ctx context.Context, notifType, jobID, title, company, category string, skills []string) {
	msg := buildMessage(notifType, title, company, category)

	if d.discord != nil {
		if err := d.discord.Send(ctx, msg); err != nil {
			d.log.Warnf("Discord send failed: %v", err)
		}
	}
	if d.slack != nil {
		if err := d.slack.Send(ctx, msg); err != nil {
			d.log.Warnf("Slack send failed: %v", err)
		}
	}
	if d.telegram != nil {
		if err := d.telegram.Send(ctx, msg); err != nil {
			d.log.Warnf("Telegram send failed: %v", err)
		}
	}
	// Email and push require recipient-level data from the DB —
	// those are handled by the backend's notification system which
	// uses the user's saved-search preferences. We log intent here.
	if d.email != nil {
		d.log.Debugf("Email notifications enabled — backend handles per-user dispatch for job %s", jobID)
	}
	if d.push != nil {
		d.log.Debugf("Push notifications enabled — backend handles per-user dispatch for job %s", jobID)
	}
}

func buildMessage(notifType, title, company, category string) string {
	switch notifType {
	case "new_job":
		return "🆕 New job: " + title + " @ " + company + " [" + category + "]"
	default:
		return "RemoteScope notification: " + notifType
	}
}
