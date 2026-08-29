package dto

type SalaryStructureRequest struct {
	EmployeeID    string `json:"employee_id" binding:"required"`
	BasicSalary   string `json:"basic_salary" binding:"required"`
	Allowances    string `json:"allowances"`
	Bonus         string `json:"bonus"`
	OvertimeRate  string `json:"overtime_rate"`
	TaxRate       string `json:"tax_rate"`
	TaxAmount     string `json:"tax_amount"`
	Deductions    string `json:"deductions"`
	EffectiveFrom string `json:"effective_from" binding:"required"`
	Status        string `json:"status"`
}

type SalaryStructureUpdateRequest struct {
	BasicSalary    string `json:"basic_salary"`
	Allowances     string `json:"allowances"`
	Bonus          string `json:"bonus"`
	OvertimeRate   string `json:"overtime_rate"`
	TaxRate        string `json:"tax_rate"`
	TaxAmount      string `json:"tax_amount"`
	Deductions     string `json:"deductions"`
	EffectiveFrom  string `json:"effective_from"`
	EffectiveUntil string `json:"effective_until"`
	Status         string `json:"status"`
}

type GeneratePayrollRequest struct {
	Month int `json:"month" binding:"required,min=1,max=12"`
	Year  int `json:"year" binding:"required,min=2000"`
}

type ProcessPayrollRequest struct {
	Bonus      string `json:"bonus"`
	Overtime   string `json:"overtime"`
	Deductions string `json:"deductions"`
	Notes      string `json:"notes" binding:"max=1000"`
}

type MarkPaidRequest struct {
	PaymentRef string `json:"payment_ref" binding:"max=150"`
	Notes      string `json:"notes" binding:"max=1000"`
}

type CancelPayrollRequest struct {
	Notes string `json:"notes" binding:"max=1000"`
}

type PayrollListParams struct {
	Month        string `form:"month"`
	Year         string `form:"year"`
	EmployeeID   string `form:"employee_id"`
	Status       string `form:"status"`
	DepartmentID string `form:"department_id"`
	Page         string `form:"page"`
	PageSize     string `form:"page_size"`
}
