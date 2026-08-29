package dto

type LoginRequest struct {
	Identifier string `json:"identifier" binding:"required"`
	Password   string `json:"password" binding:"required,min=6"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type UserDTO struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Status    string `json:"status"`
	RoleID    string `json:"role_id"`
	Role      string `json:"role"`
	LastLogin string `json:"last_login"`
}

type MeTenant struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
	Plan   string `json:"plan"`
}

type MeResponse struct {
	User        UserDTO   `json:"user"`
	Permissions []string  `json:"permissions"`
	Roles       []string  `json:"roles"`
	Scope       string    `json:"scope"`
	Tenant      *MeTenant `json:"tenant,omitempty"`
}

type CreateUserRequest struct {
	Username  string `json:"username" binding:"required,min=3,max=64"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"first_name" binding:"max=100"`
	LastName  string `json:"last_name" binding:"max=100"`
	RoleID    string `json:"role_id" binding:"required"`
	Status    string `json:"status"`
}

type UpdateUserRequest struct {
	FirstName string `json:"first_name" binding:"max=100"`
	LastName  string `json:"last_name" binding:"max=100"`
	RoleID    string `json:"role_id"`
	Status    string `json:"status"`
}

type UpdatePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}
