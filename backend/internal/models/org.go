package models

import "time"

type OrgEntityStatus string

const (
	OrgStatusActive   OrgEntityStatus = "ACTIVE"
	OrgStatusInactive OrgEntityStatus = "INACTIVE"
)

type Department struct {
	BaseModel
	TenantID    string          `gorm:"type:uuid;not null;uniqueIndex:idx_dept_code_tenant,priority:2;default:00000000-0000-0000-0000-000000000001" json:"tenant_id"`
	Name        string          `gorm:"size:150;not null" json:"name"`
	Code        string          `gorm:"size:50;uniqueIndex:idx_dept_code_tenant,priority:1;not null" json:"code"`
	Description string          `gorm:"size:500" json:"description"`
	ManagerID   *string         `gorm:"type:uuid;index" json:"manager_id"`
	Manager     *Employee       `gorm:"foreignKey:ManagerID" json:"manager,omitempty"`
	Status      OrgEntityStatus `gorm:"size:20;default:ACTIVE" json:"status"`
}

type Designation struct {
	BaseModel
	TenantID     string          `gorm:"type:uuid;not null;index;default:00000000-0000-0000-0000-000000000001" json:"tenant_id"`
	Name         string          `gorm:"size:150;not null" json:"name"`
	Description  string          `gorm:"size:500" json:"description"`
	DepartmentID *string         `gorm:"type:uuid;index" json:"department_id"`
	Department   *Department     `gorm:"foreignKey:DepartmentID" json:"department,omitempty"`
	Level        int             `gorm:"default:1" json:"level"`
	Status       OrgEntityStatus `gorm:"size:20;default:ACTIVE" json:"status"`
}

type EmploymentType string

const (
	EmploymentFullTime  EmploymentType = "FULL_TIME"
	EmploymentPartTime  EmploymentType = "PART_TIME"
	EmploymentContract  EmploymentType = "CONTRACT"
	EmploymentIntern    EmploymentType = "INTERN"
	EmploymentProbation EmploymentType = "PROBATION"
)

type Gender string

const (
	GenderMale   Gender = "MALE"
	GenderFemale Gender = "FEMALE"
	GenderOther  Gender = "OTHER"
)

type EmployeeStatus string

const (
	EmployeeActive            EmployeeStatus = "ACTIVE"
	EmployeeInactive          EmployeeStatus = "INACTIVE"
	EmployeeOnLeave           EmployeeStatus = "ON_LEAVE"
	EmployeeTerminated        EmployeeStatus = "TERMINATED"
	EmployeeOnboardingProceed EmployeeStatus = "ONBOARDING"
)

type Employee struct {
	SoftDeleteModel
	TenantID         string         `gorm:"type:uuid;not null;uniqueIndex:idx_emp_code_tenant,priority:2;uniqueIndex:idx_emp_email_tenant,priority:2;default:00000000-0000-0000-0000-000000000001" json:"tenant_id"`
	EmployeeCode     string         `gorm:"size:50;uniqueIndex:idx_emp_code_tenant,priority:1;not null" json:"employee_code"`
	FirstName        string         `gorm:"size:100;not null" json:"first_name"`
	LastName         string         `gorm:"size:100" json:"last_name"`
	Email            string         `gorm:"size:255;uniqueIndex:idx_emp_email_tenant,priority:1;not null" json:"email"`
	Phone            string         `gorm:"size:30" json:"phone"`
	DateOfBirth      *time.Time     `json:"date_of_birth,omitempty"`
	Gender           *Gender        `gorm:"size:20" json:"gender,omitempty"`
	Address          string         `gorm:"size:500" json:"address"`
	EmergencyContact string         `gorm:"size:255" json:"emergency_contact"`
	JoiningDate      *time.Time     `json:"joining_date,omitempty"`
	EmploymentType   EmploymentType `gorm:"size:20;default:FULL_TIME" json:"employment_type"`
	Status           EmployeeStatus `gorm:"size:20;index;default:ACTIVE" json:"status"`
	DepartmentID     *string        `gorm:"type:uuid;index" json:"department_id"`
	Department       *Department    `gorm:"foreignKey:DepartmentID" json:"department,omitempty"`
	DesignationID    *string        `gorm:"type:uuid;index" json:"designation_id"`
	Designation      *Designation   `gorm:"foreignKey:DesignationID" json:"designation,omitempty"`
	ManagerID        *string        `gorm:"type:uuid;index" json:"manager_id"`
	Manager          *Employee      `gorm:"foreignKey:ManagerID" json:"manager,omitempty"`
	UserID           *string        `gorm:"type:uuid;index" json:"user_id"`
	User             *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (e Employee) FullName() string {
	if e.LastName == "" {
		return e.FirstName
	}
	return e.FirstName + " " + e.LastName
}
