package handlers

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/emplyra/backend/internal/dto"
	"github.com/emplyra/backend/internal/middleware"
	"github.com/emplyra/backend/internal/responses"
	"github.com/emplyra/backend/internal/services"
	"github.com/emplyra/backend/internal/utils"
)

type PayrollHandler struct {
	salary  *services.SalaryService
	payroll *services.PayrollService
}

func NewPayrollHandler(salary *services.SalaryService, payroll *services.PayrollService) *PayrollHandler {
	return &PayrollHandler{salary: salary, payroll: payroll}
}

// --- Salary structures ---

func (h *PayrollHandler) CreateStructure(c *gin.Context) {
	var req dto.SalaryStructureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.Error(c, 400, "validation failed", utils.ConvertValidationErrors(err))
		return
	}
	actor := middleware.MustPrincipal(c)
	st, err := h.salary.CreateStructure(middleware.TenantID(c), struct {
		EmployeeID, BasicSalary, Allowances, Bonus, OvertimeRate, TaxRate, TaxAmount, Deductions, EffectiveFrom, Status string
	}{req.EmployeeID, req.BasicSalary, req.Allowances, req.Bonus, req.OvertimeRate, req.TaxRate, req.TaxAmount, req.Deductions, req.EffectiveFrom, req.Status},
		actor.UserID, clientIP(c), userAgent(c))
	if err != nil {
		mapMoneyError(c, err)
		return
	}
	responses.Created(c, "salary structure created", st)
}

func (h *PayrollHandler) UpdateStructure(c *gin.Context) {
	var req dto.SalaryStructureUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.Error(c, 400, "validation failed", utils.ConvertValidationErrors(err))
		return
	}
	actor := middleware.MustPrincipal(c)
	st, err := h.salary.UpdateStructure(middleware.TenantID(c), c.Param("id"), struct {
		BasicSalary, Allowances, Bonus, OvertimeRate, TaxRate, TaxAmount, Deductions, EffectiveFrom, EffectiveUntil, Status string
	}{req.BasicSalary, req.Allowances, req.Bonus, req.OvertimeRate, req.TaxRate, req.TaxAmount, req.Deductions, req.EffectiveFrom, req.EffectiveUntil, req.Status},
		actor.UserID, clientIP(c), userAgent(c))
	if err != nil {
		mapMoneyError(c, err)
		return
	}
	responses.OK(c, "salary structure updated", st)
}

func (h *PayrollHandler) DeleteStructure(c *gin.Context) {
	actor := middleware.MustPrincipal(c)
	if err := h.salary.DeleteStructure(middleware.TenantID(c), c.Param("id"), actor.UserID, clientIP(c), userAgent(c)); err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "salary structure deleted", nil)
}

func (h *PayrollHandler) GetStructure(c *gin.Context) {
	st, err := h.salary.GetStructure(middleware.TenantID(c), c.Param("id"))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "salary structure", st)
}

func (h *PayrollHandler) ListStructures(c *gin.Context) {
	items, err := h.salary.ListStructures(middleware.TenantID(c), c.Query("employee_id"))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "salary structures", items)
}

// --- Payroll ---

func (h *PayrollHandler) Generate(c *gin.Context) {
	var req dto.GeneratePayrollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.Error(c, 400, "validation failed", utils.ConvertValidationErrors(err))
		return
	}
	actor := middleware.MustPrincipal(c)
	created, err := h.payroll.Generate(middleware.TenantID(c), req.Month, req.Year, actor.UserID, clientIP(c), userAgent(c))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.Created(c, "payroll generated", map[string]int{"created": created, "month": req.Month, "year": req.Year})
}

func (h *PayrollHandler) List(c *gin.Context) {
	var params dto.PayrollListParams
	_ = c.ShouldBindQuery(&params)
	month, _ := strconv.Atoi(params.Month)
	year, _ := strconv.Atoi(params.Year)
	pg := utils.NewPagination(params.Page, params.PageSize)
	items, total, err := h.payroll.List(middleware.TenantID(c), pg, month, year, params.EmployeeID, params.Status, params.DepartmentID)
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "payroll", responses.List{Items: items, Total: total, Page: pg.Page, PageSize: pg.PageSize, TotalPages: pg.TotalPages(total)})
}

func (h *PayrollHandler) Get(c *gin.Context) {
	p, err := h.payroll.Get(middleware.TenantID(c), c.Param("id"))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "payroll", p)
}

func (h *PayrollHandler) Process(c *gin.Context) {
	var req dto.ProcessPayrollRequest
	_ = c.ShouldBindJSON(&req)
	actor := middleware.MustPrincipal(c)
	p, err := h.payroll.Process(middleware.TenantID(c), c.Param("id"), struct{ Bonus, Overtime, Deductions, Notes string }{
		Bonus: req.Bonus, Overtime: req.Overtime, Deductions: req.Deductions, Notes: req.Notes,
	}, actor.UserID, clientIP(c), userAgent(c))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "payroll processed", p)
}

func (h *PayrollHandler) MarkPaid(c *gin.Context) {
	var req dto.MarkPaidRequest
	_ = c.ShouldBindJSON(&req)
	actor := middleware.MustPrincipal(c)
	p, err := h.payroll.MarkPaid(middleware.TenantID(c), c.Param("id"), req.PaymentRef, req.Notes, actor.UserID, clientIP(c), userAgent(c))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "payroll marked paid", p)
}

func (h *PayrollHandler) Cancel(c *gin.Context) {
	var req dto.CancelPayrollRequest
	_ = c.ShouldBindJSON(&req)
	actor := middleware.MustPrincipal(c)
	p, err := h.payroll.Cancel(middleware.TenantID(c), c.Param("id"), req.Notes, actor.UserID, clientIP(c), userAgent(c))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "payroll cancelled", p)
}

func (h *PayrollHandler) Payslip(c *gin.Context) {
	p, err := h.payroll.Payslip(middleware.TenantID(c), c.Param("id"))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	l := middleware.GetPrincipal(c)
	if l != nil && l.Role == "EMPLOYEE" {
		if p.Employee == nil || p.Employee.UserID == nil || *p.Employee.UserID != l.UserID {
			responses.Error(c, 403, "forbidden", nil)
			return
		}
	}
	responses.OK(c, "payslip", p)
}

func mapMoneyError(c *gin.Context, err error) {
	if err == nil {
		return
	}
	if errors.Is(err, services.ErrNotFound) {
		responses.Error(c, 404, err.Error(), nil)
		return
	}
	if errors.Is(err, services.ErrDuplicate) {
		responses.Error(c, 409, err.Error(), nil)
		return
	}
	responses.Error(c, 400, err.Error(), nil)
}
