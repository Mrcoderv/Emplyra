package auditmanager

import (
	"encoding/json"

	"gorm.io/gorm"

	"github.com/emplyra/backend/internal/models"
)

type Service struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) Record(userID string, action models.AuditAction, resource, resourceID, ip, userAgent string, metadata interface{}) {
	meta := ""
	if metadata != nil {
		if b, err := json.Marshal(metadata); err == nil {
			meta = string(b)
		}
	}
	var uid *string
	if userID != "" {
		uid = &userID
	}
	entry := models.AuditLog{
		UserID:     uid,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		IPAddress:  ip,
		UserAgent:  userAgent,
		Metadata:   meta,
	}
	_ = s.db.Create(&entry).Error
}

func (s *Service) RecordWithError(userID string, action models.AuditAction, resource, resourceID, ip, userAgent string, metadata interface{}) error {
	s.Record(userID, action, resource, resourceID, ip, userAgent, metadata)
	return nil
}

func (s *Service) FailedLogin(identifier, ip, userAgent string) {
	s.Record("", models.ActionLoginFailed, "auth", identifier, ip, userAgent, nil)
}

func (s *Service) Store() *gorm.DB { return s.db }
