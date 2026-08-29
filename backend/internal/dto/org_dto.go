package dto

type DepartmentRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=150"`
	Code        string `json:"code" binding:"required,min=2,max=50"`
	Description string `json:"description" binding:"max=500"`
	ManagerID   string `json:"manager_id"`
	Status      string `json:"status"`
}

type DesignationRequest struct {
	Name         string `json:"name" binding:"required,min=2,max=150"`
	Description  string `json:"description" binding:"max=500"`
	DepartmentID string `json:"department_id"`
	Level        int    `json:"level"`
	Status       string `json:"status"`
}

type EmployeeRequest struct {
	EmployeeCode     string `json:"employee_code" binding:"required,min=2,max=50"`
	FirstName        string `json:"first_name" binding:"required,min=1,max=100"`
	LastName         string `json:"last_name" binding:"max=100"`
	Email            string `json:"email" binding:"required,email"`
	Phone            string `json:"phone" binding:"max=30"`
	DateOfBirth      string `json:"date_of_birth"`
	Gender           string `json:"gender"`
	Address          string `json:"address" binding:"max=500"`
	EmergencyContact string `json:"emergency_contact" binding:"max=255"`
	JoiningDate      string `json:"joining_date"`
	EmploymentType   string `json:"employment_type"`
	Status           string `json:"status"`
	DepartmentID     string `json:"department_id"`
	DesignationID    string `json:"designation_id"`
	ManagerID        string `json:"manager_id"`
	UserID           string `json:"user_id"`
}

type EmployeeUpdateRequest struct {
	EmployeeCode     string `json:"employee_code" binding:"max=50"`
	FirstName        string `json:"first_name" binding:"max=100"`
	LastName         string `json:"last_name" binding:"max=100"`
	Email            string `json:"email" binding:"email"`
	Phone            string `json:"phone" binding:"max=30"`
	DateOfBirth      string `json:"date_of_birth"`
	Gender           string `json:"gender"`
	Address          string `json:"address" binding:"max=500"`
	EmergencyContact string `json:"emergency_contact" binding:"max=255"`
	JoiningDate      string `json:"joining_date"`
	EmploymentType   string `json:"employment_type"`
	Status           string `json:"status"`
	DepartmentID     string `json:"department_id"`
	DesignationID    string `json:"designation_id"`
	ManagerID        string `json:"manager_id"`
	UserID           string `json:"user_id"`
}

type EmployeeListParams struct {
	Search         string `form:"search"`
	DepartmentID   string `form:"department_id"`
	DesignationID  string `form:"designation_id"`
	ManagerID      string `form:"manager_id"`
	Status         string `form:"status"`
	EmploymentType string `form:"employment_type"`
	Page           string `form:"page"`
	PageSize       string `form:"page_size"`
	SortBy         string `form:"sort_by"`
	SortOrder      string `form:"sort_order"`
}
