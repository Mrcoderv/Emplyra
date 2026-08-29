package models

import (
	"time"

	"gorm.io/datatypes"
)

type GoalStatus string

const (
	GoalNotStarted GoalStatus = "NOT_STARTED"
	GoalInProgress GoalStatus = "IN_PROGRESS"
	GoalAchieved   GoalStatus = "ACHIEVED"
	GoalCancelled  GoalStatus = "CANCELLED"
)

type Goal struct {
	BaseModel
	EmployeeID  string          `gorm:"type:uuid;not null;index" json:"employee_id"`
	Employee    *Employee       `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	Title       string          `gorm:"size:200;not null" json:"title"`
	Description string          `gorm:"size:2000" json:"description"`
	TargetDate  *datatypes.Date `json:"target_date,omitempty"`
	Weight      int             `gorm:"default:1" json:"weight"`
	Status      GoalStatus      `gorm:"size:20;default:NOT_STARTED" json:"status"`
}

type KPI struct {
	BaseModel
	EmployeeID  string    `gorm:"type:uuid;not null;index" json:"employee_id"`
	Employee    *Employee `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	Name        string    `gorm:"size:200;not null" json:"name"`
	Description string    `gorm:"size:1000" json:"description"`
	Target      string    `gorm:"size:100" json:"target"`
	Actual      string    `gorm:"size:100" json:"actual"`
	Unit        string    `gorm:"size:50" json:"unit"`
	Weight      int       `gorm:"default:1" json:"weight"`
	Period      string    `gorm:"size:50" json:"period"`
	Score       *float64  `json:"score,omitempty"`
}

type ReviewStatus string

const (
	ReviewPending       ReviewStatus = "PENDING"
	ReviewSelfSubmitted ReviewStatus = "SELF_SUBMITTED"
	ReviewManagerDone   ReviewStatus = "MANAGER_FEEDBACK"
	ReviewCompleted     ReviewStatus = "COMPLETED"
)

type PerformanceReview struct {
	BaseModel
	EmployeeID      string          `gorm:"type:uuid;not null;index" json:"employee_id"`
	Employee        *Employee       `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	ReviewerID      *string         `gorm:"type:uuid;index" json:"reviewer_id"`
	Reviewer        *Employee       `gorm:"foreignKey:ReviewerID" json:"reviewer,omitempty"`
	Period          string          `gorm:"size:50;not null" json:"period"`
	SelfEvaluation  string          `gorm:"size:4000" json:"self_evaluation"`
	ManagerFeedback string          `gorm:"size:4000" json:"manager_feedback"`
	Score           *float64        `json:"score,omitempty"`
	Status          ReviewStatus    `gorm:"size:30;default:PENDING" json:"status"`
	DueDate         *datatypes.Date `json:"due_date,omitempty"`
	CompletedAt     *time.Time      `json:"completed_at,omitempty"`
}
