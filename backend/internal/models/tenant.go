package models

import "time"

type TenantStatus string

const (
	TenantActive    TenantStatus = "ACTIVE"
	TenantTrial     TenantStatus = "TRIAL"
	TenantSuspended TenantStatus = "SUSPENDED"
	TenantInactive  TenantStatus = "INACTIVE"
)

type TenantPlan string

const (
	PlanFree         TenantPlan = "FREE"
	PlanProfessional TenantPlan = "PROFESSIONAL"
	PlanEnterprise   TenantPlan = "ENTERPRISE"
)

// DefaultTenantID is the sentinel tenant that owns legacy pre-SaaS rows.
// The seeder guarantees a Tenant row with this ID exists.
const DefaultTenantID = "00000000-0000-0000-0000-000000000001"

type Tenant struct {
	BaseModel
	Name        string       `gorm:"size:200;not null" json:"name"`
	Slug        string       `gorm:"size:120;uniqueIndex;not null" json:"slug"`
	Status      TenantStatus `gorm:"size:20;index;default:ACTIVE" json:"status"`
	Plan        TenantPlan   `gorm:"size:20;default:FREE" json:"plan"`
	TrialEndsAt *time.Time   `json:"trial_ends_at,omitempty"`
	Industry    string       `gorm:"size:100" json:"industry"`
	Settings    string       `gorm:"type:jsonb" json:"settings"`
	CreatedBy   *string      `gorm:"type:uuid" json:"created_by,omitempty"`
}

func (t Tenant) IsOperational() bool {
	return t.Status == TenantActive || t.Status == TenantTrial
}

// TenantUsage carries live counters for a tenant (platform usage views).
type TenantUsage struct {
	Users       int64 `json:"users"`
	Employees   int64 `json:"employees"`
	Departments int64 `json:"departments"`
	Jobs        int64 `json:"jobs"`
	Candidates  int64 `json:"candidates"`
	Documents   int64 `json:"documents"`
}
