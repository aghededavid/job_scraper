package repository

import (
	"github.com/google/uuid"
	"github.com/remotescope/backend/internal/models"
	"gorm.io/gorm"
)

type NotificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create(n *models.Notification) error {
	return r.db.Create(n).Error
}

func (r *NotificationRepository) List(userID uuid.UUID, page, perPage int) ([]models.Notification, int64, error) {
	var notifs []models.Notification
	var total int64
	r.db.Model(&models.Notification{}).Where("user_id = ?", userID).Count(&total)
	offset := (page - 1) * perPage
	err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).Limit(perPage).
		Find(&notifs).Error
	return notifs, total, err
}

func (r *NotificationRepository) UnreadCount(userID uuid.UUID) (int64, error) {
	var count int64
	return count, r.db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = false", userID).
		Count(&count).Error
}

func (r *NotificationRepository) MarkRead(id, userID uuid.UUID) error {
	return r.db.Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("is_read", true).Error
}

func (r *NotificationRepository) MarkAllRead(userID uuid.UUID) error {
	return r.db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = false", userID).
		Update("is_read", true).Error
}

func (r *NotificationRepository) GetPending() ([]models.Notification, error) {
	var notifs []models.Notification
	err := r.db.Where("is_sent = false").
		Order("created_at ASC").
		Limit(100).
		Find(&notifs).Error
	return notifs, err
}

func (r *NotificationRepository) MarkSent(id uuid.UUID) error {
	return r.db.Model(&models.Notification{}).Where("id = ?", id).
		Updates(map[string]interface{}{"is_sent": true, "sent_at": "NOW()"}).Error
}

func (r *NotificationRepository) ListSavedSearches(userID uuid.UUID) ([]models.SavedSearch, error) {
	var searches []models.SavedSearch
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&searches).Error
	return searches, err
}

func (r *NotificationRepository) CreateSavedSearch(s *models.SavedSearch) error {
	return r.db.Create(s).Error
}

func (r *NotificationRepository) DeleteSavedSearch(id, userID uuid.UUID) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.SavedSearch{}).Error
}

func (r *NotificationRepository) GetSavedSearchesForAlert() ([]models.SavedSearch, error) {
	var searches []models.SavedSearch
	err := r.db.Where("notify = true").Find(&searches).Error
	return searches, err
}
