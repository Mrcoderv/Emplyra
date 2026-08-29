package dto

type PermissionDTO struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Module      string `json:"module"`
}

type RoleDTO struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	IsSystem    bool            `json:"is_system"`
	Permissions []PermissionDTO `json:"permissions"`
}

type CreateRoleRequest struct {
	Name          string   `json:"name" binding:"required,min=2,max=50"`
	Description   string   `json:"description" binding:"max=255"`
	PermissionIDs []string `json:"permission_ids" binding:"required"`
}

type UpdateRoleRequest struct {
	Name          string   `json:"name" binding:"required,min=2,max=50"`
	Description   string   `json:"description" binding:"max=255"`
	PermissionIDs []string `json:"permission_ids" binding:"required"`
}

type AssignRoleRequest struct {
	UserID string `json:"user_id" binding:"required"`
	RoleID string `json:"role_id" binding:"required"`
}
