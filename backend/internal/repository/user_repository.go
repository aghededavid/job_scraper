package repository

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/remotescope/backend/internal/models"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) FindByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	if err := r.db.Preload("Preferences").First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.db.Preload("Preferences").Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByOAuth(provider, providerID string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("auth_provider = ? AND auth_provider_id = ?", provider, providerID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *UserRepository) UpdateLastLogin(id uuid.UUID) error {
	now := time.Now()
	return r.db.Model(&models.User{}).Where("id = ?", id).Update("last_login_at", now).Error
}

func (r *UserRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.User{}, "id = ?", id).Error
}

func (r *UserRepository) CreatePreferences(prefs *models.UserPreferences) error {
	return r.db.FirstOrCreate(prefs, "user_id = ?", prefs.UserID).Error
}

func (r *UserRepository) GetPreferences(userID uuid.UUID) (*models.UserPreferences, error) {
	var prefs models.UserPreferences
	if err := r.db.Where("user_id = ?", userID).First(&prefs).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &prefs, nil
}

func (r *UserRepository) UpdatePreferences(prefs *models.UserPreferences) error {
	return r.db.Save(prefs).Error
}

func (r *UserRepository) Count() (int64, error) {
	var count int64
	return count, r.db.Model(&models.User{}).Count(&count).Error
}

func (r *UserRepository) List(page, perPage int) ([]models.User, int64, error) {
	var users []models.User
	var total int64
	r.db.Model(&models.User{}).Count(&total)
	offset := (page - 1) * perPage
	err := r.db.Offset(offset).Limit(perPage).Order("created_at DESC").Find(&users).Error
	return users, total, err
}
