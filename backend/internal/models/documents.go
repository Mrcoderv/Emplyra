package models

type DocumentStatus string

const (
	DocumentActive   DocumentStatus = "ACTIVE"
	DocumentArchived DocumentStatus = "ARCHIVED"
)

type DocumentType string

const (
	DocContract    DocumentType = "CONTRACT"
	DocCertificate DocumentType = "CERTIFICATE"
	DocIdentity    DocumentType = "IDENTITY"
	DocAppointment DocumentType = "APPOINTMENT"
	DocExperience  DocumentType = "EXPERIENCE"
	DocOther       DocumentType = "OTHER"
)

type Document struct {
	BaseModel
	EmployeeID string         `gorm:"type:uuid;not null;index" json:"employee_id"`
	Employee   *Employee      `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	Title      string         `gorm:"size:200;not null" json:"title"`
	Type       DocumentType   `gorm:"size:30;index" json:"type"`
	FilePath   string         `gorm:"size:500;not null" json:"-"`
	MimeType   string         `gorm:"size:100" json:"mime_type"`
	SizeBytes  int64          `json:"size_bytes"`
	Status     DocumentStatus `gorm:"size:20;default:ACTIVE" json:"status"`
	UploadedBy *string        `gorm:"type:uuid;index" json:"uploaded_by,omitempty"`
	Uploader   *User          `gorm:"foreignKey:UploadedBy" json:"uploader,omitempty"`
}
