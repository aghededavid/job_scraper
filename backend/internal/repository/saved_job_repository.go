package repository

import (
	"github.com/google/uuid"
	"github.com/remotescope/backend/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SavedJobRepository struct {
	db *gorm.DB
}

func NewSavedJobRepository(db *gorm.DB) *SavedJobRepository {
	return &SavedJobRepository{db: db}
}

func (r *SavedJobRepository) Save(userID, jobID uuid.UUID) error {
	saved := &models.SavedJob{
		ID:     uuid.New(),
		UserID: userID,
		JobID:  jobID,
	}
	return r.db.Clauses(clause.OnConflict{DoNothing: true}).Create(saved).Error
}

func (r *SavedJobRepository) Unsave(userID, jobID uuid.UUID) error {
	return r.db.Where("user_id = ? AND job_id = ?", userID, jobID).Delete(&models.SavedJob{}).Error
}

func (r *SavedJobRepository) ListSaved(userID uuid.UUID, page, perPage int) ([]models.SavedJob, int64, error) {
	var saved []models.SavedJob
	var total int64
	r.db.Model(&models.SavedJob{}).Where("user_id = ?", userID).Count(&total)
	offset := (page - 1) * perPage
	err := r.db.
		Preload("Job.Company").
		Preload("Job.Analytics").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).Limit(perPage).
		Find(&saved).Error
	return saved, total, err
}

func (r *SavedJobRepository) IsSaved(userID, jobID uuid.UUID) bool {
	var count int64
	r.db.Model(&models.SavedJob{}).Where("user_id = ? AND job_id = ?", userID, jobID).Count(&count)
	return count > 0
}

func (r *SavedJobRepository) Hide(userID, jobID uuid.UUID, reason string) error {
	hidden := &models.HiddenJob{
		ID:     uuid.New(),
		UserID: userID,
		JobID:  jobID,
	}
	if reason != "" {
		hidden.Reason = &reason
	}
	return r.db.Clauses(clause.OnConflict{DoNothing: true}).Create(hidden).Error
}

func (r *SavedJobRepository) MarkApplied(userID, jobID uuid.UUID) error {
	applied := &models.AppliedJob{
		ID:     uuid.New(),
		UserID: userID,
		JobID:  jobID,
		Status: models.AppApplied,
	}
	return r.db.Clauses(clause.OnConflict{DoNothing: true}).Create(applied).Error
}

func (r *SavedJobRepository) ListApplied(userID uuid.UUID, page, perPage int) ([]models.AppliedJob, int64, error) {
	var applied []models.AppliedJob
	var total int64
	r.db.Model(&models.AppliedJob{}).Where("user_id = ?", userID).Count(&total)
	offset := (page - 1) * perPage
	err := r.db.
		Preload("Job.Company").
		Where("user_id = ?", userID).
		Order("applied_at DESC").
		Offset(offset).Limit(perPage).
		Find(&applied).Error
	return applied, total, err
}

func (r *SavedJobRepository) UpdateApplicationStatus(userID, jobID uuid.UUID, status models.ApplicationStatus) error {
	return r.db.Model(&models.AppliedJob{}).
		Where("user_id = ? AND job_id = ?", userID, jobID).
		Update("status", status).Error
}
