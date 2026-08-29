package repositories

import (
	"time"

	"gorm.io/gorm"

	"github.com/emplyra/backend/internal/models"
)

type PermissionRepository struct {
	db *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) *PermissionRepository {
	return &PermissionRepository{db: db}
}

func (r *PermissionRepository) List() ([]models.Permission, error) {
	var perms []models.Permission
	err := r.db.Order("module ASC, name ASC").Find(&perms).Error
	return perms, err
}

func (r *PermissionRepository) AllPermissionIDs() ([]string, error) {
	var ids []string
	err := r.db.Model(&models.Permission{}).Pluck("id", &ids).Error
	return ids, err
}

type TokenRepository struct {
	db *gorm.DB
}

func NewTokenRepository(db *gorm.DB) *TokenRepository {
	return &TokenRepository{db: db}
}

func (r *TokenRepository) Create(t *models.RefreshToken) error {
	return r.db.Create(t).Error
}

func (r *TokenRepository) FindByHash(hash string) (*models.RefreshToken, error) {
	var t models.RefreshToken
	err := r.db.Where("token_hash = ?", hash).First(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TokenRepository) Revoke(id, replacedBy string) error {
	now := time.Now()
	fields := map[string]interface{}{"revoked_at": &now}
	if replacedBy != "" {
		fields["replaced_by"] = replacedBy
	}
	return r.db.Model(&models.RefreshToken{}).Where("id = ?", id).Updates(fields).Error
}

func (r *TokenRepository) RevokeAllForUser(userID string) error {
	now := time.Now()
	return r.db.Model(&models.RefreshToken{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Updates(map[string]interface{}{"revoked_at": &now}).Error
}

func (r *TokenRepository) DeleteExpired() error {
	return r.db.Where("expires_at < ?", time.Now()).Delete(&models.RefreshToken{}).Error
}
