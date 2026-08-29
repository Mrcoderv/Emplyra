package notifications

import (
	"encoding/json"
	"log/slog"

	"gorm.io/gorm"

	"github.com/emplyra/backend/internal/models"
	"github.com/emplyra/backend/internal/utils"
)

// Provider abstraction for future email/SMS integration.
type Provider interface {
	Send(userID string, title, message, link string, meta map[string]interface{}) error
	Name() string
}

// Service currently persists notifications in the database. Adding a Provider
// (email/SMS) is a drop-in extension without changing callers.
type Service struct {
	db    *gorm.DB
	extra []Provider
}

func New(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) RegisterProvider(p Provider) {
	s.extra = append(s.extra, p)
}

func (s *Service) Notify(userID, title, message string, ntype models.NotificationType, link string, meta map[string]interface{}) error {
	raw := ""
	if meta != nil {
		if b, err := json.Marshal(meta); err == nil {
			raw = string(b)
		}
	}
	n := models.Notification{
		UserID:   userID,
		Title:    title,
		Message:  message,
		Type:     ntype,
		Link:     link,
		Metadata: raw,
	}
	if err := s.db.Create(&n).Error; err != nil {
		return err
	}
	for _, p := range s.extra {
		_ = p.Send(userID, title, message, link, meta)
	}
	return nil
}

// NotifyByEmail routes an out-of-app notification (e.g. to a candidate who does
// not have a user account) through configured providers. Currently a no-op log
// sink; email/SMS providers plug in via RegisterProvider.
func (s *Service) NotifyByEmail(email, recipientName, title, message string, ntype models.NotificationType, link string) error {
	slog.Info("out-of-app notification",
		"to", email, "type", string(ntype), "title", title, "message", message)
	for _, p := range s.extra {
		if ep, ok := p.(emailProvider); ok {
			return ep.SendEmail(email, recipientName, title, message, link)
		}
	}
	return nil
}

type emailProvider interface {
	SendEmail(email, recipientName, title, message, link string) error
}

func (s *Service) Store() *gorm.DB { return s.db }

// List returns a user's notifications, newest first.
func (s *Service) List(userID string, unreadOnly bool, p utils.Pagination) ([]models.Notification, int64, error) {
	q := s.db.Model(&models.Notification{}).Where("user_id = ?", userID)
	if unreadOnly {
		q = q.Where("is_read = ?", false)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []models.Notification
	if err := q.Order("created_at DESC").Offset(p.Offset).Limit(p.Limit).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// UnreadCount returns how many of the user's notifications are unread.
func (s *Service) UnreadCount(userID string) (int64, error) {
	var n int64
	err := s.db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).Count(&n).Error
	return n, err
}

// MarkRead marks one notification as read (scoped to the user).
func (s *Service) MarkRead(userID, id string) error {
	return s.db.Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(map[string]interface{}{"is_read": true, "read_at": gorm.Expr("NOW()")}).Error
}

// MarkAllRead marks every unread notification of the user as read.
func (s *Service) MarkAllRead(userID string) error {
	return s.db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Updates(map[string]interface{}{"is_read": true, "read_at": gorm.Expr("NOW()")}).Error
}
