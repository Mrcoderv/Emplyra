package handlers

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/emplyra/backend/internal/dto"
	"github.com/emplyra/backend/internal/middleware"
	"github.com/emplyra/backend/internal/models"
	"github.com/emplyra/backend/internal/responses"
	"github.com/emplyra/backend/internal/services"
	"github.com/emplyra/backend/internal/utils"
)

type AuthHandler struct {
	auth *services.AuthService
}

func NewAuthHandler(auth *services.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.Error(c, 400, "validation failed", utils.ConvertValidationErrors(err))
		return
	}
	user, pair, err := h.auth.Login(services.LoginInput{
		Identifier: req.Identifier,
		Password:   req.Password,
		IP:         clientIP(c),
		UserAgent:  userAgent(c),
	})
	if err != nil {
		if mapServiceError(c, err) {
			return
		}
		responses.Error(c, 500, "login failed", nil)
		return
	}
	responses.OK(c, "login successful", gin.H{
		"tokens": pair,
		"user":   toUserDTO(user),
	})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req dto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.Error(c, 400, "validation failed", utils.ConvertValidationErrors(err))
		return
	}
	user, pair, err := h.auth.Refresh(req.RefreshToken, clientIP(c), userAgent(c))
	if err != nil {
		if mapServiceError(c, err) {
			return
		}
		responses.Error(c, 500, "refresh failed", nil)
		return
	}
	responses.OK(c, "token refreshed", gin.H{"tokens": pair, "user": toUserDTO(user)})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var req dto.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.Error(c, 400, "validation failed", utils.ConvertValidationErrors(err))
		return
	}
	_ = h.auth.Logout(req.RefreshToken, clientIP(c), userAgent(c))
	responses.OK(c, "logout successful", nil)
}

func (h *AuthHandler) Me(c *gin.Context) {
	p := middleware.MustPrincipal(c)
	if p == nil {
		return
	}
	user, perms, err := h.auth.Me(p.UserID)
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "current user", dto.MeResponse{
		User:        toUserDTO(user),
		Permissions: perms,
	})
}

func toUserDTO(u *models.User) dto.UserDTO {
	d := dto.UserDTO{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Status:    string(u.Status),
		RoleID:    u.RoleID,
	}
	if u.Role != nil {
		d.Role = u.Role.Name
	}
	if u.LastLoginAt != nil {
		d.LastLogin = u.LastLoginAt.Format(time.RFC3339)
	}
	return d
}
