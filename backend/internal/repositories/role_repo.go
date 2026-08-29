package repositories

import (
	"gorm.io/gorm"

	"github.com/emplyra/backend/internal/models"
)

type RoleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

func (r *RoleRepository) FindByName(name string) (*models.Role, error) {
	var role models.Role
	err := r.db.Where("name = ?", name).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *RoleRepository) FindByID(id string) (*models.Role, error) {
	var role models.Role
	err := r.db.Preload("Permissions").First(&role, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *RoleRepository) List() ([]models.Role, error) {
	var roles []models.Role
	err := r.db.Preload("Permissions").Order("name ASC").Find(&roles).Error
	return roles, err
}

// ListByScope returns only roles anchored to the given scope (e.g. tenant).
func (r *RoleRepository) ListByScope(scope string) ([]models.Role, error) {
	var roles []models.Role
	err := r.db.Preload("Permissions").Where("scope = ?", scope).Order("name ASC").Find(&roles).Error
	return roles, err
}

func (r *RoleRepository) Create(role *models.Role, permissionIDs []string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(role).Error; err != nil {
			return err
		}
		return tx.Model(role).Association("Permissions").Replace(r.permissionModels(tx, permissionIDs))
	})
}

func (r *RoleRepository) Update(role *models.Role, permissionIDs []string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(role).Updates(map[string]interface{}{
			"name":        role.Name,
			"description": role.Description,
		}).Error; err != nil {
			return err
		}
		return tx.Model(role).Association("Permissions").Replace(r.permissionModels(tx, permissionIDs))
	})
}

func (r *RoleRepository) Delete(id string) error {
	var count int64
	r.db.Model(&models.User{}).Where("role_id = ?", id).Count(&count)
	if count > 0 {
		return ErrRoleInUse
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		role := models.Role{BaseModel: models.BaseModel{ID: id}}
		if err := tx.Model(&role).Association("Permissions").Clear(); err != nil {
			return err
		}
		return tx.Delete(&models.Role{}, "id = ?", id).Error
	})
}

func (r *RoleRepository) permissionModels(tx *gorm.DB, ids []string) []models.Permission {
	if len(ids) == 0 {
		return nil
	}
	var perms []models.Permission
	tx.Where("id IN ?", ids).Find(&perms)
	return perms
}

var ErrRoleInUse = errorString("role is assigned to users")

type errorString string

func (e errorString) Error() string { return string(e) }
