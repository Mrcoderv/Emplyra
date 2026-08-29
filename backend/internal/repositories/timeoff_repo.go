package repositories

import (
	"strings"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/emplyra/backend/internal/models"
	"github.com/emplyra/backend/internal/utils"
)

type AttendanceRepository struct {
	db *gorm.DB
}

func NewAttendanceRepository(db *gorm.DB) *AttendanceRepository {
	return &AttendanceRepository{db: db}
}

func (r *AttendanceRepository) Create(a *models.Attendance) error {
	return r.db.Create(a).Error
}

func (r *AttendanceRepository) Update(id string, fields map[string]interface{}) error {
	return r.db.Model(&models.Attendance{}).Where("id = ?", id).Updates(fields).Error
}

func (r *AttendanceRepository) FindByID(id string) (*models.Attendance, error) {
	var a models.Attendance
	err := r.db.Preload("Employee").First(&a, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *AttendanceRepository) FindToday(employeeID string, today time.Time) (*models.Attendance, error) {
	var a models.Attendance
	date := datatypes.Date(today)
	err := r.db.Where("employee_id = ? AND date = ?", employeeID, date).First(&a).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *AttendanceRepository) List(p utils.Pagination, employeeID string, from, to datatypes.Date, status string) ([]models.Attendance, int64, error) {
	var items []models.Attendance
	var total int64
	q := r.db.Model(&models.Attendance{})
	if employeeID != "" {
		q = q.Where("employee_id = ?", employeeID)
	}
	if !time.Time(from).IsZero() {
		q = q.Where("date >= ?", from)
	}
	if !time.Time(to).IsZero() {
		q = q.Where("date <= ?", to)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Preload("Employee").Order("date DESC").Offset(p.Offset).Limit(p.Limit).Find(&items).Error
	return items, total, err
}

func (r *AttendanceRepository) Between(employeeID string, from, to datatypes.Date) ([]models.Attendance, error) {
	var items []models.Attendance
	err := r.db.Where("employee_id = ? AND date >= ? AND date <= ?", employeeID, from, to).Find(&items).Error
	return items, err
}

type LeaveRepository struct {
	db *gorm.DB
}

func NewLeaveRepository(db *gorm.DB) *LeaveRepository {
	return &LeaveRepository{db: db}
}

func (r *LeaveRepository) Tx(fn func(tx *gorm.DB) error) error {
	return r.db.Transaction(fn)
}

func (r *LeaveRepository) Create(l *models.Leave) error {
	return r.db.Create(l).Error
}

func (r *LeaveRepository) Update(id string, fields map[string]interface{}) error {
	return r.db.Model(&models.Leave{}).Where("id = ?", id).Updates(fields).Error
}

func (r *LeaveRepository) FindByID(id string) (*models.Leave, error) {
	var l models.Leave
	err := r.db.Preload("Employee").Preload("LeaveType").Preload("Reviewer").First(&l, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *LeaveRepository) List(p utils.Pagination, f LeaveFilter) ([]models.Leave, int64, error) {
	var items []models.Leave
	var total int64
	q := r.db.Model(&models.Leave{})
	q = applyLeaveFilter(q, f)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Preload("Employee").Preload("LeaveType").Preload("Reviewer").
		Order("created_at DESC").Offset(p.Offset).Limit(p.Limit).Find(&items).Error
	return items, total, err
}

func (r *LeaveRepository) Overlaps(employeeID string, start, end datatypes.Date, excludeID string) ([]models.Leave, error) {
	var items []models.Leave
	q := r.db.Where("employee_id = ? AND status IN ?", employeeID, []string{string(models.LeavePending), string(models.LeaveApproved)}).
		Where("start_date <= ? AND end_date >= ?", end, start)
	if excludeID != "" {
		q = q.Where("id <> ?", excludeID)
	}
	err := q.Find(&items).Error
	return items, err
}

type LeaveFilter struct {
	EmployeeID  string
	TypeID      string
	Status      string
	From        datatypes.Date
	To          datatypes.Date
	EmployeeIDs []string
}

func applyLeaveFilter(q *gorm.DB, f LeaveFilter) *gorm.DB {
	if f.EmployeeID != "" {
		q = q.Where("employee_id = ?", f.EmployeeID)
	}
	if f.TypeID != "" {
		q = q.Where("leave_type_id = ?", f.TypeID)
	}
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	if len(f.EmployeeIDs) > 0 {
		q = q.Where("employee_id IN ?", f.EmployeeIDs)
	}
	if !time.Time(f.From).IsZero() {
		q = q.Where("start_date >= ?", f.From)
	}
	if !time.Time(f.To).IsZero() {
		q = q.Where("end_date <= ?", f.To)
	}
	return q
}

type LeaveTypeRepository struct {
	db *gorm.DB
}

func NewLeaveTypeRepository(db *gorm.DB) *LeaveTypeRepository {
	return &LeaveTypeRepository{db: db}
}

func (r *LeaveTypeRepository) List() ([]models.LeaveType, error) {
	var items []models.LeaveType
	err := r.db.Order("name ASC").Find(&items).Error
	return items, err
}

func (r *LeaveTypeRepository) FindByID(id string) (*models.LeaveType, error) {
	var lt models.LeaveType
	err := r.db.First(&lt, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &lt, nil
}

type LeaveBalanceRepository struct {
	db *gorm.DB
}

func NewLeaveBalanceRepository(db *gorm.DB) *LeaveBalanceRepository {
	return &LeaveBalanceRepository{db: db}
}

func (r *LeaveBalanceRepository) Find(employeeID, leaveTypeID string, year int) (*models.LeaveBalance, error) {
	var b models.LeaveBalance
	err := r.db.Where("employee_id = ? AND leave_type_id = ? AND year = ?", employeeID, leaveTypeID, year).First(&b).Error
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *LeaveBalanceRepository) List(employeeID string, year int) ([]models.LeaveBalance, error) {
	var items []models.LeaveBalance
	q := r.db.Order("created_at DESC")
	if employeeID != "" {
		q = q.Where("employee_id = ?", employeeID)
	}
	if year > 0 {
		q = q.Where("year = ?", year)
	}
	err := q.Find(&items).Error
	return items, err
}

func (r *LeaveBalanceRepository) Upsert(b *models.LeaveBalance) error {
	if b.ID != "" {
		return r.db.Save(b).Error
	}
	return r.db.Create(b).Error
}

func (r *LeaveBalanceRepository) IncrementUsed(employeeID, leaveTypeID string, year, days int) error {
	return r.db.Model(&models.LeaveBalance{}).
		Where("employee_id = ? AND leave_type_id = ? AND year = ?", employeeID, leaveTypeID, year).
		UpdateColumn("used", gorm.Expr("used + ?", days)).Error
}

func (r *LeaveBalanceRepository) InitializeIfMissing(employeeID, leaveTypeID string, year, entitlement int) error {
	var count int64
	r.db.Model(&models.LeaveBalance{}).
		Where("employee_id = ? AND leave_type_id = ? AND year = ?", employeeID, leaveTypeID, year).
		Count(&count)
	if count > 0 {
		return nil
	}
	b := &models.LeaveBalance{EmployeeID: employeeID, LeaveTypeID: leaveTypeID, Year: year, Entitlement: entitlement}
	return r.db.Create(b).Error
}

type HolidayRepository struct {
	db *gorm.DB
}

func NewHolidayRepository(db *gorm.DB) *HolidayRepository {
	return &HolidayRepository{db: db}
}

func (r *HolidayRepository) Create(h *models.Holiday) error { return r.db.Create(h).Error }

func (r *HolidayRepository) Update(id string, fields map[string]interface{}) error {
	return r.db.Model(&models.Holiday{}).Where("id = ?", id).Updates(fields).Error
}

func (r *HolidayRepository) Delete(id string) error {
	return r.db.Delete(&models.Holiday{}, "id = ?", id).Error
}

func (r *HolidayRepository) FindByID(id string) (*models.Holiday, error) {
	var h models.Holiday
	err := r.db.First(&h, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &h, nil
}

func (r *HolidayRepository) List(from, to datatypes.Date, status string) ([]models.Holiday, error) {
	q := r.db.Order("date ASC")
	if !time.Time(from).IsZero() {
		q = q.Where("date >= ?", from)
	}
	if !time.Time(to).IsZero() {
		q = q.Where("date <= ?", to)
	}
	if status != "" {
		q = q.Where("status = ?", strings.ToUpper(status))
	}
	var items []models.Holiday
	err := q.Find(&items).Error
	return items, err
}
