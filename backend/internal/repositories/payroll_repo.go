package repositories

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/emplyra/backend/internal/models"
	"github.com/emplyra/backend/internal/utils"
)

type SalaryStructureRepository struct {
	db *gorm.DB
}

func NewSalaryStructureRepository(db *gorm.DB) *SalaryStructureRepository {
	return &SalaryStructureRepository{db: db}
}

func (r *SalaryStructureRepository) Create(s *models.SalaryStructure) error {
	return r.db.Create(s).Error
}
func (r *SalaryStructureRepository) Update(id string, fields map[string]interface{}) error {
	return r.db.Model(&models.SalaryStructure{}).Where("id = ?", id).Updates(fields).Error
}
func (r *SalaryStructureRepository) Delete(id string) error {
	return r.db.Delete(&models.SalaryStructure{}, "id = ?", id).Error
}
func (r *SalaryStructureRepository) FindByID(id string) (*models.SalaryStructure, error) {
	var s models.SalaryStructure
	err := r.db.Preload("Employee").First(&s, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// EffectiveFor returns the active salary structure for an employee on/after a date.
func (r *SalaryStructureRepository) EffectiveFor(employeeID string, on datatypes.Date) (*models.SalaryStructure, error) {
	var s models.SalaryStructure
	err := r.db.Preload("Employee").
		Where("employee_id = ? AND status = ? AND effective_from <= ?", employeeID, models.SalaryActive, on).
		Order("effective_from DESC").First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SalaryStructureRepository) List(employeeID string) ([]models.SalaryStructure, error) {
	q := r.db.Preload("Employee").Order("effective_from DESC")
	if employeeID != "" {
		q = q.Where("employee_id = ?", employeeID)
	}
	var items []models.SalaryStructure
	err := q.Find(&items).Error
	return items, err
}

func (r *SalaryStructureRepository) Tx(fn func(tx *gorm.DB) error) error {
	return r.db.Transaction(fn)
}

type PayrollRepository struct {
	db *gorm.DB
}

func NewPayrollRepository(db *gorm.DB) *PayrollRepository {
	return &PayrollRepository{db: db}
}

func (r *PayrollRepository) CreateMany(items []models.Payroll) error {
	if len(items) == 0 {
		return nil
	}
	return r.db.Create(&items).Error
}

func (r *PayrollRepository) Update(id string, fields map[string]interface{}) error {
	return r.db.Model(&models.Payroll{}).Where("id = ?", id).Updates(fields).Error
}

func (r *PayrollRepository) FindByID(id string) (*models.Payroll, error) {
	var p models.Payroll
	err := r.db.Preload("Employee").First(&p, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PayrollRepository) List(p utils.Pagination, month, year int, employeeID, status, departmentID string) ([]models.Payroll, int64, error) {
	var items []models.Payroll
	var total int64
	q := r.db.Model(&models.Payroll{})
	if month > 0 {
		q = q.Where("month = ?", month)
	}
	if year > 0 {
		q = q.Where("year = ?", year)
	}
	if employeeID != "" {
		q = q.Where("employee_id = ?", employeeID)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if departmentID != "" {
		q = q.Joins("JOIN employees ON employees.id = payroll.employee_id").
			Where("employees.department_id = ?", departmentID)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Preload("Employee.Department").Order("year DESC, month DESC, created_at DESC").
		Offset(p.Offset).Limit(p.Limit).Find(&items).Error
	return items, total, err
}

func (r *PayrollRepository) ExistsForMonth(month, year int) (bool, error) {
	var n int64
	err := r.db.Model(&models.Payroll{}).Where("month = ? AND year = ?", month, year).Count(&n).Error
	return n > 0, err
}

func (r *PayrollRepository) ExistsForEmployee(employeeID string, month, year int) (bool, error) {
	var n int64
	err := r.db.Model(&models.Payroll{}).Where("employee_id = ? AND month = ? AND year = ?", employeeID, month, year).Count(&n).Error
	return n > 0, err
}

func (r *PayrollRepository) Tx(fn func(tx *gorm.DB) error) error {
	return r.db.Transaction(fn)
}

// ActiveEmployeesWithSalary returns active employees that have an active salary
// structure effective on or before `on`.
func (r *PayrollRepository) ActiveEmployeesWithSalary(on datatypes.Date) ([]models.Employee, error) {
	var items []models.Employee
	err := r.db.Model(&models.Employee{}).
		Joins("JOIN salary_structures ss ON ss.employee_id = employees.id").
		Where("employees.status = ? AND ss.status = ? AND ss.effective_from <= ?", models.EmployeeActive, models.SalaryActive, on).
		Group("employees.id").
		Find(&items).Error
	return items, err
}

func (r *PayrollRepository) EmployeeAttendanceOvertime(employeeID string, month, year int) (float64, error) {
	var sum float64
	err := r.db.Model(&models.Attendance{}).
		Where("employee_id = ? AND EXTRACT(MONTH FROM date) = ? AND EXTRACT(YEAR FROM date) = ?", employeeID, month, year).
		Select("COALESCE(SUM(overtime), 0)").
		Scan(&sum).Error
	return sum, err
}

func (r *PayrollRepository) EmployeeByID(id string) (*models.Employee, error) {
	var e models.Employee
	err := r.db.Preload("Department").First(&e, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &e, nil
}
