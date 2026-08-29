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
func (r *GoalRepository) Update(tenantID, id string, f map[string]interface{}) error {
	return r.db.Model(&models.Goal{}).Scopes(TenantScope(tenantID)).Where("id = ?", id).Updates(f).Error
}
func (r *GoalRepository) Delete(tenantID, id string) error {
	return r.db.Scopes(TenantScope(tenantID)).Delete(&models.Goal{}, "id = ?", id).Error
}
func (r *GoalRepository) FindByID(tenantID, id string) (*models.Goal, error) {
	var m models.Goal
	err := r.db.Scopes(TenantScope(tenantID)).Preload("Employee").First(&m, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}
func (r *GoalRepository) List(tenantID string, p utils.Pagination, employeeIDs []string, status string) ([]models.Goal, int64, error) {
	var items []models.Goal
	var total int64
	q := r.db.Scopes(TenantScope(tenantID)).Model(&models.Goal{})
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
func (r *KpiRepository) Update(tenantID, id string, f map[string]interface{}) error {
	return r.db.Model(&models.KPI{}).Scopes(TenantScope(tenantID)).Where("id = ?", id).Updates(f).Error
}
func (r *KpiRepository) Delete(tenantID, id string) error {
	return r.db.Scopes(TenantScope(tenantID)).Delete(&models.KPI{}, "id = ?", id).Error
}
func (r *KpiRepository) FindByID(tenantID, id string) (*models.KPI, error) {
	var m models.KPI
	err := r.db.Scopes(TenantScope(tenantID)).Preload("Employee").First(&m, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}
func (r *KpiRepository) List(tenantID string, p utils.Pagination, employeeIDs []string, period string) ([]models.KPI, int64, error) {
	var items []models.KPI
	var total int64
	q := r.db.Scopes(TenantScope(tenantID)).Model(&models.KPI{})
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
func (r *ReviewRepository) Update(tenantID, id string, f map[string]interface{}) error {
	return r.db.Model(&models.PerformanceReview{}).Scopes(TenantScope(tenantID)).Where("id = ?", id).Updates(f).Error
}
func (r *ReviewRepository) FindByID(tenantID, id string) (*models.PerformanceReview, error) {
	var m models.PerformanceReview
	err := r.db.Scopes(TenantScope(tenantID)).Preload("Employee").Preload("Reviewer").First(&m, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}
func (r *ReviewRepository) List(tenantID string, p utils.Pagination, employeeIDs []string, reviewerID, status string) ([]models.PerformanceReview, int64, error) {
	var items []models.PerformanceReview
	var total int64
	q := r.db.Scopes(TenantScope(tenantID)).Model(&models.PerformanceReview{})
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
