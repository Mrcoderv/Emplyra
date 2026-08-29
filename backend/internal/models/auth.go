package models

import "time"

type RoleName string

const (
	RoleSuperAdmin RoleName = "SUPER_ADMIN"
	RoleHRAdmin    RoleName = "HR_ADMIN"
	RoleManager    RoleName = "MANAGER"
	RoleEmployee   RoleName = "EMPLOYEE"
	RoleRecruiter  RoleName = "RECRUITER"
	RoleAccountant RoleName = "ACCOUNTANT"
)

type UserStatus string

const (
	UserStatusActive    UserStatus = "ACTIVE"
	UserStatusInactive  UserStatus = "INACTIVE"
	UserStatusSuspended UserStatus = "SUSPENDED"
)

type User struct {
	BaseModel
	Username     string     `gorm:"size:64;uniqueIndex;not null" json:"username"`
	Email        string     `gorm:"size:255;uniqueIndex;not null" json:"email"`
	PasswordHash string     `gorm:"size:255;not null" json:"-"`
	FirstName    string     `gorm:"size:100" json:"first_name"`
	LastName     string     `gorm:"size:100" json:"last_name"`
	Status       UserStatus `gorm:"size:20;index;default:ACTIVE" json:"status"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
	RoleID       string     `gorm:"type:uuid;index" json:"role_id"`
	Role         *Role      `gorm:"foreignKey:RoleID" json:"role,omitempty"`
}

func (u User) FullName() string {
	if u.FirstName == "" && u.LastName == "" {
		return u.Username
	}
	if u.FirstName == "" {
		return u.LastName
	}
	if u.LastName == "" {
		return u.FirstName
	}
	return u.FirstName + " " + u.LastName
}

type Role struct {
	BaseModel
	Name        string       `gorm:"size:50;uniqueIndex;not null" json:"name"`
	Description string       `gorm:"size:255" json:"description"`
	IsSystem    bool         `gorm:"default:false" json:"is_system"`
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
}

type Permission struct {
	BaseModel
	Name        string `gorm:"size:100;uniqueIndex;not null" json:"name"`
	Description string `gorm:"size:255" json:"description"`
	Module      string `gorm:"size:50;index" json:"module"`
}

type RefreshToken struct {
	BaseModel
	UserID     string     `gorm:"type:uuid;index;not null" json:"-"`
	TokenHash  string     `gorm:"size:64;uniqueIndex;not null" json:"-"`
	ExpiresAt  time.Time  `gorm:"index;not null" json:"expires_at"`
	RevokedAt  *time.Time `json:"-"`
	ReplacedBy string     `gorm:"type:uuid" json:"-"`
	IP         string     `gorm:"size:64" json:"-"`
	UserAgent  string     `gorm:"size:255" json:"-"`
	User       *User      `gorm:"foreignKey:UserID" json:"-"`
}
