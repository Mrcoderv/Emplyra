package repositories

import (
	"gorm.io/gorm"

	"github.com/emplyra/backend/internal/models"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	var u models.User
	err := r.db.Preload("Role").Where("LOWER(email) = LOWER(?)", email).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) FindByUsername(username string) (*models.User, error) {
	var u models.User
	err := r.db.Preload("Role").Where("LOWER(username) = LOWER(?)", username).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) FindByEmailOrUsername(identifier string) (*models.User, error) {
	var u models.User
	err := r.db.Preload("Role").
		Where("LOWER(email) = LOWER(?) OR LOWER(username) = LOWER(?)", identifier, identifier).
		First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) FindByID(id string) (*models.User, error) {
	var u models.User
	err := r.db.Preload("Role").First(&u, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) Create(u *models.User) error {
	return r.db.Create(u).Error
}

func (r *UserRepository) Update(id string, fields map[string]interface{}) error {
	return r.db.Model(&models.User{}).Where("id = ?", id).Updates(fields).Error
}

func (r *UserRepository) Delete(id string) error {
	return r.db.Delete(&models.User{}, "id = ?", id).Error
}

func (r *UserRepository) List(pagination func(*gorm.DB) *gorm.DB, filter func(*gorm.DB) *gorm.DB) ([]models.User, int64, error) {
	var users []models.User
	var total int64
	q := r.db.Model(&models.User{})
	if filter != nil {
		q = filter(q)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if pagination != nil {
		q = pagination(q)
	}
	err := q.Preload("Role").Order("created_at DESC").Find(&users).Error
	return users, total, err
}

func (r *UserRepository) RolePermissions(roleID string) ([]string, error) {
	var perms []string
	err := r.db.Model(&models.Permission{}).
		Joins("JOIN role_permissions rp ON rp.permission_id = permissions.id").
		Where("rp.role_id = ?", roleID).
		Pluck("permissions.name", &perms).Error
	return perms, err
}

func (r *UserRepository) Count() (int64, error) {
	var n int64
	err := r.db.Model(&models.User{}).Count(&n).Error
	return n, err
}
