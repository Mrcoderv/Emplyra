package repositories

import (
	"gorm.io/gorm"

	"github.com/emplyra/backend/internal/models"
	"github.com/emplyra/backend/internal/utils"
)

// --- Documents ---

type DocumentRepository struct {
	db *gorm.DB
}

func NewDocumentRepository(db *gorm.DB) *DocumentRepository { return &DocumentRepository{db: db} }

func (r *DocumentRepository) Create(d *models.Document) error {
	return r.db.Create(d).Error
}

func (r *DocumentRepository) FindByID(id string) (*models.Document, error) {
	var d models.Document
	if err := r.db.Preload("Employee").Preload("Uploader").First(&d, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *DocumentRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&models.Document{}).Error
}

func (r *DocumentRepository) List(p utils.Pagination, employeeID, docType, status string) ([]models.Document, int64, error) {
	q := r.db.Model(&models.Document{})
	if employeeID != "" {
		q = q.Where("employee_id = ?", employeeID)
	}
	if docType != "" {
		q = q.Where("type = ?", docType)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []models.Document
	err := q.Preload("Employee").Preload("Uploader").
		Order("created_at DESC").Offset(p.Offset).Limit(p.Limit).Find(&items).Error
	return items, total, err
}

// --- Notifications ---

type NotificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) ForUser(userID string, p utils.Pagination, unreadOnly bool) ([]models.Notification, int64, error) {
	q := r.db.Model(&models.Notification{}).Where("user_id = ?", userID)
	if unreadOnly {
		q = q.Where("is_read = ?", false)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []models.Notification
	err := q.Order("created_at DESC").Offset(p.Offset).Limit(p.Limit).Find(&items).Error
	return items, total, err
}

func (r *NotificationRepository) CountUnread(userID string) (int64, error) {
	var n int64
	err := r.db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).Count(&n).Error
	return n, err
}

func (r *NotificationRepository) MarkRead(id, userID string) error {
	return r.db.Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(map[string]interface{}{"is_read": true, "read_at": gorm.Expr("NOW()")}).Error
}

func (r *NotificationRepository) MarkAllRead(userID string) error {
	return r.db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Updates(map[string]interface{}{"is_read": true, "read_at": gorm.Expr("NOW()")}).Error
}

// --- Audit logs ---

type AuditLogRepository struct {
	db *gorm.DB
}

func NewAuditLogRepository(db *gorm.DB) *AuditLogRepository { return &AuditLogRepository{db: db} }

func (r *AuditLogRepository) List(p utils.Pagination, userID, resource, action string) ([]models.AuditLog, int64, error) {
	q := r.db.Model(&models.AuditLog{})
	if userID != "" {
		q = q.Where("user_id = ?", userID)
	}
	if resource != "" {
		q = q.Where("resource = ?", resource)
	}
	if action != "" {
		q = q.Where("action = ?", action)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []models.AuditLog
	err := q.Preload("User").Order("created_at DESC").Offset(p.Offset).Limit(p.Limit).Find(&items).Error
	return items, total, err
}
