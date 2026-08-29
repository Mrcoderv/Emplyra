package repositories

import (
	"gorm.io/gorm"

	"github.com/emplyra/backend/internal/models"
	"github.com/emplyra/backend/internal/utils"
)

type TenantRepository struct {
	db *gorm.DB
}

func NewTenantRepository(db *gorm.DB) *TenantRepository {
	return &TenantRepository{db: db}
}

func (r *TenantRepository) Create(t *models.Tenant) error { return r.db.Create(t).Error }

func (r *TenantRepository) Update(id string, fields map[string]interface{}) error {
	return r.db.Model(&models.Tenant{}).Where("id = ?", id).Updates(fields).Error
}

func (r *TenantRepository) Delete(id string) error {
	return r.db.Delete(&models.Tenant{}, "id = ?", id).Error
}

func (r *TenantRepository) FindByID(id string) (*models.Tenant, error) {
	var t models.Tenant
	err := r.db.First(&t, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TenantRepository) FindBySlug(slug string) (*models.Tenant, error) {
	var t models.Tenant
	err := r.db.Where("slug = ?", slug).First(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TenantRepository) List(p utils.Pagination, search, status string) ([]models.Tenant, int64, error) {
	var items []models.Tenant
	var total int64
	q := r.db.Model(&models.Tenant{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if search != "" {
		like := "%" + search + "%"
		q = q.Where("LOWER(name) LIKE LOWER(?) OR LOWER(slug) LIKE LOWER(?)", like, like)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Order("created_at DESC").Offset(p.Offset).Limit(p.Limit).Find(&items).Error
	return items, total, err
}

func (r *TenantRepository) CountByStatus() (map[models.TenantStatus]int64, error) {
	rows := []struct {
		Status models.TenantStatus
		Count  int64
	}{}
	err := r.db.Model(&models.Tenant{}).Select("status, COUNT(*) as count").Group("status").Scan(&rows).Error
	out := map[models.TenantStatus]int64{}
	if err != nil {
		return out, err
	}
	for _, rw := range rows {
		out[rw.Status] = rw.Count
	}
	return out, nil
}

func (r *TenantRepository) Count() (int64, error) {
	var n int64
	err := r.db.Model(&models.Tenant{}).Count(&n).Error
	return n, err
}

// Usage returns marketer-facing counters for a tenant.
func (r *TenantRepository) Usage(tenantID string) (models.TenantUsage, error) {
	var u models.TenantUsage
	if err := r.db.Model(&models.User{}).Where("tenant_id = ?", tenantID).Count(&u.Users).Error; err != nil {
		return u, err
	}
	if err := r.db.Model(&models.Employee{}).Where("tenant_id = ?", tenantID).Count(&u.Employees).Error; err != nil {
		return u, err
	}
	if err := r.db.Model(&models.Department{}).Where("tenant_id = ?", tenantID).Count(&u.Departments).Error; err != nil {
		return u, err
	}
	if err := r.db.Model(&models.JobPost{}).Where("tenant_id = ?", tenantID).Count(&u.Jobs).Error; err != nil {
		return u, err
	}
	if err := r.db.Model(&models.Candidate{}).Where("tenant_id = ?", tenantID).Count(&u.Candidates).Error; err != nil {
		return u, err
	}
	if err := r.db.Model(&models.Document{}).Where("tenant_id = ?", tenantID).Count(&u.Documents).Error; err != nil {
		return u, err
	}
	return u, nil
}
