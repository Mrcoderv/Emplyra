package repositories

import (
	"strings"

	"gorm.io/gorm"

	"github.com/emplyra/backend/internal/models"
	"github.com/emplyra/backend/internal/utils"
)

type EmployeeRepository struct {
	db *gorm.DB
}

func NewEmployeeRepository(db *gorm.DB) *EmployeeRepository {
	return &EmployeeRepository{db: db}
}

const employeePreloads = "Department,Designation,Manager,User"

func (r *EmployeeRepository) Create(e *models.Employee) error {
	return r.db.Create(e).Error
}

func (r *EmployeeRepository) Update(tenantID, id string, fields map[string]interface{}) error {
	return r.db.Model(&models.Employee{}).Scopes(TenantScope(tenantID)).Where("id = ?", id).Updates(fields).Error
}

func (r *EmployeeRepository) Delete(tenantID, id string) error {
	return r.db.Delete(&models.Employee{}, "id = ? AND tenant_id = ?", id, tenantID).Error
}

func (r *EmployeeRepository) FindByID(tenantID, id string) (*models.Employee, error) {
	var e models.Employee
	err := r.db.Scopes(TenantScope(tenantID)).Preload(employeePreloads).First(&e, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *EmployeeRepository) FindByCode(tenantID, code string) (*models.Employee, error) {
	var e models.Employee
	err := r.db.Scopes(TenantScope(tenantID)).Where("employee_code = ?", code).First(&e).Error
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *EmployeeRepository) FindByEmail(tenantID, email string) (*models.Employee, error) {
	var e models.Employee
	err := r.db.Scopes(TenantScope(tenantID)).Where("LOWER(email) = LOWER(?)", email).First(&e).Error
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *EmployeeRepository) FindByUserID(userID string) (*models.Employee, error) {
	var e models.Employee
	err := r.db.Where("user_id = ?", userID).First(&e).Error
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *EmployeeRepository) List(tenantID string, p utils.Pagination, filter func(*gorm.DB) *gorm.DB, sortBy, sortOrder string) ([]models.Employee, int64, error) {
	var items []models.Employee
	var total int64
	q := r.db.Scopes(TenantScope(tenantID)).Model(&models.Employee{})
	if filter != nil {
		q = filter(q)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	order := sortClause(sortBy, sortOrder)
	q = q.Order(order).Offset(p.Offset).Limit(p.Limit)
	err := q.Preload(employeePreloads).Find(&items).Error
	return items, total, err
}

func (r *EmployeeRepository) CountByStatus(tenantID string) (map[models.EmployeeStatus]int64, error) {
	rows := []struct {
		Status models.EmployeeStatus
		Count  int64
	}{}
	err := r.db.Scopes(TenantScope(tenantID)).Model(&models.Employee{}).Select("status, COUNT(*) as count").Group("status").Scan(&rows).Error
	out := map[models.EmployeeStatus]int64{}
	if err != nil {
		return out, err
	}
	for _, rw := range rows {
		out[rw.Status] = rw.Count
	}
	return out, nil
}

func (r *EmployeeRepository) CountActive(tenantID string) (int64, error) {
	var n int64
	err := r.db.Scopes(TenantScope(tenantID)).Model(&models.Employee{}).Where("status = ?", models.EmployeeActive).Count(&n).Error
	return n, err
}

func (r *EmployeeRepository) DirectReports(managerID string) ([]string, error) {
	var ids []string
	err := r.db.Model(&models.Employee{}).Where("manager_id = ?", managerID).Pluck("id", &ids).Error
	return ids, err
}

func (r *EmployeeRepository) ExistsByEmailExcluding(tenantID, email, excludeID string) bool {
	var n int64
	q := r.db.Scopes(TenantScope(tenantID)).Model(&models.Employee{}).Where("LOWER(email) = LOWER(?)", email)
	if excludeID != "" {
		q = q.Where("id <> ?", excludeID)
	}
	q.Count(&n)
	return n > 0
}

func (r *EmployeeRepository) ExistsByCodeExcluding(tenantID, code, excludeID string) bool {
	var n int64
	q := r.db.Scopes(TenantScope(tenantID)).Model(&models.Employee{}).Where("employee_code = ?", code)
	if excludeID != "" {
		q = q.Where("id <> ?", excludeID)
	}
	q.Count(&n)
	return n > 0
}

func (r *EmployeeRepository) Count() (int64, error) {
	var n int64
	err := r.db.Model(&models.Employee{}).Count(&n).Error
	return n, err
}

func sortClause(sortBy, sortOrder string) string {
	allowed := map[string]string{
		"employee_code": "employee_code",
		"first_name":    "first_name",
		"last_name":     "last_name",
		"email":         "email",
		"joining_date":  "joining_date",
		"status":        "status",
		"created_at":    "created_at",
	}
	col, ok := allowed[strings.ToLower(sortBy)]
	if !ok || col == "" {
		return "created_at DESC"
	}
	if strings.ToLower(sortOrder) == "asc" {
		return col + " ASC"
	}
	return col + " DESC"
}
