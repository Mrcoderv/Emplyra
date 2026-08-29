package models

import "time"

type AuditAction string

const (
	ActionLogin            AuditAction = "LOGIN"
	ActionLogout           AuditAction = "LOGOUT"
	ActionLoginFailed      AuditAction = "LOGIN_FAILED"
	ActionCreate           AuditAction = "CREATE"
	ActionUpdate           AuditAction = "UPDATE"
	ActionDelete           AuditAction = "DELETE"
	ActionApprove          AuditAction = "APPROVE"
	ActionReject           AuditAction = "REJECT"
	ActionPayrollProcess   AuditAction = "PAYROLL_PROCESS"
	ActionRoleChange       AuditAction = "ROLE_CHANGE"
	ActionPermissionChange AuditAction = "PERMISSION_CHANGE"
	ActionOther            AuditAction = "OTHER"
)

type AuditLog struct {
	BaseModel
	UserID     *string     `gorm:"type:uuid;index" json:"user_id,omitempty"`
	User       *User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Action     AuditAction `gorm:"size:40;index;not null" json:"action"`
	Resource   string      `gorm:"size:100;index" json:"resource"`
	ResourceID string      `gorm:"size:64;index" json:"resource_id"`
	IPAddress  string      `gorm:"size:64" json:"ip_address"`
	UserAgent  string      `gorm:"size:255" json:"user_agent"`
	Metadata   string      `gorm:"type:jsonb" json:"metadata"`
	CreatedAt  time.Time   `json:"created_at"`
}

type NotificationType string

const (
	NotifyLeaveApproval  NotificationType = "LEAVE_APPROVAL"
	NotifyLeaveRejection NotificationType = "LEAVE_REJECTION"
	NotifyLeaveRequest   NotificationType = "LEAVE_REQUEST"
	NotifyPayroll        NotificationType = "PAYROLL"
	NotifyAttendance     NotificationType = "ATTENDANCE"
	NotifyRecruitment    NotificationType = "RECRUITMENT"
	NotifyTraining       NotificationType = "TRAINING"
	NotifyAnnouncement   NotificationType = "ANNOUNCEMENT"
	NotifySystem         NotificationType = "SYSTEM"
)

type Notification struct {
	BaseModel
	UserID   string           `gorm:"type:uuid;index;not null" json:"user_id"`
	User     *User            `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Title    string           `gorm:"size:200;not null" json:"title"`
	Message  string           `gorm:"size:2000" json:"message"`
	Type     NotificationType `gorm:"size:40;index" json:"type"`
	IsRead   bool             `gorm:"index;default:false" json:"is_read"`
	ReadAt   *time.Time       `json:"read_at,omitempty"`
	Link     string           `gorm:"size:255" json:"link"`
	Metadata string           `gorm:"type:jsonb" json:"metadata"`
}
