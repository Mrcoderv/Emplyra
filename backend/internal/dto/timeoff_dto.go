package dto

type CheckInRequest struct {
	EmployeeID string `json:"employee_id"`
	Remarks    string `json:"remarks" binding:"max=500"`
}

type CheckOutRequest struct {
	EmployeeID string  `json:"employee_id"`
	Remarks    string  `json:"remarks" binding:"max=500"`
	Overtime   float64 `json:"overtime"`
}

type AttendanceListParams struct {
	EmployeeID string `form:"employee_id"`
	From       string `form:"from"`
	To         string `form:"to"`
	Status     string `form:"status"`
	Page       string `form:"page"`
	PageSize   string `form:"page_size"`
}

type UpdateAttendanceRequest struct {
	CheckIn     *string `json:"check_in"`
	CheckOut    *string `json:"check_out"`
	Status      string  `json:"status"`
	LateMinutes int     `json:"late_minutes"`
	Overtime    float64 `json:"overtime"`
	Remarks     string  `json:"remarks" binding:"max=500"`
}

type CreateLeaveRequest struct {
	EmployeeID  string `json:"employee_id"`
	LeaveTypeID string `json:"leave_type_id" binding:"required"`
	StartDate   string `json:"start_date" binding:"required"`
	EndDate     string `json:"end_date" binding:"required"`
	Reason      string `json:"reason" binding:"required,max=1000"`
}

type LeaveDecisionRequest struct {
	Note string `json:"note" binding:"max=1000"`
}

type LeaveListParams struct {
	EmployeeID string `form:"employee_id"`
	Type       string `form:"leave_type_id"`
	Status     string `form:"status"`
	From       string `form:"from"`
	To         string `form:"to"`
	Page       string `form:"page"`
	PageSize   string `form:"page_size"`
}

type HolidayRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=150"`
	Date        string `json:"date" binding:"required"`
	Description string `json:"description" binding:"max=500"`
	Type        string `json:"type" binding:"max=50"`
	Status      string `json:"status"`
}

type LeaveBalanceUpdate struct {
	Entitlement int `json:"entitlement"`
}
