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
)

// scopeEmployeeIDs returns the employee IDs a principal may see. Empty slice
// and empty-ok means "all employees".
func scopeEmployeeIDs(c *gin.Context, emp *repositories.EmployeeRepository) ([]string, bool, bool) {
	p := middleware.GetPrincipal(c)
	if p == nil {
		return nil, false, true
	}
	switch p.Role {
	case string(models.RoleEmployee):
		own := middleware.EmployeeID(c)
		if own == "" {
			if e, err := emp.FindByUserID(p.UserID); err == nil {
				own = e.ID
			}
		}
		if own == "" {
			return nil, false, true
		}
		return []string{own}, true, false
	case string(models.RoleManager):
		own := middleware.EmployeeID(c)
		if e, err := emp.FindByUserID(p.UserID); err == nil {
			own = e.ID
		}
		ids := []string{own}
		reports, _ := emp.DirectReports(own)
		ids = append(ids, reports...)
		return ids, true, false
	default:
		return nil, true, true
	}
}

func readScope(c *gin.Context, emp *repositories.EmployeeRepository) []string {
	ids, fullAccess, _ := scopeEmployeeIDs(c, emp)
	if fullAccess {
		return nil
	}
	return ids
}

type PerformanceHandler struct {
	svc *services.PerformanceService
	emp *repositories.EmployeeRepository
}

func NewPerformanceHandler(svc *services.PerformanceService, emp *repositories.EmployeeRepository) *PerformanceHandler {
	return &PerformanceHandler{svc: svc, emp: emp}
}

// --- Goals ---

func (h *PerformanceHandler) CreateGoal(c *gin.Context) {
	var req dto.GoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.Error(c, 400, "validation failed", utils.ConvertValidationErrors(err))
		return
	}
	setDefault := func() string {
		if req.EmployeeID == "" {
			return middleware.EmployeeID(c)
		}
		return req.EmployeeID
	}
	req.EmployeeID = setDefault()
	a := middleware.MustPrincipal(c)
	g, err := h.svc.CreateGoal(middleware.TenantID(c), services.GoalInput{EmployeeID: req.EmployeeID, Title: req.Title, Description: req.Description, TargetDate: req.TargetDate, Weight: req.Weight, Status: req.Status}, a.UserID, clientIP(c), userAgent(c))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.Created(c, "goal created", g)
}

func (h *PerformanceHandler) UpdateGoal(c *gin.Context) {
	var req dto.GoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.Error(c, 400, "validation failed", utils.ConvertValidationErrors(err))
		return
	}
	a := middleware.MustPrincipal(c)
	g, err := h.svc.UpdateGoal(middleware.TenantID(c), c.Param("id"), services.GoalInput{Title: req.Title, Description: req.Description, TargetDate: req.TargetDate, Weight: req.Weight, Status: req.Status}, a.UserID, clientIP(c), userAgent(c))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "goal updated", g)
}

func (h *PerformanceHandler) DeleteGoal(c *gin.Context) {
	a := middleware.MustPrincipal(c)
	if err := h.svc.DeleteGoal(middleware.TenantID(c), c.Param("id"), a.UserID, clientIP(c), userAgent(c)); err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "goal deleted", nil)
}

func (h *PerformanceHandler) GetGoal(c *gin.Context) {
	g, err := h.svc.Goal(middleware.TenantID(c), c.Param("id"))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "goal", g)
}

func (h *PerformanceHandler) ListGoals(c *gin.Context) {
	p := utils.NewPagination(c.Query("page"), c.Query("page_size"))
	items, total, err := h.svc.Goals(middleware.TenantID(c), p, readScope(c, h.emp), c.Query("status"))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "goals", responses.List{Items: items, Total: total, Page: p.Page, PageSize: p.PageSize, TotalPages: p.TotalPages(total)})
}

// --- KPIs ---

func (h *PerformanceHandler) CreateKPI(c *gin.Context) {
	var req dto.KPIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.Error(c, 400, "validation failed", utils.ConvertValidationErrors(err))
		return
	}
	if req.EmployeeID == "" {
		req.EmployeeID = middleware.EmployeeID(c)
	}
	a := middleware.MustPrincipal(c)
	k, err := h.svc.CreateKPI(middleware.TenantID(c), services.KPIInput{EmployeeID: req.EmployeeID, Name: req.Name, Description: req.Description, Target: req.Target, Actual: req.Actual, Unit: req.Unit, Weight: req.Weight, Period: req.Period, Score: req.Score}, a.UserID, clientIP(c), userAgent(c))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.Created(c, "kpi created", k)
}

func (h *PerformanceHandler) UpdateKPI(c *gin.Context) {
	var req dto.KPIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.Error(c, 400, "validation failed", utils.ConvertValidationErrors(err))
		return
	}
	a := middleware.MustPrincipal(c)
	k, err := h.svc.UpdateKPI(middleware.TenantID(c), c.Param("id"), services.KPIInput{Name: req.Name, Description: req.Description, Target: req.Target, Actual: req.Actual, Unit: req.Unit, Weight: req.Weight, Period: req.Period, Score: req.Score}, a.UserID, clientIP(c), userAgent(c))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "kpi updated", k)
}

func (h *PerformanceHandler) DeleteKPI(c *gin.Context) {
	a := middleware.MustPrincipal(c)
	if err := h.svc.DeleteKPI(middleware.TenantID(c), c.Param("id"), a.UserID, clientIP(c), userAgent(c)); err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "kpi deleted", nil)
}

func (h *PerformanceHandler) GetKPI(c *gin.Context) {
	k, err := h.svc.KPI(middleware.TenantID(c), c.Param("id"))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "kpi", k)
}

func (h *PerformanceHandler) ListKPIs(c *gin.Context) {
	p := utils.NewPagination(c.Query("page"), c.Query("page_size"))
	items, total, err := h.svc.KPIs(middleware.TenantID(c), p, readScope(c, h.emp), c.Query("period"))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "kpis", responses.List{Items: items, Total: total, Page: p.Page, PageSize: p.PageSize, TotalPages: p.TotalPages(total)})
}

// --- Reviews ---

func (h *PerformanceHandler) CreateReview(c *gin.Context) {
	var req dto.ReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.Error(c, 400, "validation failed", utils.ConvertValidationErrors(err))
		return
	}
	a := middleware.MustPrincipal(c)
	r, err := h.svc.CreateReview(middleware.TenantID(c), services.ReviewInput{EmployeeID: req.EmployeeID, ReviewerID: req.ReviewerID, Period: req.Period, DueDate: req.DueDate}, a.UserID, clientIP(c), userAgent(c))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.Created(c, "review created", r)
}

func (h *PerformanceHandler) SubmitReview(c *gin.Context) {
	var req dto.ReviewSubmissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.Error(c, 400, "validation failed", utils.ConvertValidationErrors(err))
		return
	}
	a := middleware.MustPrincipal(c)
	r, err := h.svc.SubmitReview(middleware.TenantID(c), c.Param("id"), req.SelfEvaluation, req.ManagerFeedback, req.Status, req.Score, a.UserID, clientIP(c), userAgent(c))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "review submitted", r)
}

func (h *PerformanceHandler) GetReview(c *gin.Context) {
	r, err := h.svc.Review(middleware.TenantID(c), c.Param("id"))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "review", r)
}

func (h *PerformanceHandler) ListReviews(c *gin.Context) {
	p := utils.NewPagination(c.Query("page"), c.Query("page_size"))
	items, total, err := h.svc.Reviews(middleware.TenantID(c), p, readScope(c, h.emp), c.Query("reviewer_id"), c.Query("status"))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "reviews", responses.List{Items: items, Total: total, Page: p.Page, PageSize: p.PageSize, TotalPages: p.TotalPages(total)})
}
