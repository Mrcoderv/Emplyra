package handlers

import (
	"github.com/gin-gonic/gin"

	"github.com/emplyra/backend/internal/dto"
	"github.com/emplyra/backend/internal/middleware"
	"github.com/emplyra/backend/internal/responses"
	"github.com/emplyra/backend/internal/services"
	"github.com/emplyra/backend/internal/utils"
)

type UserHandler struct {
	users *services.UserService
}

func NewUserHandler(users *services.UserService) *UserHandler {
	return &UserHandler{users: users}
}

func (h *UserHandler) List(c *gin.Context) {
	p := utils.NewPagination(c.Query("page"), c.Query("page_size"))
	users, total, err := h.users.List(p, c.Query("search"))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	items := make([]dto.UserDTO, 0, len(users))
	for i := range users {
		items = append(items, toUserDTO(&users[i]))
	}
	responses.OK(c, "users", responses.List{Items: items, Total: total, Page: p.Page, PageSize: p.PageSize, TotalPages: p.TotalPages(total)})
}

func (h *UserHandler) Get(c *gin.Context) {
	u, err := h.users.GetByID(c.Param("id"))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "user", toUserDTO(u))
}

func (h *UserHandler) Create(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.Error(c, 400, "validation failed", utils.ConvertValidationErrors(err))
		return
	}
	actor := middleware.MustPrincipal(c)

	u, err := h.users.Create(struct {
		Username  string
		Email     string
		Password  string
		FirstName string
		LastName  string
		RoleID    string
		Status    string
	}{
		Username:  sanitizeField(req.Username),
		Email:     sanitizeField(req.Email),
		Password:  req.Password,
		FirstName: sanitizeField(req.FirstName),
		LastName:  sanitizeField(req.LastName),
		RoleID:    req.RoleID,
		Status:    req.Status,
	}, actor.UserID, clientIP(c), userAgent(c))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.Created(c, "user created", toUserDTO(u))
}

func (h *UserHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.Error(c, 400, "validation failed", utils.ConvertValidationErrors(err))
		return
	}
	actor := middleware.MustPrincipal(c)
	u, err := h.users.Update(id, struct {
		FirstName string
		LastName  string
		RoleID    string
		Status    string
	}{
		FirstName: sanitizeField(req.FirstName),
		LastName:  sanitizeField(req.LastName),
		RoleID:    req.RoleID,
		Status:    req.Status,
	}, actor.UserID, clientIP(c), userAgent(c))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "user updated", toUserDTO(u))
}

func (h *UserHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	actor := middleware.MustPrincipal(c)
	if err := h.users.Delete(id, actor.UserID, clientIP(c), userAgent(c)); err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "user deleted", nil)
}

func (h *UserHandler) ChangePassword(c *gin.Context) {
	var req dto.UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.Error(c, 400, "validation failed", utils.ConvertValidationErrors(err))
		return
	}
	actor := middleware.MustPrincipal(c)
	if err := h.users.ChangePassword(actor.UserID, req.CurrentPassword, req.NewPassword, clientIP(c), userAgent(c)); err != nil {
		if err.Error() == "current password is incorrect" {
			responses.Error(c, 400, err.Error(), nil)
			return
		}
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "password changed", nil)
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	var req struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.Error(c, 400, "validation failed", nil)
		return
	}
	actor := middleware.MustPrincipal(c)
	u, err := h.users.Update(actor.UserID, struct {
		FirstName string
		LastName  string
		RoleID    string
		Status    string
	}{FirstName: sanitizeField(req.FirstName), LastName: sanitizeField(req.LastName)}, actor.UserID, clientIP(c), userAgent(c))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "profile updated", toUserDTO(u))
}
