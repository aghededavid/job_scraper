package repository

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/remotescope/backend/internal/models"
	"gorm.io/gorm"
)

type JobRepository struct {
	db *gorm.DB
}

func NewJobRepository(db *gorm.DB) *JobRepository {
	return &JobRepository{db: db}
}

// ──────────────────────────────────────────
// Create
// ──────────────────────────────────────────

func (r *JobRepository) Create(job *models.Job) error {
	return r.db.Create(job).Error
}

func (r *JobRepository) Upsert(job *models.Job) error {
	return r.db.Where("fingerprint = ?", job.Fingerprint).
		Assign(job).
		FirstOrCreate(job).Error
}

// ──────────────────────────────────────────
// Read
// ──────────────────────────────────────────

func (r *JobRepository) FindByID(id uuid.UUID) (*models.Job, error) {
	var job models.Job
	err := r.db.
		Preload("Company").
		Preload("Analytics").
		Preload("SourceURLs.Source").
		First(&job, "id = ? AND is_active = true", id).Error
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *JobRepository) FindByFingerprint(fp string) (*models.Job, bool) {
	var job models.Job
	err := r.db.Where("fingerprint = ?", fp).First(&job).Error
	return &job, err == nil
}

func (r *JobRepository) List(filters *models.JobFilters, userID *uuid.UUID) ([]models.Job, int64, error) {
	q := r.db.Model(&models.Job{}).
		Preload("Company").
		Preload("Analytics").
		Where("jobs.is_active = true AND jobs.is_fake_suspected = false")

	// Exclude hidden jobs for authenticated users
	if userID != nil {
		q = q.Where("jobs.id NOT IN (SELECT job_id FROM hidden_jobs WHERE user_id = ?)", *userID)
	}

	// Full text search
	if filters.Query != "" {
		q = q.Where(
			"to_tsvector('english', coalesce(title,'') || ' ' || coalesce(description,'') || ' ' || array_to_string(skills,' ') || ' ' || array_to_string(technologies,' ')) @@ plainto_tsquery('english', ?)",
			filters.Query,
		)
	}

	// Category filter
	if filters.Category != "" {
		q = q.Where("category = ?", filters.Category)
	}

	// Experience level
	if len(filters.ExperienceLevel) > 0 {
		q = q.Where("experience_level = ANY(?)", filters.ExperienceLevel)
	}

	// Employment type
	if len(filters.EmploymentType) > 0 {
		q = q.Where("employment_type = ANY(?)", filters.EmploymentType)
	}

	// Remote type
	if filters.RemoteType != "" {
		q = q.Where("remote_type = ?", filters.RemoteType)
	}

	// Salary range
	if filters.SalaryMin != nil {
		q = q.Where("salary_max >= ?", *filters.SalaryMin)
	}
	if filters.SalaryMax != nil {
		q = q.Where("salary_min <= ?", *filters.SalaryMax)
	}

	// Technologies
	if len(filters.Technologies) > 0 {
		q = q.Where("technologies && ?", filters.Technologies)
	}

	// Company search
	if filters.Company != "" {
		q = q.Joins("JOIN companies ON companies.id = jobs.company_id").
			Where("companies.name ILIKE ?", "%"+filters.Company+"%")
	}

	// Visa sponsorship
	if filters.VisaSponsorship != nil {
		q = q.Where("visa_sponsorship = ?", *filters.VisaSponsorship)
	}

	// Posted within
	if filters.PostedWithin != "" {
		cutoff := parseDuration(filters.PostedWithin)
		if cutoff != nil {
			q = q.Where("first_seen_at >= ?", cutoff)
		}
	}

	// Count total
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Sort
	sortCol := "jobs.first_seen_at"
	switch filters.SortBy {
	case "posted_at":
		sortCol = "jobs.posted_at"
	case "salary":
		sortCol = "jobs.salary_max"
	case "freshness":
		sortCol = "jobs.freshness_score"
	case "relevance":
		sortCol = "jobs.freshness_score" // default; vector sort applied elsewhere
	}
	dir := "DESC"
	if filters.SortDir == "asc" {
		dir = "ASC"
	}
	q = q.Order(fmt.Sprintf("%s %s NULLS LAST", sortCol, dir))

	// Paginate
	if filters.PerPage <= 0 {
		filters.PerPage = 20
	}
	if filters.PerPage > 100 {
		filters.PerPage = 100
	}
	if filters.Page <= 0 {
		filters.Page = 1
	}
	offset := (filters.Page - 1) * filters.PerPage
	q = q.Offset(offset).Limit(filters.PerPage)

	var jobs []models.Job
	if err := q.Find(&jobs).Error; err != nil {
		return nil, 0, err
	}
	return jobs, total, nil
}

func (r *JobRepository) GetLatest(limit int) ([]models.Job, error) {
	var jobs []models.Job
	err := r.db.
		Preload("Company").
		Where("is_active = true AND is_fake_suspected = false").
		Order("first_seen_at DESC").
		Limit(limit).
		Find(&jobs).Error
	return jobs, err
}

func (r *JobRepository) GetTrending(limit int) ([]models.Job, error) {
	var jobs []models.Job
	err := r.db.
		Preload("Company").
		Joins("LEFT JOIN job_analytics ON job_analytics.job_id = jobs.id").
		Where("jobs.is_active = true AND jobs.is_fake_suspected = false AND jobs.first_seen_at >= ?", time.Now().AddDate(0, 0, -7)).
		Order("(COALESCE(job_analytics.views,0) + COALESCE(job_analytics.saves,0)*3 + COALESCE(job_analytics.applications,0)*5) DESC").
		Limit(limit).
		Find(&jobs).Error
	return jobs, err
}

func (r *JobRepository) GetRecommended(userID uuid.UUID, limit int) ([]models.Job, error) {
	// Simple preference-based recommendation
	var prefs models.UserPreferences
	if err := r.db.Where("user_id = ?", userID).First(&prefs).Error; err != nil {
		return r.GetLatest(limit)
	}

	q := r.db.Model(&models.Job{}).
		Preload("Company").
		Where("jobs.is_active = true AND jobs.is_fake_suspected = false").
		Where("jobs.id NOT IN (SELECT job_id FROM hidden_jobs WHERE user_id = ?)", userID).
		Where("jobs.id NOT IN (SELECT job_id FROM applied_jobs WHERE user_id = ?)", userID)

	if len(prefs.Technologies) > 0 {
		q = q.Where("technologies && ?", prefs.Technologies)
	}
	if len(prefs.ExperienceLevels) > 0 {
		q = q.Where("experience_level = ANY(?)", prefs.ExperienceLevels)
	}

	var jobs []models.Job
	err := q.Order("freshness_score DESC, first_seen_at DESC").Limit(limit).Find(&jobs).Error
	return jobs, err
}

// ──────────────────────────────────────────
// Vector search
// ──────────────────────────────────────────

func (r *JobRepository) VectorSearch(embedding []float32, limit int, filters *models.JobFilters) ([]models.Job, error) {
	var jobIDs []uuid.UUID

	// Convert to pgvector format
	embStr := floatsToVectorStr(embedding)
	query := fmt.Sprintf(`
		SELECT je.job_id
		FROM job_embeddings je
		JOIN jobs j ON j.id = je.job_id
		WHERE j.is_active = true AND j.is_fake_suspected = false
		ORDER BY je.embedding <=> '%s'
		LIMIT %d
	`, embStr, limit)

	rows, err := r.db.Raw(query).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id uuid.UUID
		rows.Scan(&id)
		jobIDs = append(jobIDs, id)
	}

	if len(jobIDs) == 0 {
		return nil, nil
	}

	var jobs []models.Job
	err = r.db.
		Preload("Company").
		Preload("Analytics").
		Where("id IN ?", jobIDs).
		Find(&jobs).Error
	return jobs, err
}

// ──────────────────────────────────────────
// Update
// ──────────────────────────────────────────

func (r *JobRepository) Update(job *models.Job) error {
	return r.db.Save(job).Error
}

func (r *JobRepository) UpdateFreshnessScores() error {
	return r.db.Exec(`
		UPDATE jobs 
		SET freshness_score = calculate_freshness_score(posted_at),
		    is_active = CASE 
		        WHEN posted_at < NOW() - INTERVAL '60 days' THEN false 
		        ELSE is_active 
		    END
		WHERE is_active = true
	`).Error
}

func (r *JobRepository) MarkExpired() error {
	return r.db.Exec(`
		UPDATE jobs SET is_active = false 
		WHERE expires_at IS NOT NULL AND expires_at < NOW() AND is_active = true
	`).Error
}

func (r *JobRepository) IncrementView(jobID uuid.UUID) error {
	return r.db.Exec(`
		INSERT INTO job_analytics (id, job_id, views, last_activity_at)
		VALUES (uuid_generate_v4(), ?, 1, NOW())
		ON CONFLICT (job_id) DO UPDATE
		SET views = job_analytics.views + 1, last_activity_at = NOW()
	`, jobID).Error
}

func (r *JobRepository) Count() (int64, error) {
	var count int64
	return count, r.db.Model(&models.Job{}).Where("is_active = true").Count(&count).Error
}

// ──────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────

func parseDuration(d string) *time.Time {
	var dur time.Duration
	switch d {
	case "5m":
		dur = 5 * time.Minute
	case "15m":
		dur = 15 * time.Minute
	case "1h":
		dur = time.Hour
	case "24h":
		dur = 24 * time.Hour
	case "7d":
		dur = 7 * 24 * time.Hour
	default:
		return nil
	}
	t := time.Now().Add(-dur)
	return &t
}

func floatsToVectorStr(v []float32) string {
	s := "["
	for i, f := range v {
		if i > 0 {
			s += ","
		}
		s += fmt.Sprintf("%f", f)
	}
	return s + "]"
}
