package repositories

import (
	"gorm.io/gorm"

	"github.com/emplyra/backend/internal/models"
	"github.com/emplyra/backend/internal/utils"
)

type TrainingRepository struct {
	db *gorm.DB
}

func NewTrainingRepository(db *gorm.DB) *TrainingRepository { return &TrainingRepository{db: db} }

func (r *TrainingRepository) Create(m *models.TrainingProgram) error { return r.db.Create(m).Error }
func (r *TrainingRepository) Update(tenantID, id string, f map[string]interface{}) error {
	return r.db.Model(&models.TrainingProgram{}).Scopes(TenantScope(tenantID)).Where("id = ?", id).Updates(f).Error
}
func (r *TrainingRepository) Delete(tenantID, id string) error {
	return r.db.Scopes(TenantScope(tenantID)).Delete(&models.TrainingProgram{}, "id = ?", id).Error
}
func (r *TrainingRepository) FindByID(tenantID, id string) (*models.TrainingProgram, error) {
	var m models.TrainingProgram
	err := r.db.Scopes(TenantScope(tenantID)).First(&m, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}
func (r *TrainingRepository) List(tenantID string, p utils.Pagination, status string) ([]models.TrainingProgram, int64, error) {
	var items []models.TrainingProgram
	var total int64
	q := r.db.Scopes(TenantScope(tenantID)).Model(&models.TrainingProgram{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Order("start_date DESC").Offset(p.Offset).Limit(p.Limit).Find(&items).Error
	return items, total, err
}

type TrainingScheduleRepository struct {
	db *gorm.DB
}

func NewTrainingScheduleRepository(db *gorm.DB) *TrainingScheduleRepository {
	return &TrainingScheduleRepository{db: db}
}

func (r *TrainingScheduleRepository) Create(m *models.TrainingSchedule) error {
	return r.db.Create(m).Error
}
func (r *TrainingScheduleRepository) Update(tenantID, id string, f map[string]interface{}) error {
	return r.db.Model(&models.TrainingSchedule{}).Scopes(TenantScope(tenantID)).Where("id = ?", id).Updates(f).Error
}
func (r *TrainingScheduleRepository) Delete(tenantID, id string) error {
	return r.db.Scopes(TenantScope(tenantID)).Delete(&models.TrainingSchedule{}, "id = ?", id).Error
}
func (r *TrainingScheduleRepository) FindByID(tenantID, id string) (*models.TrainingSchedule, error) {
	var m models.TrainingSchedule
	err := r.db.Scopes(TenantScope(tenantID)).Preload("Program").First(&m, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}
func (r *TrainingScheduleRepository) ListByProgram(tenantID, programID string) ([]models.TrainingSchedule, error) {
	var items []models.TrainingSchedule
	q := r.db.Scopes(TenantScope(tenantID)).Preload("Program").Order("date ASC")
	if programID != "" {
		q = q.Where("program_id = ?", programID)
	}
	err := q.Find(&items).Error
	return items, err
}

type EnrollmentRepository struct {
	db *gorm.DB
}

func NewEnrollmentRepository(db *gorm.DB) *EnrollmentRepository { return &EnrollmentRepository{db: db} }

func (r *EnrollmentRepository) Create(m *models.TrainingEnrollment) error {
	return r.db.Create(m).Error
}
func (r *EnrollmentRepository) Update(tenantID, id string, f map[string]interface{}) error {
	return r.db.Model(&models.TrainingEnrollment{}).Scopes(TenantScope(tenantID)).Where("id = ?", id).Updates(f).Error
}
func (r *EnrollmentRepository) FindByID(tenantID, id string) (*models.TrainingEnrollment, error) {
	var m models.TrainingEnrollment
	err := r.db.Scopes(TenantScope(tenantID)).Preload("Program").Preload("Employee").First(&m, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}
func (r *EnrollmentRepository) Exists(tenantID, programID, employeeID string) (bool, error) {
	var n int64
	err := r.db.Scopes(TenantScope(tenantID)).Model(&models.TrainingEnrollment{}).
		Where("program_id = ? AND employee_id = ?", programID, employeeID).Count(&n).Error
	return n > 0, err
}
func (r *EnrollmentRepository) List(tenantID string, p utils.Pagination, programID, employeeID, status string) ([]models.TrainingEnrollment, int64, error) {
	var items []models.TrainingEnrollment
	var total int64
	q := r.db.Scopes(TenantScope(tenantID)).Model(&models.TrainingEnrollment{})
	if programID != "" {
		q = q.Where("program_id = ?", programID)
	}
	if employeeID != "" {
		q = q.Where("employee_id = ?", employeeID)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Preload("Program").Preload("Employee").Order("enrolled_at DESC").Offset(p.Offset).Limit(p.Limit).Find(&items).Error
	return items, total, err
}
