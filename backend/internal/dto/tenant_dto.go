package dto

import (
	"strings"
	"time"

	"github.com/emplyra/backend/internal/models"
)

type CreateTenantRequest struct {
	Name           string `json:"name" binding:"required,min=2,max=200"`
	Slug           string `json:"slug" binding:"required,min=2,max=120"`
	Plan           string `json:"plan"`
	TrialDays      int    `json:"trial_days"`
	Industry       string `json:"industry"`
	OwnerEmail     string `json:"owner_email" binding:"omitempty,email"`
	OwnerPassword  string `json:"owner_password" binding:"omitempty,min=8"`
	OwnerFirstName string `json:"owner_first_name"`
	OwnerLastName  string `json:"owner_last_name"`
	OwnerUsername  string `json:"owner_username"`
}

type UpdateTenantRequest struct {
	Name     string `json:"name" binding:"omitempty,min=2,max=200"`
	Plan     string `json:"plan"`
	Industry string `json:"industry"`
}

type RegisterTenantRequest struct {
	Name      string `json:"name" binding:"required,min=2,max=200"`
	TrialDays int    `json:"trial_days"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"first_name" binding:"max=100"`
	LastName  string `json:"last_name" binding:"max=100"`
}

type CreateTenantOwnerRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type PlatformUserRequest struct {
	Username  string `json:"username" binding:"required,min=3,max=64"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Role      string `json:"role" binding:"required"`
	Status    string `json:"status"`
}

type TenantDTO struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Slug        string             `json:"slug"`
	Status      string             `json:"status"`
	Plan        string             `json:"plan"`
	TrialEndsAt *time.Time         `json:"trial_ends_at,omitempty"`
	Industry    string             `json:"industry"`
	CreatedAt   time.Time          `json:"created_at"`
	Usage       models.TenantUsage `json:"usage,omitempty"`
}

func ToTenantDTO(t *models.Tenant, usage models.TenantUsage, withUsage bool) TenantDTO {
	out := TenantDTO{
		ID:          t.ID,
		Name:        t.Name,
		Slug:        t.Slug,
		Status:      string(t.Status),
		Plan:        string(t.Plan),
		TrialEndsAt: t.TrialEndsAt,
		Industry:    t.Industry,
		CreatedAt:   t.CreatedAt,
	}
	if withUsage {
		out.Usage = usage
	}
	return out
}

func Slugify(s string) string {
	out := strings.Builder{}
	var prevDash bool
	for _, r := range strings.ToLower(strings.TrimSpace(s)) {
		ok := r >= 'a' && r <= 'z' || r >= '0' && r <= '9'
		if ok {
			out.WriteRune(r)
			prevDash = false
			continue
		}
		if !prevDash {
			out.WriteRune('-')
			prevDash = true
		}
	}
	slug := strings.Trim(out.String(), "-")
	if len(slug) > 120 {
		slug = slug[:120]
	}
	return slug
}
