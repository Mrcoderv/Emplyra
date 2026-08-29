package models

import (
	"time"

	"gorm.io/datatypes"
)

const (
	GoogleFormStatusPending      = "PENDING"
	GoogleFormStatusConnected    = "CONNECTED"
	GoogleFormStatusDisconnected = "DISCONNECTED"
	GoogleFormStatusError        = "ERROR"
)

const (
	GoogleFormResponseImported  = "IMPORTED"
	GoogleFormResponseDuplicate = "DUPLICATE"
	GoogleFormResponseError     = "ERROR"
)

type GoogleFormIntegration struct {
	BaseModel
	TenantID      string         `gorm:"type:uuid;not null;index;default:00000000-0000-0000-0000-000000000001" json:"tenant_id"`
	JobID         string         `gorm:"type:uuid;not null;uniqueIndex" json:"job_id"`
	JobPost       *JobPost       `gorm:"foreignKey:JobID" json:"job_post,omitempty"`
	Provider      string         `gorm:"size:40;not null;default:google_forms" json:"provider"`
	FormURL       string         `gorm:"size:500" json:"google_form_url"`
	SpreadsheetID string         `gorm:"size:200" json:"google_sheet_id"`
	SheetName     string         `gorm:"size:200;column:response_sheet_name" json:"response_sheet_name"`
	HeaderRow     int            `gorm:"default:1" json:"header_row"`
	FieldMapping  datatypes.JSON `gorm:"type:jsonb" json:"field_mapping"`
	Status        string         `gorm:"size:20;index;default:PENDING" json:"status"`
	LastSyncedAt  *time.Time     `json:"last_synced_at,omitempty"`
	SyncedRows    int            `gorm:"default:0" json:"synced_rows"`
	SyncError     string         `gorm:"size:1000" json:"sync_error,omitempty"`
	StatusDetail  string         `gorm:"size:255" json:"status_detail,omitempty"`
}

type GoogleFormResponse struct {
	BaseModel
	TenantID           string                 `gorm:"type:uuid;not null;index;default:00000000-0000-0000-0000-000000000001" json:"tenant_id"`
	IntegrationID      string                 `gorm:"type:uuid;not null;index" json:"integration_id"`
	Integration        *GoogleFormIntegration `gorm:"foreignKey:IntegrationID" json:"integration,omitempty"`
	ExternalResponseID string                 `gorm:"size:255;not null;uniqueIndex" json:"external_response_id"`
	CandidateID        *string                `gorm:"type:uuid;index" json:"candidate_id,omitempty"`
	ApplicationID      *string                `gorm:"type:uuid;index" json:"application_id,omitempty"`
	RawResponse        datatypes.JSON         `gorm:"type:jsonb" json:"raw_response"`
	SubmittedAt        *time.Time             `json:"submitted_at,omitempty"`
	ImportedAt         *time.Time             `json:"imported_at,omitempty"`
	Status             string                 `gorm:"size:20;index;default:IMPORTED" json:"status"`
	ErrorMessage       string                 `gorm:"size:1000" json:"error_message,omitempty"`
}

type GoogleOAuthToken struct {
	BaseModel
	Key       string     `gorm:"size:64;not null;uniqueIndex" json:"key"`
	Data      string     `gorm:"type:text" json:"data,omitempty"`
	ExpiresAt *time.Time `gorm:"index" json:"expires_at,omitempty"`
}
