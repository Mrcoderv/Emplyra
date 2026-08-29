package models

import (
	"time"

	"gorm.io/datatypes"
)

type SalaryStructureStatus string

const (
	SalaryActive   SalaryStructureStatus = "ACTIVE"
	SalaryInactive SalaryStructureStatus = "INACTIVE"
)

type SalaryStructure struct {
	BaseModel
	TenantID       string                `gorm:"type:uuid;not null;index;default:00000000-0000-0000-0000-000000000001" json:"tenant_id"`
	EmployeeID     string                `gorm:"type:uuid;not null;index" json:"employee_id"`
	Employee       *Employee             `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	BasicSalary    Decimal               `gorm:"type:numeric(14,2);not null" json:"basic_salary"`
	Allowances     Decimal               `gorm:"type:numeric(14,2);default:0" json:"allowances"`
	Bonus          Decimal               `gorm:"type:numeric(14,2);default:0" json:"bonus"`
	OvertimeRate   Decimal               `gorm:"type:numeric(14,2);default:0" json:"overtime_rate"`
	TaxRate        Decimal               `gorm:"type:numeric(5,2);default:0" json:"tax_rate"`
	TaxAmount      Decimal               `gorm:"type:numeric(14,2);default:0" json:"tax_amount"`
	Deductions     Decimal               `gorm:"type:numeric(14,2);default:0" json:"deductions"`
	EffectiveFrom  datatypes.Date        `gorm:"not null" json:"effective_from"`
	EffectiveUntil *datatypes.Date       `json:"effective_until,omitempty"`
	Status         SalaryStructureStatus `gorm:"size:20;default:ACTIVE" json:"status"`
}

type PayrollStatus string

const (
	PayrollDraft      PayrollStatus = "DRAFT"
	PayrollProcessing PayrollStatus = "PROCESSING"
	PayrollProcessed  PayrollStatus = "PROCESSED"
	PayrollPaid       PayrollStatus = "PAID"
	PayrollCancelled  PayrollStatus = "CANCELLED"
)

type Payroll struct {
	BaseModel
	TenantID          string        `gorm:"type:uuid;not null;index;default:00000000-0000-0000-0000-000000000001" json:"tenant_id"`
	Month             int           `gorm:"not null;uniqueIndex:idx_payroll_emp_month_year" json:"month"`
	Year              int           `gorm:"not null;uniqueIndex:idx_payroll_emp_month_year" json:"year"`
	EmployeeID        string        `gorm:"type:uuid;not null;index;uniqueIndex:idx_payroll_emp_month_year" json:"employee_id"`
	Employee          *Employee     `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	SalaryStructureID string        `gorm:"type:uuid;index" json:"salary_structure_id"`
	BasicSalary       Decimal       `gorm:"type:numeric(14,2);not null" json:"basic_salary"`
	Allowances        Decimal       `gorm:"type:numeric(14,2);default:0" json:"allowances"`
	Bonus             Decimal       `gorm:"type:numeric(14,2);default:0" json:"bonus"`
	Overtime          Decimal       `gorm:"type:numeric(14,2);default:0" json:"overtime"`
	GrossSalary       Decimal       `gorm:"type:numeric(14,2);default:0" json:"gross_salary"`
	Tax               Decimal       `gorm:"type:numeric(14,2);default:0" json:"tax"`
	Deductions        Decimal       `gorm:"type:numeric(14,2);default:0" json:"deductions"`
	NetSalary         Decimal       `gorm:"type:numeric(14,2);default:0" json:"net_salary"`
	Status            PayrollStatus `gorm:"size:20;index;default:DRAFT" json:"status"`
	ProcessedBy       *string       `gorm:"type:uuid;index" json:"processed_by,omitempty"`
	ProcessedAt       *time.Time    `json:"processed_at,omitempty"`
	PaidOn            *time.Time    `json:"paid_on,omitempty"`
	PaymentRef        string        `gorm:"size:150" json:"payment_ref"`
	Notes             string        `gorm:"size:1000" json:"notes"`
}
