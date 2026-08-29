package handlers

import (
	"errors"

	"github.com/gin-gonic/gin"

	"github.com/emplyra/backend/internal/dto"
	"github.com/emplyra/backend/internal/middleware"
	"github.com/emplyra/backend/internal/models"
	"github.com/emplyra/backend/internal/repositories"
	"github.com/emplyra/backend/internal/responses"
	"github.com/emplyra/backend/internal/services"
	"github.com/emplyra/backend/internal/utils"
)

type RoleHandler struct {
	roles *services.RoleService
}

func NewRoleHandler(roles *services.RoleService) *RoleHandler {
	return &RoleHandler{roles: roles}
}

func (h *RoleHandler) List(c *gin.Context) {
	actor := middleware.MustPrincipal(c)
	roles, err := h.roles.List(actor.Scope)
	if err != nil {
		mapServiceError(c, err)
		return
	}
	items := make([]dto.RoleDTO, 0, len(roles))
	for i := range roles {
		items = append(items, toRoleDTO(&roles[i]))
	}
	responses.OK(c, "roles", items)
}

func (h *RoleHandler) Get(c *gin.Context) {
	actor := middleware.MustPrincipal(c)
	r, err := h.roles.Get(c.Param("id"), actor.Scope)
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "role", toRoleDTO(r))
}

func (h *RoleHandler) Create(c *gin.Context) {
	var req dto.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.Error(c, 400, "validation failed", utils.ConvertValidationErrors(err))
		return
	}
	actor := middleware.MustPrincipal(c)
	r, err := h.roles.Create(req.Name, req.Description, req.PermissionIDs, actor.Scope, actor.UserID, clientIP(c), userAgent(c))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.Created(c, "role created", toRoleDTO(r))
}

func (h *RoleHandler) Update(c *gin.Context) {
	var req dto.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.Error(c, 400, "validation failed", utils.ConvertValidationErrors(err))
		return
	}
	actor := middleware.MustPrincipal(c)
	r, err := h.roles.Update(c.Param("id"), req.Name, req.Description, req.PermissionIDs, actor.Scope, actor.UserID, clientIP(c), userAgent(c))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "role updated", toRoleDTO(r))
}

func (h *RoleHandler) Delete(c *gin.Context) {
	actor := middleware.MustPrincipal(c)
	err := h.roles.Delete(c.Param("id"), actor.UserID, clientIP(c), userAgent(c), actor.Scope)
	if err != nil {
		if errors.Is(err, repositories.ErrRoleInUse) {
			responses.Error(c, 409, "role is assigned to users", nil)
			return
		}
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "role deleted", nil)
}

func (h *RoleHandler) Permissions(c *gin.Context) {
	perms, err := h.roles.Permissions()
	if err != nil {
		mapServiceError(c, err)
		return
	}
	items := make([]dto.PermissionDTO, 0, len(perms))
	for _, p := range perms {
		items = append(items, dto.PermissionDTO{ID: p.ID, Name: p.Name, Description: p.Description, Module: p.Module})
	}
	responses.OK(c, "permissions", items)
}

func toRoleDTO(r *models.Role) dto.RoleDTO {
	out := dto.RoleDTO{ID: r.ID, Name: r.Name, Description: r.Description, IsSystem: r.IsSystem}
	for _, p := range r.Permissions {
		out.Permissions = append(out.Permissions, dto.PermissionDTO{ID: p.ID, Name: p.Name, Description: p.Description, Module: p.Module})
	}
	return out
}
