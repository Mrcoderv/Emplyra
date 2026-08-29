package models

import (
	"time"

	"gorm.io/datatypes"
)

type JobPostStatus string

const (
	JobOpen   JobPostStatus = "OPEN"
	JobClosed JobPostStatus = "CLOSED"
	JobOnHold JobPostStatus = "ON_HOLD"
)

type JobPost struct {
	BaseModel
	Title        string          `gorm:"size:200;not null" json:"title"`
	DepartmentID *string         `gorm:"type:uuid;index" json:"department_id"`
	Department   *Department     `gorm:"foreignKey:DepartmentID" json:"department,omitempty"`
	Description  string          `gorm:"size:4000" json:"description"`
	Requirements string          `gorm:"size:4000" json:"requirements"`
	Vacancies    int             `gorm:"default:1" json:"vacancies"`
	Status       JobPostStatus   `gorm:"size:20;index;default:OPEN" json:"status"`
	PostedBy     *string         `gorm:"type:uuid;index" json:"posted_by"`
	PostedByUser *User           `gorm:"foreignKey:PostedBy" json:"posted_by_user,omitempty"`
	Deadline     *datatypes.Date `json:"deadline,omitempty"`
}

type CandidateStatus string

const (
	CandidateNew          CandidateStatus = "NEW"
	CandidateScreening    CandidateStatus = "SCREENING"
	CandidateShortlisted  CandidateStatus = "SHORTLISTED"
	CandidateInterviewing CandidateStatus = "INTERVIEWING"
	CandidateOffered      CandidateStatus = "OFFERED"
	CandidateHired        CandidateStatus = "HIRED"
	CandidateRejected     CandidateStatus = "REJECTED"
)

type Candidate struct {
	SoftDeleteModel
	FirstName  string          `gorm:"size:100;not null" json:"first_name"`
	LastName   string          `gorm:"size:100" json:"last_name"`
	Email      string          `gorm:"size:255;not null" json:"email"`
	Phone      string          `gorm:"size:30" json:"phone"`
	ResumePath string          `gorm:"size:500" json:"resume_path,omitempty"`
	Source     string          `gorm:"size:100" json:"source"`
	Status     CandidateStatus `gorm:"size:20;index;default:NEW" json:"status"`
	Notes      string          `gorm:"size:2000" json:"notes"`
}

type ApplicationStatus string

const (
	AppApplied      ApplicationStatus = "APPLIED"
	AppShortlisted  ApplicationStatus = "SHORTLISTED"
	AppInterviewing ApplicationStatus = "INTERVIEWING"
	AppOffered      ApplicationStatus = "OFFERED"
	AppHired        ApplicationStatus = "HIRED"
	AppRejected     ApplicationStatus = "REJECTED"
	AppWithdrawn    ApplicationStatus = "WITHDRAWN"
)

type Application struct {
	BaseModel
	JobPostID   string            `gorm:"type:uuid;not null;index" json:"job_post_id"`
	JobPost     *JobPost          `gorm:"foreignKey:JobPostID" json:"job_post,omitempty"`
	CandidateID string            `gorm:"type:uuid;not null;index" json:"candidate_id"`
	Candidate   *Candidate        `gorm:"foreignKey:CandidateID" json:"candidate,omitempty"`
	AppliedDate datatypes.Date    `gorm:"not null" json:"applied_date"`
	CoverLetter string            `gorm:"size:4000" json:"cover_letter"`
	Status      ApplicationStatus `gorm:"size:20;index;default:APPLIED" json:"status"`
	ReviewerID  *string           `gorm:"type:uuid;index" json:"reviewer_id"`
	Reviewer    *User             `gorm:"foreignKey:ReviewerID" json:"reviewer,omitempty"`
}

type InterviewStatus string

const (
	InterviewScheduled InterviewStatus = "SCHEDULED"
	InterviewCompleted InterviewStatus = "COMPLETED"
	InterviewCancelled InterviewStatus = "CANCELLED"
	InterviewNoShow    InterviewStatus = "NO_SHOW"
)

type InterviewType string

const (
	InterviewPhone     InterviewType = "PHONE"
	InterviewVideo     InterviewType = "VIDEO"
	InterviewInPerson  InterviewType = "IN_PERSON"
	InterviewTechnical InterviewType = "TECHNICAL"
	InterviewHR        InterviewType = "HR"
)

type Interview struct {
	BaseModel
	ApplicationID string          `gorm:"type:uuid;not null;index" json:"application_id"`
	Application   *Application    `gorm:"foreignKey:ApplicationID" json:"application,omitempty"`
	InterviewerID *string         `gorm:"type:uuid;index" json:"interviewer_id"`
	Interviewer   *User           `gorm:"foreignKey:InterviewerID" json:"interviewer,omitempty"`
	ScheduledAt   time.Time       `gorm:"not null" json:"scheduled_at"`
	DurationMin   int             `json:"duration_minutes"`
	Type          InterviewType   `gorm:"size:20" json:"type"`
	Status        InterviewStatus `gorm:"size:20;index;default:SCHEDULED" json:"status"`
	Feedback      string          `gorm:"size:4000" json:"feedback"`
	Score         *float64        `json:"score,omitempty"`
}

type OnboardingStatus string

const (
	OnboardingPending    OnboardingStatus = "PENDING"
	OnboardingInProgress OnboardingStatus = "IN_PROGRESS"
	OnboardingCompleted  OnboardingStatus = "COMPLETED"
)

type Onboarding struct {
	BaseModel
	EmployeeID  string           `gorm:"type:uuid;not null;index" json:"employee_id"`
	Employee    *Employee        `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	CandidateID *string          `gorm:"type:uuid;index" json:"candidate_id"`
	Candidate   *Candidate       `gorm:"foreignKey:CandidateID" json:"candidate,omitempty"`
	StartDate   datatypes.Date   `gorm:"not null" json:"start_date"`
	Status      OnboardingStatus `gorm:"size:20;default:PENDING" json:"status"`
	Tasks       datatypes.JSON   `gorm:"type:jsonb" json:"tasks"`
	Notes       string           `gorm:"size:2000" json:"notes"`
}
