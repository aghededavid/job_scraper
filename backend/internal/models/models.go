package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

// ──────────────────────────────────────────
// Enums
// ──────────────────────────────────────────

type UserRole string
type AuthProvider string
type ExperienceLevel string
type EmploymentType string
type RemoteType string
type NotificationFreq string
type FetchMethod string
type ScraperStatus string
type NotificationChannel string
type NotificationType string
type EventType string
type ApplicationStatus string

const (
	RoleUser      UserRole = "user"
	RoleAdmin     UserRole = "admin"
	RoleModerator UserRole = "moderator"

	AuthEmail    AuthProvider = "email"
	AuthGoogle   AuthProvider = "google"
	AuthGitHub   AuthProvider = "github"
	AuthLinkedIn AuthProvider = "linkedin"

	ExpInternship  ExperienceLevel = "internship"
	ExpEntry       ExperienceLevel = "entry"
	ExpJunior      ExperienceLevel = "junior"
	ExpMid         ExperienceLevel = "mid"
	ExpSenior      ExperienceLevel = "senior"
	ExpStaff       ExperienceLevel = "staff"
	ExpPrincipal   ExperienceLevel = "principal"
	ExpAny         ExperienceLevel = "any"

	EmpFullTime  EmploymentType = "full_time"
	EmpPartTime  EmploymentType = "part_time"
	EmpContract  EmploymentType = "contract"
	EmpFreelance EmploymentType = "freelance"
	EmpTemporary EmploymentType = "temporary"
	EmpAny       EmploymentType = "any"

	RemoteWorldwide  RemoteType = "worldwide"
	RemoteRestricted RemoteType = "region_restricted"
	RemoteAny        RemoteType = "any"

	NotifInstant NotificationFreq = "instant"
	NotifHourly  NotificationFreq = "hourly"
	NotifDaily   NotificationFreq = "daily"
	NotifOff     NotificationFreq = "off"

	ChannelEmail    NotificationChannel = "email"
	ChannelPush     NotificationChannel = "push"
	ChannelSMS      NotificationChannel = "sms"
	ChannelTelegram NotificationChannel = "telegram"
	ChannelDiscord  NotificationChannel = "discord"
	ChannelSlack    NotificationChannel = "slack"
	ChannelBrowser  NotificationChannel = "browser"

	NotifNewJob        NotificationType = "new_job"
	NotifSavedSearch   NotificationType = "saved_search_match"
	NotifAppUpdate     NotificationType = "application_update"
	NotifSystem        NotificationType = "system"
	NotifMarketing     NotificationType = "marketing"

	EventView   EventType = "view"
	EventClick  EventType = "click"
	EventSave   EventType = "save"
	EventApply  EventType = "apply"
	EventHide   EventType = "hide"
	EventShare  EventType = "share"
	EventSearch EventType = "search"

	AppSaved      ApplicationStatus = "saved"
	AppApplied    ApplicationStatus = "applied"
	AppPhone      ApplicationStatus = "phone_screen"
	AppInterview  ApplicationStatus = "interview"
	AppOffer      ApplicationStatus = "offer"
	AppRejected   ApplicationStatus = "rejected"
	AppWithdrawn  ApplicationStatus = "withdrawn"
)

// ──────────────────────────────────────────
// StringArray — custom type for TEXT[]
// ──────────────────────────────────────────

type StringArray []string

// ──────────────────────────────────────────
// User
// ──────────────────────────────────────────

type User struct {
	ID               uuid.UUID    `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Email            string       `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash     *string      `gorm:"column:password_hash" json:"-"`
	Name             string       `gorm:"not null;default:''" json:"name"`
	AvatarURL        *string      `json:"avatar_url,omitempty"`
	AuthProvider     AuthProvider `gorm:"type:auth_provider;not null;default:'email'" json:"auth_provider"`
	AuthProviderID   *string      `json:"-"`
	EmailVerified    bool         `gorm:"default:false" json:"email_verified"`
	Role             UserRole     `gorm:"type:user_role;not null;default:'user'" json:"role"`
	IsActive         bool         `gorm:"default:true" json:"is_active"`
	LastLoginAt      *time.Time   `json:"last_login_at,omitempty"`
	CreatedAt        time.Time    `json:"created_at"`
	UpdatedAt        time.Time    `json:"updated_at"`

	// Relations
	Preferences *UserPreferences `gorm:"foreignKey:UserID" json:"preferences,omitempty"`
}

func (User) TableName() string { return "users" }

// ──────────────────────────────────────────
// UserPreferences
// ──────────────────────────────────────────

type UserPreferences struct {
	ID                   uuid.UUID            `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID               uuid.UUID            `gorm:"type:uuid;uniqueIndex;not null" json:"user_id"`
	JobTypes             StringArray          `gorm:"type:text[]" json:"job_types"`
	ExperienceLevels     []ExperienceLevel    `gorm:"type:experience_level[]" json:"experience_levels"`
	EmploymentTypes      []EmploymentType     `gorm:"type:employment_type[]" json:"employment_types"`
	Technologies         StringArray          `gorm:"type:text[]" json:"technologies"`
	Timezones            StringArray          `gorm:"type:text[]" json:"timezones"`
	SalaryMin            *int                 `json:"salary_min,omitempty"`
	SalaryMax            *int                 `json:"salary_max,omitempty"`
	SalaryCurrency       string               `gorm:"default:'USD'" json:"salary_currency"`
	RemoteType           RemoteType           `gorm:"type:remote_type;default:'any'" json:"remote_type"`
	WorkAuthorization    *string              `json:"work_authorization,omitempty"`
	NotificationFreq     NotificationFreq     `gorm:"type:notification_freq;default:'daily'" json:"notification_freq"`
	NotificationChannels []NotificationChannel `gorm:"type:notification_channel[]" json:"notification_channels"`
	OnboardingCompleted  bool                 `gorm:"default:false" json:"onboarding_completed"`
	CreatedAt            time.Time            `json:"created_at"`
	UpdatedAt            time.Time            `json:"updated_at"`
}

func (UserPreferences) TableName() string { return "user_preferences" }

// ──────────────────────────────────────────
// Company
// ──────────────────────────────────────────

type Company struct {
	ID              uuid.UUID   `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Name            string      `gorm:"not null" json:"name"`
	Slug            string      `gorm:"uniqueIndex;not null" json:"slug"`
	Website         *string     `json:"website,omitempty"`
	LogoURL         *string     `json:"logo_url,omitempty"`
	Description     *string     `json:"description,omitempty"`
	Size            *string     `json:"size,omitempty"`
	FundingStage    *string     `json:"funding_stage,omitempty"`
	TechStack       StringArray `gorm:"type:text[]" json:"tech_stack"`
	GlassdoorURL    *string     `json:"glassdoor_url,omitempty"`
	LinkedInURL     *string     `json:"linkedin_url,omitempty"`
	TwitterURL      *string     `json:"twitter_url,omitempty"`
	IsVerified      bool        `gorm:"default:false" json:"is_verified"`
	IsBlacklisted   bool        `gorm:"default:false" json:"-"`
	BlacklistReason *string     `json:"-"`
	TotalJobs       int         `gorm:"default:0" json:"total_jobs"`
	ActiveJobs      int         `gorm:"default:0" json:"active_jobs"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`

	// Relations
	Jobs []Job `gorm:"foreignKey:CompanyID" json:"jobs,omitempty"`
}

func (Company) TableName() string { return "companies" }

// ──────────────────────────────────────────
// JobSource
// ──────────────────────────────────────────

type JobSource struct {
	ID                  uuid.UUID   `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Name                string      `gorm:"not null" json:"name"`
	Slug                string      `gorm:"uniqueIndex;not null" json:"slug"`
	AdapterClass        string      `gorm:"not null" json:"adapter_class"`
	BaseURL             string      `gorm:"not null" json:"base_url"`
	FetchMethod         FetchMethod `gorm:"type:fetch_method;default:'api'" json:"fetch_method"`
	ScheduleCron        string      `gorm:"default:'*/15 * * * *'" json:"schedule_cron"`
	IsActive            bool        `gorm:"default:true" json:"is_active"`
	RespectsRobots      bool        `gorm:"default:true" json:"respects_robots"`
	RateLimitRPM        int         `gorm:"default:30" json:"rate_limit_rpm"`
	LastRunAt           *time.Time  `json:"last_run_at,omitempty"`
	LastSuccessAt       *time.Time  `json:"last_success_at,omitempty"`
	ConsecutiveFailures int         `gorm:"default:0" json:"consecutive_failures"`
	Config              interface{} `gorm:"type:jsonb;default:'{}'" json:"config"`
	CreatedAt           time.Time   `json:"created_at"`
	UpdatedAt           time.Time   `json:"updated_at"`
}

func (JobSource) TableName() string { return "job_sources" }

// ──────────────────────────────────────────
// Job
// ──────────────────────────────────────────

type Job struct {
	ID                   uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Title                string         `gorm:"not null" json:"title"`
	NormalizedTitle      *string        `json:"normalized_title,omitempty"`
	CompanyID            *uuid.UUID     `gorm:"type:uuid" json:"company_id,omitempty"`
	Description          string         `gorm:"default:''" json:"description"`
	DescriptionSummary   *string        `json:"description_summary,omitempty"`
	Location             *string        `json:"location,omitempty"`
	IsRemote             bool           `gorm:"default:true" json:"is_remote"`
	RemoteType           RemoteType     `gorm:"type:remote_type;default:'worldwide'" json:"remote_type"`
	TimezoneRestrictions StringArray    `gorm:"type:text[]" json:"timezone_restrictions"`
	SalaryMin            *int           `json:"salary_min,omitempty"`
	SalaryMax            *int           `json:"salary_max,omitempty"`
	SalaryCurrency       string         `gorm:"default:'USD'" json:"salary_currency"`
	SalaryEstimated      bool           `gorm:"default:false" json:"salary_estimated"`
	ExperienceLevel      ExperienceLevel `gorm:"type:experience_level;default:'any'" json:"experience_level"`
	EmploymentType       EmploymentType `gorm:"type:employment_type;default:'full_time'" json:"employment_type"`
	Skills               StringArray    `gorm:"type:text[]" json:"skills"`
	Technologies         StringArray    `gorm:"type:text[]" json:"technologies"`
	Benefits             StringArray    `gorm:"type:text[]" json:"benefits"`
	VisaSponsorship      *bool          `json:"visa_sponsorship,omitempty"`
	Category             *string        `json:"category,omitempty"`
	Subcategory          *string        `json:"subcategory,omitempty"`
	Keywords             StringArray    `gorm:"type:text[]" json:"keywords"`
	ApplyURL             string         `gorm:"not null" json:"apply_url"`
	CanonicalURL         *string        `json:"canonical_url,omitempty"`
	Fingerprint          *string        `gorm:"uniqueIndex" json:"-"`
	FirstSeenAt          time.Time      `gorm:"default:now()" json:"first_seen_at"`
	PostedAt             *time.Time     `json:"posted_at,omitempty"`
	UpdatedAt            time.Time      `json:"updated_at"`
	ExpiresAt            *time.Time     `json:"expires_at,omitempty"`
	FreshnessScore       float64        `gorm:"default:1.0" json:"freshness_score"`
	IsActive             bool           `gorm:"default:true" json:"is_active"`
	IsFakeSuspected      bool           `gorm:"default:false" json:"is_fake_suspected"`
	FakeScore            *float64       `json:"fake_score,omitempty"`
	AIProcessed          bool           `gorm:"default:false" json:"ai_processed"`
	CreatedAt            time.Time      `json:"created_at"`

	// Relations
	Company    *Company      `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	Analytics  *JobAnalytics `gorm:"foreignKey:JobID" json:"analytics,omitempty"`
	SourceURLs []JobSourceURL `gorm:"foreignKey:JobID" json:"source_urls,omitempty"`
}

func (Job) TableName() string { return "jobs" }

// ──────────────────────────────────────────
// JobEmbedding
// ──────────────────────────────────────────

type JobEmbedding struct {
	ID           uuid.UUID     `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	JobID        uuid.UUID     `gorm:"type:uuid;uniqueIndex;not null" json:"job_id"`
	Embedding    pgvector.Vector `gorm:"type:vector(1536)" json:"-"`
	ModelVersion string        `gorm:"default:'text-embedding-3-small'" json:"model_version"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

func (JobEmbedding) TableName() string { return "job_embeddings" }

// ──────────────────────────────────────────
// JobSourceURL
// ──────────────────────────────────────────

type JobSourceURL struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	JobID       uuid.UUID `gorm:"type:uuid;not null" json:"job_id"`
	SourceID    uuid.UUID `gorm:"type:uuid;not null" json:"source_id"`
	SourceURL   string    `gorm:"not null" json:"source_url"`
	SourceJobID *string   `json:"source_job_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`

	Source *JobSource `gorm:"foreignKey:SourceID" json:"source,omitempty"`
}

func (JobSourceURL) TableName() string { return "job_source_urls" }

// ──────────────────────────────────────────
// ScraperRun
// ──────────────────────────────────────────

type ScraperRun struct {
	ID           uuid.UUID     `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	SourceID     uuid.UUID     `gorm:"type:uuid;not null" json:"source_id"`
	StartedAt    time.Time     `gorm:"default:now()" json:"started_at"`
	CompletedAt  *time.Time    `json:"completed_at,omitempty"`
	Status       ScraperStatus `gorm:"type:scraper_status;default:'running'" json:"status"`
	JobsFound    int           `gorm:"default:0" json:"jobs_found"`
	JobsNew      int           `gorm:"default:0" json:"jobs_new"`
	JobsUpdated  int           `gorm:"default:0" json:"jobs_updated"`
	JobsRemoved  int           `gorm:"default:0" json:"jobs_removed"`
	ErrorMessage *string       `json:"error_message,omitempty"`
	Metadata     interface{}   `gorm:"type:jsonb;default:'{}'" json:"metadata"`
	CreatedAt    time.Time     `json:"created_at"`

	Source *JobSource `gorm:"foreignKey:SourceID" json:"source,omitempty"`
}

func (ScraperRun) TableName() string { return "scraper_runs" }

// ──────────────────────────────────────────
// SavedJob
// ──────────────────────────────────────────

type SavedJob struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	JobID     uuid.UUID `gorm:"type:uuid;not null" json:"job_id"`
	Notes     *string   `json:"notes,omitempty"`
	CreatedAt time.Time `json:"created_at"`

	Job *Job `gorm:"foreignKey:JobID" json:"job,omitempty"`
}

func (SavedJob) TableName() string { return "saved_jobs" }

// ──────────────────────────────────────────
// AppliedJob
// ──────────────────────────────────────────

type AppliedJob struct {
	ID        uuid.UUID         `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID    uuid.UUID         `gorm:"type:uuid;not null" json:"user_id"`
	JobID     uuid.UUID         `gorm:"type:uuid;not null" json:"job_id"`
	Status    ApplicationStatus `gorm:"type:application_status;default:'applied'" json:"status"`
	AppliedAt time.Time         `gorm:"default:now()" json:"applied_at"`
	Notes     *string           `json:"notes,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`

	Job *Job `gorm:"foreignKey:JobID" json:"job,omitempty"`
}

func (AppliedJob) TableName() string { return "applied_jobs" }

// ──────────────────────────────────────────
// HiddenJob
// ──────────────────────────────────────────

type HiddenJob struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	JobID     uuid.UUID `gorm:"type:uuid;not null" json:"job_id"`
	Reason    *string   `json:"reason,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

func (HiddenJob) TableName() string { return "hidden_jobs" }

// ──────────────────────────────────────────
// SavedSearch
// ──────────────────────────────────────────

type SavedSearch struct {
	ID                   uuid.UUID            `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID               uuid.UUID            `gorm:"type:uuid;not null" json:"user_id"`
	Name                 string               `gorm:"not null" json:"name"`
	Query                *string              `json:"query,omitempty"`
	Filters              interface{}          `gorm:"type:jsonb;default:'{}'" json:"filters"`
	Notify               bool                 `gorm:"default:true" json:"notify"`
	NotificationChannels []NotificationChannel `gorm:"type:notification_channel[]" json:"notification_channels"`
	LastNotifiedAt       *time.Time           `json:"last_notified_at,omitempty"`
	CreatedAt            time.Time            `json:"created_at"`
	UpdatedAt            time.Time            `json:"updated_at"`
}

func (SavedSearch) TableName() string { return "saved_searches" }

// ──────────────────────────────────────────
// Notification
// ──────────────────────────────────────────

type Notification struct {
	ID        uuid.UUID           `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID    uuid.UUID           `gorm:"type:uuid;not null" json:"user_id"`
	Type      NotificationType    `gorm:"type:notification_type;not null" json:"type"`
	Title     string              `gorm:"not null" json:"title"`
	Body      string              `gorm:"not null" json:"body"`
	Data      interface{}         `gorm:"type:jsonb;default:'{}'" json:"data"`
	Channel   NotificationChannel `gorm:"type:notification_channel;default:'browser'" json:"channel"`
	IsRead    bool                `gorm:"default:false" json:"is_read"`
	IsSent    bool                `gorm:"default:false" json:"is_sent"`
	SentAt    *time.Time          `json:"sent_at,omitempty"`
	CreatedAt time.Time           `json:"created_at"`
}

func (Notification) TableName() string { return "notifications" }

// ──────────────────────────────────────────
// UserEvent
// ──────────────────────────────────────────

type UserEvent struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID    *uuid.UUID `gorm:"type:uuid" json:"user_id,omitempty"`
	SessionID *string   `json:"session_id,omitempty"`
	JobID     *uuid.UUID `gorm:"type:uuid" json:"job_id,omitempty"`
	EventType EventType `gorm:"type:event_type;not null" json:"event_type"`
	Metadata  interface{} `gorm:"type:jsonb;default:'{}'" json:"metadata"`
	CreatedAt time.Time `json:"created_at"`
}

func (UserEvent) TableName() string { return "user_events" }

// ──────────────────────────────────────────
// JobAnalytics
// ──────────────────────────────────────────

type JobAnalytics struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	JobID          uuid.UUID  `gorm:"type:uuid;uniqueIndex;not null" json:"job_id"`
	Views          int        `gorm:"default:0" json:"views"`
	Clicks         int        `gorm:"default:0" json:"clicks"`
	Saves          int        `gorm:"default:0" json:"saves"`
	Applications   int        `gorm:"default:0" json:"applications"`
	LastActivityAt *time.Time `json:"last_activity_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func (JobAnalytics) TableName() string { return "job_analytics" }

// ──────────────────────────────────────────
// Category
// ──────────────────────────────────────────

type Category struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Name      string     `gorm:"uniqueIndex;not null" json:"name"`
	Slug      string     `gorm:"uniqueIndex;not null" json:"slug"`
	ParentID  *uuid.UUID `gorm:"type:uuid" json:"parent_id,omitempty"`
	Icon      *string    `json:"icon,omitempty"`
	Color     *string    `json:"color,omitempty"`
	SortOrder int        `gorm:"default:0" json:"sort_order"`
	CreatedAt time.Time  `json:"created_at"`
}

func (Category) TableName() string { return "categories" }

// ──────────────────────────────────────────
// AuditLog
// ──────────────────────────────────────────

type AuditLog struct {
	ID           uuid.UUID   `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID       *uuid.UUID  `gorm:"type:uuid" json:"user_id,omitempty"`
	Action       string      `gorm:"not null" json:"action"`
	ResourceType string      `gorm:"not null" json:"resource_type"`
	ResourceID   *uuid.UUID  `gorm:"type:uuid" json:"resource_id,omitempty"`
	OldValue     interface{} `gorm:"type:jsonb" json:"old_value,omitempty"`
	NewValue     interface{} `gorm:"type:jsonb" json:"new_value,omitempty"`
	IPAddress    *string     `json:"ip_address,omitempty"`
	UserAgent    *string     `json:"user_agent,omitempty"`
	CreatedAt    time.Time   `json:"created_at"`
}

func (AuditLog) TableName() string { return "audit_log" }

// ──────────────────────────────────────────
// API Response Types
// ──────────────────────────────────────────

type PaginatedResponse[T any] struct {
	Data       []T   `json:"data"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	TotalPages int   `json:"total_pages"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func Success(data interface{}) APIResponse {
	return APIResponse{Success: true, Data: data}
}

func SuccessMsg(msg string) APIResponse {
	return APIResponse{Success: true, Message: msg}
}

func ErrorResp(err string) APIResponse {
	return APIResponse{Success: false, Error: err}
}

// ──────────────────────────────────────────
// Search / Filter DTOs
// ──────────────────────────────────────────

type JobFilters struct {
	Query          string          `form:"q" json:"q"`
	Category       string          `form:"category" json:"category"`
	ExperienceLevel []ExperienceLevel `form:"experience" json:"experience"`
	EmploymentType  []EmploymentType  `form:"employment_type" json:"employment_type"`
	RemoteType     string          `form:"remote_type" json:"remote_type"`
	SalaryMin      *int            `form:"salary_min" json:"salary_min"`
	SalaryMax      *int            `form:"salary_max" json:"salary_max"`
	Technologies   []string        `form:"tech" json:"tech"`
	Company        string          `form:"company" json:"company"`
	PostedWithin   string          `form:"posted_within" json:"posted_within"` // "5m", "1h", "24h", "7d"
	VisaSponsorship *bool          `form:"visa" json:"visa"`
	Page           int             `form:"page,default=1" json:"page"`
	PerPage        int             `form:"per_page,default=20" json:"per_page"`
	SortBy         string          `form:"sort_by,default=posted_at" json:"sort_by"`
	SortDir        string          `form:"sort_dir,default=desc" json:"sort_dir"`
}

// BeforeCreate hook for gorm
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
