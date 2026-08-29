package repositories

import (
	"gorm.io/gorm"

	"github.com/emplyra/backend/internal/models"
)

type DepartmentRepository struct {
	db *gorm.DB
}

func NewDepartmentRepository(db *gorm.DB) *DepartmentRepository {
	return &DepartmentRepository{db: db}
}

func (r *DepartmentRepository) Create(d *models.Department) error { return r.db.Create(d).Error }
func (r *DepartmentRepository) Update(id string, fields map[string]interface{}) error {
	return r.db.Model(&models.Department{}).Where("id = ?", id).Updates(fields).Error
}
func (r *DepartmentRepository) Delete(id string) error {
	return r.db.Delete(&models.Department{}, "id = ?", id).Error
}
func (r *DepartmentRepository) FindByID(id string) (*models.Department, error) {
	var d models.Department
	err := r.db.Preload("Manager").First(&d, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &d, nil
}
func (r *DepartmentRepository) FindByCode(code string) (*models.Department, error) {
	var d models.Department
	err := r.db.Where("code = ?", code).First(&d).Error
	if err != nil {
		return nil, err
	}
	return &d, nil
}
func (r *DepartmentRepository) List(filter func(*gorm.DB) *gorm.DB) ([]models.Department, error) {
	var deps []models.Department
	q := r.db.Preload("Manager").Order("name ASC")
	if filter != nil {
		q = filter(q)
	}
	err := q.Find(&deps).Error
	return deps, err
}

type DesignationRepository struct {
	db *gorm.DB
}

func NewDesignationRepository(db *gorm.DB) *DesignationRepository {
	return &DesignationRepository{db: db}
}

func (r *DesignationRepository) Create(d *models.Designation) error { return r.db.Create(d).Error }
func (r *DesignationRepository) Update(id string, fields map[string]interface{}) error {
	return r.db.Model(&models.Designation{}).Where("id = ?", id).Updates(fields).Error
}
func (r *DesignationRepository) Delete(id string) error {
	return r.db.Delete(&models.Designation{}, "id = ?", id).Error
}
func (r *DesignationRepository) FindByID(id string) (*models.Designation, error) {
	var d models.Designation
	err := r.db.Preload("Department").First(&d, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &d, nil
}
func (r *DesignationRepository) List() ([]models.Designation, error) {
	var items []models.Designation
	err := r.db.Preload("Department").Order("level ASC, name ASC").Find(&items).Error
	return items, err
}
