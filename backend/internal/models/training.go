package models

import (
	"time"

	"gorm.io/datatypes"
)

type TrainingStatus string

const (
	TrainingScheduled TrainingStatus = "SCHEDULED"
	TrainingOngoing   TrainingStatus = "ONGOING"
	TrainingCompleted TrainingStatus = "COMPLETED"
	TrainingCancelled TrainingStatus = "CANCELLED"
)

type TrainingProgram struct {
	BaseModel
	TenantID    string         `gorm:"type:uuid;not null;index;default:00000000-0000-0000-0000-000000000001" json:"tenant_id"`
	Title       string         `gorm:"size:200;not null" json:"title"`
	Description string         `gorm:"size:4000" json:"description"`
	Provider    string         `gorm:"size:200" json:"provider"`
	StartDate   datatypes.Date `gorm:"not null" json:"start_date"`
	EndDate     datatypes.Date `gorm:"not null" json:"end_date"`
	Location    string         `gorm:"size:200" json:"location"`
	MaxSeats    int            `json:"max_seats"`
	Status      TrainingStatus `gorm:"size:20;index;default:SCHEDULED" json:"status"`
}

type TrainingSchedule struct {
	BaseModel
	TenantID  string           `gorm:"type:uuid;not null;index;default:00000000-0000-0000-0000-000000000001" json:"tenant_id"`
	ProgramID string           `gorm:"type:uuid;not null;index" json:"program_id"`
	Program   *TrainingProgram `gorm:"foreignKey:ProgramID" json:"program,omitempty"`
	Date      datatypes.Date   `gorm:"not null" json:"date"`
	StartTime string           `gorm:"size:10" json:"start_time"`
	EndTime   string           `gorm:"size:10" json:"end_time"`
	Trainer   string           `gorm:"size:200" json:"trainer"`
	Location  string           `gorm:"size:200" json:"location"`
	MaxSeats  int              `json:"max_seats"`
}

type EnrollmentStatus string

const (
	EnrollmentEnrolled   EnrollmentStatus = "ENROLLED"
	EnrollmentInProgress EnrollmentStatus = "IN_PROGRESS"
	EnrollmentCompleted  EnrollmentStatus = "COMPLETED"
	EnrollmentCancelled  EnrollmentStatus = "CANCELLED"
)

type TrainingEnrollment struct {
	BaseModel
	TenantID    string           `gorm:"type:uuid;not null;index;default:00000000-0000-0000-0000-000000000001" json:"tenant_id"`
	ProgramID   string           `gorm:"type:uuid;not null;index" json:"program_id"`
	Program     *TrainingProgram `gorm:"foreignKey:ProgramID" json:"program,omitempty"`
	EmployeeID  string           `gorm:"type:uuid;not null;index" json:"employee_id"`
	Employee    *Employee        `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	Status      EnrollmentStatus `gorm:"size:20;index;default:ENROLLED" json:"status"`
	EnrolledAt  time.Time        `json:"enrolled_at"`
	CompletedAt *time.Time       `json:"completed_at,omitempty"`
}
