package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/emplyra/backend/internal/middleware"
	"github.com/emplyra/backend/internal/models"
	"github.com/emplyra/backend/internal/repositories"
	"github.com/emplyra/backend/internal/responses"
	"github.com/emplyra/backend/internal/services"
	"github.com/emplyra/backend/internal/utils"
)

type ReportHandler struct {
	svc *services.ReportService
	emp *repositories.EmployeeRepository
}

func NewReportHandler(svc *services.ReportService, emp *repositories.EmployeeRepository) *ReportHandler {
	return &ReportHandler{svc: svc, emp: emp}
}

func (h *ReportHandler) Headcount(c *gin.Context) {
	rows, err := h.svc.HeadcountByDepartment()
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "headcount report", rows)
}

func (h *ReportHandler) Attendance(c *gin.Context) {
	rows, avg, err := h.svc.AttendanceSummary(c.Query("from"), c.Query("to"))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "attendance report", gin.H{"summary": rows, "avg_work_hours": avg})
}

func (h *ReportHandler) Leaves(c *gin.Context) {
	s, err := h.svc.LeaveSummary()
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "leave report", s)
}

func (h *ReportHandler) Payroll(c *gin.Context) {
	month, _ := strconv.Atoi(c.Query("month"))
	year, _ := strconv.Atoi(c.Query("year"))
	s, err := h.svc.PayrollSummary(month, year)
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "payroll report", s)
}

func (h *ReportHandler) Recruitment(c *gin.Context) {
	rows, err := h.svc.RecruitmentFunnel()
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "recruitment report", rows)
}

func (h *ReportHandler) Holidays(c *gin.Context) {
	items, err := h.svc.UpcomingHolidays(20)
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "upcoming holidays", items)
}

func (h *ReportHandler) Dashboard(c *gin.Context) {
	p := middleware.GetPrincipal(c)
	employeeID := ""
	if p != nil && p.Role == string(models.RoleEmployee) {
		if emp, err := h.emp.FindByUserID(p.UserID); err == nil {
			employeeID = emp.ID
		}
	}
	s, err := h.svc.Summary(employeeID)
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "dashboard summary", s)
}

// AuditHandler lists activity logs (SUPER_ADMIN only).
type AuditHandler struct {
	logs *repositories.AuditLogRepository
}

func NewAuditHandler(logs *repositories.AuditLogRepository) *AuditHandler {
	return &AuditHandler{logs: logs}
}

func (h *AuditHandler) List(c *gin.Context) {
	p := utils.NewPagination(c.Query("page"), c.Query("page_size"))
	items, total, err := h.logs.List(p, c.Query("user_id"), c.Query("resource"), c.Query("action"))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "audit logs", responses.List{Items: items, Total: total, Page: p.Page, PageSize: p.PageSize, TotalPages: p.TotalPages(total)})
}
