package repositories

import (
	"gorm.io/gorm"

	"github.com/emplyra/backend/internal/models"
	"github.com/emplyra/backend/internal/utils"
)

type GoalRepository struct {
	db *gorm.DB
}

func NewGoalRepository(db *gorm.DB) *GoalRepository { return &GoalRepository{db: db} }

func (r *GoalRepository) Create(m *models.Goal) error { return r.db.Create(m).Error }
func (r *GoalRepository) Update(id string, f map[string]interface{}) error {
	return r.db.Model(&models.Goal{}).Where("id = ?", id).Updates(f).Error
}
func (r *GoalRepository) Delete(id string) error {
	return r.db.Delete(&models.Goal{}, "id = ?", id).Error
}
func (r *GoalRepository) FindByID(id string) (*models.Goal, error) {
	var m models.Goal
	err := r.db.Preload("Employee").First(&m, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}
func (r *GoalRepository) List(p utils.Pagination, employeeIDs []string, status string) ([]models.Goal, int64, error) {
	var items []models.Goal
	var total int64
	q := r.db.Model(&models.Goal{})
	if len(employeeIDs) > 0 {
		q = q.Where("employee_id IN ?", employeeIDs)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Preload("Employee").Order("created_at DESC").Offset(p.Offset).Limit(p.Limit).Find(&items).Error
	return items, total, err
}

type KpiRepository struct {
	db *gorm.DB
}

func NewKpiRepository(db *gorm.DB) *KpiRepository { return &KpiRepository{db: db} }

func (r *KpiRepository) Create(m *models.KPI) error { return r.db.Create(m).Error }
func (r *KpiRepository) Update(id string, f map[string]interface{}) error {
	return r.db.Model(&models.KPI{}).Where("id = ?", id).Updates(f).Error
}
func (r *KpiRepository) Delete(id string) error {
	return r.db.Delete(&models.KPI{}, "id = ?", id).Error
}
func (r *KpiRepository) FindByID(id string) (*models.KPI, error) {
	var m models.KPI
	err := r.db.Preload("Employee").First(&m, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}
func (r *KpiRepository) List(p utils.Pagination, employeeIDs []string, period string) ([]models.KPI, int64, error) {
	var items []models.KPI
	var total int64
	q := r.db.Model(&models.KPI{})
	if len(employeeIDs) > 0 {
		q = q.Where("employee_id IN ?", employeeIDs)
	}
	if period != "" {
		q = q.Where("period = ?", period)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Preload("Employee").Order("created_at DESC").Offset(p.Offset).Limit(p.Limit).Find(&items).Error
	return items, total, err
}

type ReviewRepository struct {
	db *gorm.DB
}

func NewReviewRepository(db *gorm.DB) *ReviewRepository { return &ReviewRepository{db: db} }

func (r *ReviewRepository) Create(m *models.PerformanceReview) error { return r.db.Create(m).Error }
func (r *ReviewRepository) Update(id string, f map[string]interface{}) error {
	return r.db.Model(&models.PerformanceReview{}).Where("id = ?", id).Updates(f).Error
}
func (r *ReviewRepository) FindByID(id string) (*models.PerformanceReview, error) {
	var m models.PerformanceReview
	err := r.db.Preload("Employee").Preload("Reviewer").First(&m, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}
func (r *ReviewRepository) List(p utils.Pagination, employeeIDs []string, reviewerID, status string) ([]models.PerformanceReview, int64, error) {
	var items []models.PerformanceReview
	var total int64
	q := r.db.Model(&models.PerformanceReview{})
	if len(employeeIDs) > 0 {
		q = q.Where("employee_id IN ?", employeeIDs)
	}
	if reviewerID != "" {
		q = q.Where("reviewer_id = ?", reviewerID)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Preload("Employee").Preload("Reviewer").Order("created_at DESC").Offset(p.Offset).Limit(p.Limit).Find(&items).Error
	return items, total, err
}
