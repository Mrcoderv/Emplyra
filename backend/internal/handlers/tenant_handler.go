package handlers

import (
	"github.com/gin-gonic/gin"

	"github.com/emplyra/backend/internal/dto"
	"github.com/emplyra/backend/internal/middleware"
	"github.com/emplyra/backend/internal/models"
	"github.com/emplyra/backend/internal/responses"
	"github.com/emplyra/backend/internal/services"
	"github.com/emplyra/backend/internal/utils"
)

type TenantHandler struct {
	tenants *services.TenantService
}

func NewTenantHandler(tenants *services.TenantService) *TenantHandler {
	return &TenantHandler{tenants: tenants}
}

func (h *TenantHandler) List(c *gin.Context) {
	p := utils.NewPagination(c.Query("page"), c.Query("page_size"))
	items, total, err := h.tenants.List(p, c.Query("search"), c.Query("status"))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	dtos := make([]dto.TenantDTO, 0, len(items))
	for i := range items {
		dtos = append(dtos, dto.ToTenantDTO(&items[i], models.TenantUsage{}, false))
	}
	responses.OK(c, "tenants", responses.List{Items: dtos, Total: total, Page: p.Page, PageSize: p.PageSize, TotalPages: p.TotalPages(total)})
}

func (h *TenantHandler) Get(c *gin.Context) {
	t, usage, err := h.tenants.Get(c.Param("id"))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "tenant", dto.ToTenantDTO(t, usage, true))
}

func (h *TenantHandler) Create(c *gin.Context) {
	var req dto.CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.Error(c, 400, "validation failed", utils.ConvertValidationErrors(err))
		return
	}
	actor := middleware.MustPrincipal(c)
	t, err := h.tenants.Create(struct {
		Name, Slug, Plan, Industry                                              string
		TrialDays                                                               int
		OwnerEmail, OwnerPassword, OwnerFirstName, OwnerLastName, OwnerUsername string
	}{
		Name: sanitizeField(req.Name), Slug: req.Slug, Plan: req.Plan, Industry: sanitizeField(req.Industry),
		TrialDays:  req.TrialDays,
		OwnerEmail: req.OwnerEmail, OwnerPassword: req.OwnerPassword,
		OwnerFirstName: sanitizeField(req.OwnerFirstName), OwnerLastName: sanitizeField(req.OwnerLastName),
		OwnerUsername: sanitizeField(req.OwnerUsername),
	}, actor.UserID, clientIP(c), userAgent(c))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.Created(c, "tenant created", dto.ToTenantDTO(t, models.TenantUsage{}, true))
}

func (h *TenantHandler) Update(c *gin.Context) {
	var req dto.UpdateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.Error(c, 400, "validation failed", utils.ConvertValidationErrors(err))
		return
	}
	actor := middleware.MustPrincipal(c)
	t, err := h.tenants.Update(c.Param("id"), struct {
		Name, Plan, Industry string
	}{Name: sanitizeField(req.Name), Plan: req.Plan, Industry: sanitizeField(req.Industry)}, actor.UserID, clientIP(c), userAgent(c))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "tenant updated", dto.ToTenantDTO(t, models.TenantUsage{}, true))
}

func (h *TenantHandler) Activate(c *gin.Context) {
	actor := middleware.MustPrincipal(c)
	t, err := h.tenants.Activate(c.Param("id"), actor.UserID, clientIP(c), userAgent(c))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "tenant activated", dto.ToTenantDTO(t, models.TenantUsage{}, true))
}

func (h *TenantHandler) Suspend(c *gin.Context) {
	actor := middleware.MustPrincipal(c)
	t, err := h.tenants.Suspend(c.Param("id"), actor.UserID, clientIP(c), userAgent(c))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "tenant suspended", dto.ToTenantDTO(t, models.TenantUsage{}, true))
}

func (h *TenantHandler) CreateOwner(c *gin.Context) {
	var req dto.CreateTenantOwnerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.Error(c, 400, "validation failed", utils.ConvertValidationErrors(err))
		return
	}
	actor := middleware.MustPrincipal(c)
	u, err := h.tenants.CreateOwner(c.Param("id"), struct {
		Email, Password, FirstName, LastName string
	}{Email: req.Email, Password: req.Password, FirstName: sanitizeField(req.FirstName), LastName: sanitizeField(req.LastName)}, actor.UserID, clientIP(c), userAgent(c))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.Created(c, "tenant owner created", toUserDTO(u))
}

func (h *TenantHandler) PlatformUsers(c *gin.Context) {
	p := utils.NewPagination(c.Query("page"), c.Query("page_size"))
	users, total, err := h.tenants.PlatformUsersList(p, c.Query("search"))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	items := make([]dto.UserDTO, 0, len(users))
	for i := range users {
		items = append(items, toUserDTO(&users[i]))
	}
	responses.OK(c, "platform users", responses.List{Items: items, Total: total, Page: p.Page, PageSize: p.PageSize, TotalPages: p.TotalPages(total)})
}

func (h *TenantHandler) CreatePlatformUser(c *gin.Context) {
	var req dto.PlatformUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.Error(c, 400, "validation failed", utils.ConvertValidationErrors(err))
		return
	}
	actor := middleware.MustPrincipal(c)
	u, err := h.tenants.CreatePlatformUser(struct {
		Username, Email, Password, FirstName, LastName, Role, Status string
	}{
		Username: sanitizeField(req.Username), Email: req.Email, Password: req.Password,
		FirstName: sanitizeField(req.FirstName), LastName: sanitizeField(req.LastName),
		Role: req.Role, Status: req.Status,
	}, actor.UserID, clientIP(c), userAgent(c))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.Created(c, "platform user created", toUserDTO(u))
}

func (h *TenantHandler) UpdatePlatformUser(c *gin.Context) {
	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.Error(c, 400, "validation failed", utils.ConvertValidationErrors(err))
		return
	}
	actor := middleware.MustPrincipal(c)
	u, err := h.tenants.UpdatePlatformUser(c.Param("id"), struct {
		FirstName, LastName, Role, Status string
	}{FirstName: sanitizeField(req.FirstName), LastName: sanitizeField(req.LastName), Role: req.RoleID, Status: req.Status}, actor.UserID, clientIP(c), userAgent(c))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "platform user updated", toUserDTO(u))
}

func (h *TenantHandler) DeletePlatformUser(c *gin.Context) {
	actor := middleware.MustPrincipal(c)
	if err := h.tenants.DeletePlatformUser(c.Param("id"), actor.UserID, clientIP(c), userAgent(c)); err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "platform user deleted", nil)
}

func (h *TenantHandler) Register(c *gin.Context) {
	var req dto.RegisterTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.Error(c, 400, "validation failed", utils.ConvertValidationErrors(err))
		return
	}
	t, err := h.tenants.Register(struct {
		Name, Email, Password, FirstName, LastName string
		TrialDays                                  int
	}{
		Name: sanitizeField(req.Name), Email: req.Email, Password: req.Password,
		FirstName: sanitizeField(req.FirstName), LastName: sanitizeField(req.LastName),
		TrialDays: req.TrialDays,
	}, clientIP(c), userAgent(c))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.Created(c, "tenant registered", dto.ToTenantDTO(t, models.TenantUsage{}, false))
}

func (h *TenantHandler) Dashboard(c *gin.Context) {
	byStatus, totalTenants, totalUsers, totalEmployees, err := h.tenants.Dashboard()
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "platform dashboard", gin.H{
		"organizations":           totalTenants,
		"active_organizations":    byStatus[models.TenantActive],
		"trial_organizations":     byStatus[models.TenantTrial],
		"suspended_organizations": byStatus[models.TenantSuspended],
		"total_users":             totalUsers,
		"total_employees":         totalEmployees,
		"system_usage":            gin.H{"storage": "unavailable", "api_calls": "unavailable"},
	})
}
