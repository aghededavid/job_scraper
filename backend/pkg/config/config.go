package config

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App          AppConfig
	Database     DatabaseConfig
	Redis        RedisConfig
	JWT          JWTConfig
	OAuth        OAuthConfig
	AI           AIConfig
	Queue        QueueConfig
	RateLimit    RateLimitConfig
	Notification NotificationConfig
	Security     SecurityConfig
}

type AppConfig struct {
	Name            string
	Env             string
	Port            int
	URL             string
	AllowedOrigins  []string
}

type DatabaseConfig struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type RedisConfig struct {
	URL      string
	Password string
}

type JWTConfig struct {
	Secret              string
	ExpiryHours         int
	RefreshExpiryDays   int
}

type OAuthConfig struct {
	Google   OAuthProvider
	GitHub   OAuthProvider
	LinkedIn OAuthProvider
}

type OAuthProvider struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

type AIConfig struct {
	ServiceURL          string
	OpenAIKey           string
	EmbeddingModel      string
	ChatModel           string
}

type QueueConfig struct {
	Scrape      string
	Parse       string
	Classify    string
	Embed       string
	Deduplicate string
	Notify      string
	DeadLetter  string
}

type RateLimitConfig struct {
	RequestsPerMinute int
	SearchPerMinute   int
}

type NotificationConfig struct {
	ResendAPIKey    string
	EmailFrom       string
	EmailFromName   string
	VAPIDPublicKey  string
	VAPIDPrivateKey string
	VAPIDSubject    string
}

type SecurityConfig struct {
	BcryptCost int
}

func Load() *Config {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Config file not found, using environment variables: %v", err)
	}

	viper.SetDefault("BACKEND_PORT", 8080)
	viper.SetDefault("APP_ENV", "development")
	viper.SetDefault("JWT_EXPIRY_HOURS", 24)
	viper.SetDefault("JWT_REFRESH_EXPIRY_DAYS", 30)
	viper.SetDefault("RATE_LIMIT_REQUESTS_PER_MINUTE", 60)
	viper.SetDefault("RATE_LIMIT_SEARCH_PER_MINUTE", 30)
	viper.SetDefault("BCRYPT_COST", 12)
	viper.SetDefault("OPENAI_EMBEDDING_MODEL", "text-embedding-3-small")
	viper.SetDefault("OPENAI_CHAT_MODEL", "gpt-4o-mini")

	return &Config{
		App: AppConfig{
			Name: viper.GetString("APP_NAME"),
			Env:  viper.GetString("APP_ENV"),
			Port: viper.GetInt("BACKEND_PORT"),
			URL:  viper.GetString("APP_URL"),
			AllowedOrigins: viper.GetStringSlice("ALLOWED_ORIGINS"),
		},
		Database: DatabaseConfig{
			URL:             viper.GetString("DATABASE_URL"),
			MaxOpenConns:    25,
			MaxIdleConns:    5,
			ConnMaxLifetime: 5 * time.Minute,
		},
		Redis: RedisConfig{
			URL:      viper.GetString("REDIS_URL"),
			Password: viper.GetString("REDIS_PASSWORD"),
		},
		JWT: JWTConfig{
			Secret:            viper.GetString("JWT_SECRET"),
			ExpiryHours:       viper.GetInt("JWT_EXPIRY_HOURS"),
			RefreshExpiryDays: viper.GetInt("JWT_REFRESH_EXPIRY_DAYS"),
		},
		OAuth: OAuthConfig{
			Google: OAuthProvider{
				ClientID:     viper.GetString("GOOGLE_CLIENT_ID"),
				ClientSecret: viper.GetString("GOOGLE_CLIENT_SECRET"),
				RedirectURI:  viper.GetString("GOOGLE_REDIRECT_URI"),
			},
			GitHub: OAuthProvider{
				ClientID:     viper.GetString("GITHUB_CLIENT_ID"),
				ClientSecret: viper.GetString("GITHUB_CLIENT_SECRET"),
				RedirectURI:  viper.GetString("GITHUB_REDIRECT_URI"),
			},
			LinkedIn: OAuthProvider{
				ClientID:     viper.GetString("LINKEDIN_CLIENT_ID"),
				ClientSecret: viper.GetString("LINKEDIN_CLIENT_SECRET"),
				RedirectURI:  viper.GetString("LINKEDIN_REDIRECT_URI"),
			},
		},
		AI: AIConfig{
			ServiceURL:     viper.GetString("AI_SERVICE_URL"),
			OpenAIKey:      viper.GetString("OPENAI_API_KEY"),
			EmbeddingModel: viper.GetString("OPENAI_EMBEDDING_MODEL"),
			ChatModel:      viper.GetString("OPENAI_CHAT_MODEL"),
		},
		Queue: QueueConfig{
			Scrape:      viper.GetString("QUEUE_SCRAPE"),
			Parse:       viper.GetString("QUEUE_PARSE"),
			Classify:    viper.GetString("QUEUE_CLASSIFY"),
			Embed:       viper.GetString("QUEUE_EMBED"),
			Deduplicate: viper.GetString("QUEUE_DEDUPLICATE"),
			Notify:      viper.GetString("QUEUE_NOTIFY"),
			DeadLetter:  viper.GetString("QUEUE_DEAD_LETTER"),
		},
		RateLimit: RateLimitConfig{
			RequestsPerMinute: viper.GetInt("RATE_LIMIT_REQUESTS_PER_MINUTE"),
			SearchPerMinute:   viper.GetInt("RATE_LIMIT_SEARCH_PER_MINUTE"),
		},
		Notification: NotificationConfig{
			ResendAPIKey:    viper.GetString("RESEND_API_KEY"),
			EmailFrom:       viper.GetString("EMAIL_FROM"),
			EmailFromName:   viper.GetString("EMAIL_FROM_NAME"),
			VAPIDPublicKey:  viper.GetString("VAPID_PUBLIC_KEY"),
			VAPIDPrivateKey: viper.GetString("VAPID_PRIVATE_KEY"),
			VAPIDSubject:    viper.GetString("VAPID_SUBJECT"),
		},
		Security: SecurityConfig{
			BcryptCost: viper.GetInt("BCRYPT_COST"),
		},
	}
}
