package repository

import (
	"errors"

	"github.com/google/uuid"
	"github.com/remotescope/backend/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CompanyRepository struct {
	db *gorm.DB
}

func NewCompanyRepository(db *gorm.DB) *CompanyRepository {
	return &CompanyRepository{db: db}
}

func (r *CompanyRepository) Create(company *models.Company) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "slug"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "website", "logo_url", "description", "updated_at"}),
	}).Create(company).Error
}

func (r *CompanyRepository) FindByID(id uuid.UUID) (*models.Company, error) {
	var company models.Company
	err := r.db.First(&company, "id = ? AND is_blacklisted = false", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &company, nil
}

func (r *CompanyRepository) FindBySlug(slug string) (*models.Company, error) {
	var company models.Company
	err := r.db.Where("slug = ? AND is_blacklisted = false", slug).First(&company).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &company, nil
}

func (r *CompanyRepository) FindByName(name string) (*models.Company, error) {
	var company models.Company
	err := r.db.Where("name ILIKE ?", name).First(&company).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &company, nil
}

func (r *CompanyRepository) List(page, perPage int, query string) ([]models.Company, int64, error) {
	var companies []models.Company
	var total int64

	q := r.db.Model(&models.Company{}).Where("is_blacklisted = false")
	if query != "" {
		q = q.Where("name ILIKE ?", "%"+query+"%")
	}
	q.Count(&total)

	offset := (page - 1) * perPage
	err := q.Order("active_jobs DESC, is_verified DESC, name ASC").
		Offset(offset).Limit(perPage).
		Find(&companies).Error
	return companies, total, err
}

func (r *CompanyRepository) Update(company *models.Company) error {
	return r.db.Save(company).Error
}

func (r *CompanyRepository) UpdateJobCounts(companyID uuid.UUID) error {
	return r.db.Exec(`
		UPDATE companies SET 
			total_jobs = (SELECT COUNT(*) FROM jobs WHERE company_id = ?),
			active_jobs = (SELECT COUNT(*) FROM jobs WHERE company_id = ? AND is_active = true)
		WHERE id = ?
	`, companyID, companyID, companyID).Error
}

func (r *CompanyRepository) Blacklist(id uuid.UUID, reason string) error {
	return r.db.Model(&models.Company{}).Where("id = ?", id).Updates(map[string]interface{}{
		"is_blacklisted":   true,
		"blacklist_reason": reason,
	}).Error
}

func (r *CompanyRepository) GetTopHiring(limit int) ([]models.Company, error) {
	var companies []models.Company
	err := r.db.Where("is_blacklisted = false AND active_jobs > 0").
		Order("active_jobs DESC").
		Limit(limit).
		Find(&companies).Error
	return companies, err
}
