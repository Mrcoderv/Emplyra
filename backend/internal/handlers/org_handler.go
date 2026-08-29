package handlers

import (
	"github.com/gin-gonic/gin"

	"github.com/emplyra/backend/internal/dto"
	"github.com/emplyra/backend/internal/middleware"
	"github.com/emplyra/backend/internal/responses"
	"github.com/emplyra/backend/internal/services"
	"github.com/emplyra/backend/internal/utils"
)

type DepartmentHandler struct {
	svc *services.DepartmentService
}

func NewDepartmentHandler(svc *services.DepartmentService) *DepartmentHandler {
	return &DepartmentHandler{svc: svc}
}

func (h *DepartmentHandler) List(c *gin.Context) {
	items, err := h.svc.List(middleware.TenantID(c))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "departments", items)
}

func (h *DepartmentHandler) Get(c *gin.Context) {
	d, err := h.svc.Get(middleware.TenantID(c), c.Param("id"))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "department", d)
}

func (h *DepartmentHandler) Create(c *gin.Context) {
	var req dto.DepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.Error(c, 400, "validation failed", utils.ConvertValidationErrors(err))
		return
	}
	actor := middleware.MustPrincipal(c)
	d, err := h.svc.Create(middleware.TenantID(c), struct{ Name, Code, Description, ManagerID, Status string }{
		Name: sanitizeField(req.Name), Code: sanitizeField(req.Code), Description: req.Description, ManagerID: req.ManagerID, Status: req.Status,
	}, actor.UserID, clientIP(c), userAgent(c))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.Created(c, "department created", d)
}

func (h *DepartmentHandler) Update(c *gin.Context) {
	var req dto.DepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.Error(c, 400, "validation failed", utils.ConvertValidationErrors(err))
		return
	}
	actor := middleware.MustPrincipal(c)
	d, err := h.svc.Update(middleware.TenantID(c), c.Param("id"), struct{ Name, Code, Description, ManagerID, Status string }{
		Name: sanitizeField(req.Name), Code: sanitizeField(req.Code), Description: req.Description, ManagerID: req.ManagerID, Status: req.Status,
	}, actor.UserID, clientIP(c), userAgent(c))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "department updated", d)
}

func (h *DepartmentHandler) Delete(c *gin.Context) {
	actor := middleware.MustPrincipal(c)
	if err := h.svc.Delete(middleware.TenantID(c), c.Param("id"), actor.UserID, clientIP(c), userAgent(c)); err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "department deleted", nil)
}

type DesignationHandler struct {
	svc *services.DesignationService
}

func NewDesignationHandler(svc *services.DesignationService) *DesignationHandler {
	return &DesignationHandler{svc: svc}
}

func (h *DesignationHandler) List(c *gin.Context) {
	items, err := h.svc.List(middleware.TenantID(c))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "designations", items)
}

func (h *DesignationHandler) Get(c *gin.Context) {
	d, err := h.svc.Get(middleware.TenantID(c), c.Param("id"))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "designation", d)
}

func (h *DesignationHandler) Create(c *gin.Context) {
	var req dto.DesignationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.Error(c, 400, "validation failed", utils.ConvertValidationErrors(err))
		return
	}
	actor := middleware.MustPrincipal(c)
	d, err := h.svc.Create(middleware.TenantID(c), struct {
		Name, Description, DepartmentID, Status string
		Level                                   int
	}{
		Name: sanitizeField(req.Name), Description: req.Description, DepartmentID: req.DepartmentID, Status: req.Status, Level: req.Level,
	}, actor.UserID, clientIP(c), userAgent(c))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.Created(c, "designation created", d)
}

func (h *DesignationHandler) Update(c *gin.Context) {
	var req dto.DesignationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.Error(c, 400, "validation failed", utils.ConvertValidationErrors(err))
		return
	}
	actor := middleware.MustPrincipal(c)
	d, err := h.svc.Update(middleware.TenantID(c), c.Param("id"), struct {
		Name, Description, DepartmentID, Status string
		Level                                   int
	}{
		Name: sanitizeField(req.Name), Description: req.Description, DepartmentID: req.DepartmentID, Status: req.Status, Level: req.Level,
	}, actor.UserID, clientIP(c), userAgent(c))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "designation updated", d)
}

func (h *DesignationHandler) Delete(c *gin.Context) {
	actor := middleware.MustPrincipal(c)
	if err := h.svc.Delete(middleware.TenantID(c), c.Param("id"), actor.UserID, clientIP(c), userAgent(c)); err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "designation deleted", nil)
}
