package models

import (
	"time"

	"gorm.io/datatypes"
)

type AttendanceStatus string

const (
	AttendancePresent AttendanceStatus = "PRESENT"
	AttendanceAbsent  AttendanceStatus = "ABSENT"
	AttendanceHalfDay AttendanceStatus = "HALF_DAY"
	AttendanceLate    AttendanceStatus = "LATE"
	AttendanceOnLeave AttendanceStatus = "ON_LEAVE"
	AttendanceHoliday AttendanceStatus = "HOLIDAY"
)

type Attendance struct {
	BaseModel
	EmployeeID   string           `gorm:"type:uuid;not null;index;uniqueIndex:idx_attendance_emp_date" json:"employee_id"`
	Employee     *Employee        `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	Date         datatypes.Date   `gorm:"not null;uniqueIndex:idx_attendance_emp_date" json:"date"`
	CheckIn      *time.Time       `json:"check_in,omitempty"`
	CheckOut     *time.Time       `json:"check_out,omitempty"`
	WorkingHours float64          `json:"working_hours"`
	Status       AttendanceStatus `gorm:"size:20;index;default:PRESENT" json:"status"`
	LateMinutes  int              `json:"late_minutes"`
	Overtime     float64          `json:"overtime"`
	Remarks      string           `gorm:"size:500" json:"remarks"`
}

type LeaveType struct {
	BaseModel
	Name        string `gorm:"size:100;not null" json:"name"`
	Code        string `gorm:"size:50;uniqueIndex;not null" json:"code"`
	Description string `gorm:"size:500" json:"description"`
	IsPaid      bool   `gorm:"default:true" json:"is_paid"`
}

type LeaveStatus string

const (
	LeavePending   LeaveStatus = "PENDING"
	LeaveApproved  LeaveStatus = "APPROVED"
	LeaveRejected  LeaveStatus = "REJECTED"
	LeaveCancelled LeaveStatus = "CANCELLED"
)

type Leave struct {
	BaseModel
	EmployeeID  string         `gorm:"type:uuid;not null;index" json:"employee_id"`
	Employee    *Employee      `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	LeaveTypeID string         `gorm:"type:uuid;not null;index" json:"leave_type_id"`
	LeaveType   *LeaveType     `gorm:"foreignKey:LeaveTypeID" json:"leave_type,omitempty"`
	StartDate   datatypes.Date `gorm:"not null" json:"start_date"`
	EndDate     datatypes.Date `gorm:"not null" json:"end_date"`
	Days        int            `json:"days"`
	Reason      string         `gorm:"size:1000" json:"reason"`
	Status      LeaveStatus    `gorm:"size:20;index;default:PENDING" json:"status"`
	ReviewerID  *string        `gorm:"type:uuid;index" json:"reviewer_id"`
	Reviewer    *Employee      `gorm:"foreignKey:ReviewerID" json:"reviewer,omitempty"`
	ReviewedAt  *time.Time     `json:"reviewed_at,omitempty"`
	ReviewNote  string         `gorm:"size:1000" json:"review_note"`
}

type LeaveBalance struct {
	BaseModel
	EmployeeID  string `gorm:"type:uuid;not null;index" json:"employee_id"`
	LeaveTypeID string `gorm:"type:uuid;not null;index" json:"leave_type_id"`
	Year        int    `gorm:"index;not null" json:"year"`
	Entitlement int    `json:"entitlement"`
	Used        int    `json:"used"`
}

type HolidayStatus string

const (
	HolidayActive   HolidayStatus = "ACTIVE"
	HolidayInactive HolidayStatus = "INACTIVE"
)

type Holiday struct {
	BaseModel
	Name        string         `gorm:"size:150;not null" json:"name"`
	Date        datatypes.Date `gorm:"not null;index" json:"date"`
	Description string         `gorm:"size:500" json:"description"`
	Type        string         `gorm:"size:50" json:"type"`
	Status      HolidayStatus  `gorm:"size:20;default:ACTIVE" json:"status"`
}
