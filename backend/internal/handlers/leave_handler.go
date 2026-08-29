package handlers

import (
	"github.com/gin-gonic/gin"

	"github.com/emplyra/backend/internal/dto"
	"github.com/emplyra/backend/internal/middleware"
	"github.com/emplyra/backend/internal/models"
	"github.com/emplyra/backend/internal/repositories"
	"github.com/emplyra/backend/internal/responses"
	"github.com/emplyra/backend/internal/services"
	"github.com/emplyra/backend/internal/utils"
	"strconv"
)

type LeaveHandler struct {
	svc *services.LeaveService
}

func NewLeaveHandler(svc *services.LeaveService) *LeaveHandler {
	return &LeaveHandler{svc: svc}
}

func (h *LeaveHandler) Create(c *gin.Context) {
	var req dto.CreateLeaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.Error(c, 400, "validation failed", utils.ConvertValidationErrors(err))
		return
	}
	p := middleware.MustPrincipal(c)
	employeeID := req.EmployeeID
	if employeeID == "" {
		empID, err := h.empIDForUser(c)
		if err != nil {
			responses.Error(c, 404, err.Error(), nil)
			return
		}
		employeeID = empID
	} else if p.Role == string(models.RoleEmployee) {
		responses.Error(c, 403, "forbidden", nil)
		return
	}
	l, err := h.svc.Create(employeeID, req.LeaveTypeID, req.StartDate, req.EndDate, req.Reason, p.UserID, clientIP(c), userAgent(c))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.Created(c, "leave request created", l)
}

func (h *LeaveHandler) List(c *gin.Context) {
	var params dto.LeaveListParams
	_ = c.ShouldBindQuery(&params)
	p := middleware.GetPrincipal(c)
	filter := repositories.LeaveFilter{
		EmployeeID: params.EmployeeID,
		TypeID:     params.Type,
		Status:     params.Status,
	}
	if f, ok := parseFormDate(params.From); ok {
		filter.From = f
	}
	if t, okT := parseFormDate(params.To); okT {
		filter.To = t
	}
	if p != nil && p.Role == string(models.RoleEmployee) {
		empID, err := h.empIDForUser(c)
		if err != nil {
			responses.Error(c, 404, err.Error(), nil)
			return
		}
		filter.EmployeeID = empID
	}
	pg := utils.NewPagination(params.Page, params.PageSize)
	items, total, err := h.svc.List(pg, filter)
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "leaves", responses.List{Items: items, Total: total, Page: pg.Page, PageSize: pg.PageSize, TotalPages: pg.TotalPages(total)})
}

func (h *LeaveHandler) Get(c *gin.Context) {
	l, err := h.svc.Get(c.Param("id"))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	if p := middleware.GetPrincipal(c); p != nil && p.Role == string(models.RoleEmployee) {
		empID, err := h.empIDForUser(c)
		if err != nil || l.EmployeeID != empID {
			responses.Error(c, 403, "forbidden", nil)
			return
		}
	}
	responses.OK(c, "leave", l)
}

func (h *LeaveHandler) Approve(c *gin.Context) {
	var req dto.LeaveDecisionRequest
	_ = c.ShouldBindJSON(&req)
	p := middleware.MustPrincipal(c)
	l, err := h.svc.Approve(c.Param("id"), req.Note, p.UserID, clientIP(c), userAgent(c))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "leave approved", l)
}

func (h *LeaveHandler) Reject(c *gin.Context) {
	var req dto.LeaveDecisionRequest
	_ = c.ShouldBindJSON(&req)
	p := middleware.MustPrincipal(c)
	l, err := h.svc.Reject(c.Param("id"), req.Note, p.UserID, clientIP(c), userAgent(c))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "leave rejected", l)
}

func (h *LeaveHandler) Types(c *gin.Context) {
	items, err := h.svc.LeaveTypes()
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "leave types", items)
}

func (h *LeaveHandler) SetBalance(c *gin.Context) {
	var req struct {
		EmployeeID  string `json:"employee_id" binding:"required"`
		LeaveTypeID string `json:"leave_type_id" binding:"required"`
		Year        int    `json:"year"`
		Entitlement int    `json:"entitlement" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.Error(c, 400, "validation failed", utils.ConvertValidationErrors(err))
		return
	}
	p := middleware.MustPrincipal(c)
	bal, err := h.svc.SetBalance(req.EmployeeID, req.LeaveTypeID, req.Year, req.Entitlement, p.UserID, clientIP(c), userAgent(c))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "leave balance updated", bal)
}

func (h *LeaveHandler) Balances(c *gin.Context) {
	p := middleware.GetPrincipal(c)
	employeeID := c.Query("employee_id")
	year, _ := strconv.Atoi(c.Query("year"))
	if p != nil && p.Role == string(models.RoleEmployee) {
		var err error
		employeeID, err = h.empIDForUser(c)
		if err != nil {
			responses.Error(c, 404, err.Error(), nil)
			return
		}
	}
	items, err := h.svc.Balances(employeeID, year)
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "leave balances", items)
}

func (h *LeaveHandler) empIDForUser(c *gin.Context) (string, error) {
	empID, _ := c.Get("employee_id")
	if id, ok := empID.(string); ok && id != "" {
		return id, nil
	}
	return "", errNoEmployeeLink
}
