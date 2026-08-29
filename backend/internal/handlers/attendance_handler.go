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

type AttendanceHandler struct {
	svc *services.AttendanceService
}

func NewAttendanceHandler(svc *services.AttendanceService) *AttendanceHandler {
	return &AttendanceHandler{svc: svc}
}

func (h *AttendanceHandler) CheckIn(c *gin.Context) {
	var req dto.CheckInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.Error(c, 400, "validation failed", utils.ConvertValidationErrors(err))
		return
	}
	employeeID, ok := h.resolveEmployee(c, req.EmployeeID)
	if !ok {
		return
	}
	actor := middleware.MustPrincipal(c)
	a, err := h.svc.CheckIn(middleware.TenantID(c), employeeID, req.Remarks, actor.UserID, clientIP(c), userAgent(c))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.Created(c, "check in recorded", a)
}

func (h *AttendanceHandler) CheckOut(c *gin.Context) {
	var req dto.CheckOutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.Error(c, 400, "validation failed", utils.ConvertValidationErrors(err))
		return
	}
	employeeID, ok := h.resolveEmployee(c, req.EmployeeID)
	if !ok {
		return
	}
	actor := middleware.MustPrincipal(c)
	a, err := h.svc.CheckOut(middleware.TenantID(c), employeeID, req.Remarks, req.Overtime, actor.UserID, clientIP(c), userAgent(c))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "check out recorded", a)
}

func (h *AttendanceHandler) List(c *gin.Context) {
	var params dto.AttendanceListParams
	_ = c.ShouldBindQuery(&params)
	employeeID := params.EmployeeID
	p := middleware.GetPrincipal(c)
	if !h.canReadAll(p) {
		var err error
		employeeID, err = h.svc.EmployeeIDForUser(p.UserID)
		if err != nil {
			responses.Error(c, 404, err.Error(), nil)
			return
		}
	}
	pg := utils.NewPagination(params.Page, params.PageSize)
	items, total, err := h.svc.List(middleware.TenantID(c), pg, employeeID, params.From, params.To, params.Status)
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "attendance", responses.List{Items: items, Total: total, Page: pg.Page, PageSize: pg.PageSize, TotalPages: pg.TotalPages(total)})
}

func (h *AttendanceHandler) Get(c *gin.Context) {
	a, err := h.svc.Get(middleware.TenantID(c), c.Param("id"))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	if !h.canReadAll(middleware.GetPrincipal(c)) {
		p := middleware.GetPrincipal(c)
		myID, err := h.svc.EmployeeIDForUser(p.UserID)
		if err != nil || a.EmployeeID != myID {
			responses.Error(c, 403, "forbidden", nil)
			return
		}
	}
	responses.OK(c, "attendance", a)
}

func (h *AttendanceHandler) Update(c *gin.Context) {
	var req dto.UpdateAttendanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.Error(c, 400, "validation failed", utils.ConvertValidationErrors(err))
		return
	}
	actor := middleware.MustPrincipal(c)
	a, err := h.svc.Update(middleware.TenantID(c), c.Param("id"), struct {
		CheckOut    *string
		CheckIn     *string
		Status      string
		LateMinutes int
		Overtime    float64
		Remarks     string
	}{CheckOut: req.CheckOut, CheckIn: req.CheckIn, Status: req.Status, LateMinutes: req.LateMinutes, Overtime: req.Overtime, Remarks: req.Remarks}, actor.UserID, clientIP(c), userAgent(c))
	if err != nil {
		if err == services.ErrNotFound {
			mapServiceError(c, err)
			return
		}
		responses.Error(c, 400, err.Error(), nil)
		return
	}
	responses.OK(c, "attendance updated", a)
}

func (h *AttendanceHandler) resolveEmployee(c *gin.Context, requested string) (string, bool) {
	p := middleware.MustPrincipal(c)
	if requested != "" {
		if !h.canReadAll(p) {
			responses.Error(c, 403, "forbidden", nil)
			return "", false
		}
		return requested, true
	}
	empID, err := h.svc.EmployeeIDForUser(p.UserID)
	if err != nil {
		responses.Error(c, 404, err.Error(), nil)
		return "", false
	}
	return empID, true
}

func (h *AttendanceHandler) canReadAll(p *middleware.Principal) bool {
	return p != nil && p.Role != string(models.RoleEmployee)
}
