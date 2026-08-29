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

type TrainingHandler struct {
	svc *services.TrainingService
}

func NewTrainingHandler(svc *services.TrainingService) *TrainingHandler {
	return &TrainingHandler{svc: svc}
}

func (h *TrainingHandler) CreateProgram(c *gin.Context) {
	var req dto.TrainingProgramRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.Error(c, 400, "validation failed", utils.ConvertValidationErrors(err))
		return
	}
	a := middleware.MustPrincipal(c)
	p, err := h.svc.CreateProgram(services.ProgramInput{Title: req.Title, Description: req.Description, Provider: req.Provider, StartDate: req.StartDate, EndDate: req.EndDate, Location: req.Location, MaxSeats: req.MaxSeats, Status: req.Status}, a.UserID, clientIP(c), userAgent(c))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.Created(c, "training program created", p)
}

func (h *TrainingHandler) UpdateProgram(c *gin.Context) {
	var req dto.TrainingProgramRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.Error(c, 400, "validation failed", utils.ConvertValidationErrors(err))
		return
	}
	a := middleware.MustPrincipal(c)
	p, err := h.svc.UpdateProgram(c.Param("id"), services.ProgramInput{Title: req.Title, Description: req.Description, Provider: req.Provider, StartDate: req.StartDate, EndDate: req.EndDate, Location: req.Location, MaxSeats: req.MaxSeats, Status: req.Status}, a.UserID, clientIP(c), userAgent(c))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "training program updated", p)
}

func (h *TrainingHandler) DeleteProgram(c *gin.Context) {
	a := middleware.MustPrincipal(c)
	if err := h.svc.DeleteProgram(c.Param("id"), a.UserID, clientIP(c), userAgent(c)); err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "training program deleted", nil)
}

func (h *TrainingHandler) GetProgram(c *gin.Context) {
	p, err := h.svc.Program(c.Param("id"))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "training program", p)
}

func (h *TrainingHandler) ListPrograms(c *gin.Context) {
	p := utils.NewPagination(c.Query("page"), c.Query("page_size"))
	items, total, err := h.svc.Programs(p, c.Query("status"))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "training programs", responses.List{Items: items, Total: total, Page: p.Page, PageSize: p.PageSize, TotalPages: p.TotalPages(total)})
}

func (h *TrainingHandler) CreateSchedule(c *gin.Context) {
	var req dto.TrainingScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.Error(c, 400, "validation failed", utils.ConvertValidationErrors(err))
		return
	}
	a := middleware.MustPrincipal(c)
	s, err := h.svc.CreateSchedule(struct {
		ProgramID, Date, StartTime, EndTime, Trainer, Location string
		MaxSeats                                               int
	}{req.ProgramID, req.Date, req.StartTime, req.EndTime, req.Trainer, req.Location, req.MaxSeats}, a.UserID, clientIP(c), userAgent(c))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.Created(c, "training schedule created", s)
}

func (h *TrainingHandler) UpdateSchedule(c *gin.Context) {
	var req dto.TrainingScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.Error(c, 400, "validation failed", utils.ConvertValidationErrors(err))
		return
	}
	a := middleware.MustPrincipal(c)
	s, err := h.svc.UpdateSchedule(c.Param("id"), struct {
		ProgramID, Date, StartTime, EndTime, Trainer, Location string
		MaxSeats                                               int
	}{req.ProgramID, req.Date, req.StartTime, req.EndTime, req.Trainer, req.Location, req.MaxSeats}, a.UserID, clientIP(c), userAgent(c))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "training schedule updated", s)
}

func (h *TrainingHandler) DeleteSchedule(c *gin.Context) {
	a := middleware.MustPrincipal(c)
	if err := h.svc.DeleteSchedule(c.Param("id"), a.UserID, clientIP(c), userAgent(c)); err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "training schedule deleted", nil)
}

func (h *TrainingHandler) ListSchedules(c *gin.Context) {
	items, err := h.svc.Schedules(c.Query("program_id"))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "training schedules", items)
}

func (h *TrainingHandler) Enroll(c *gin.Context) {
	var req dto.EnrollmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.Error(c, 400, "validation failed", utils.ConvertValidationErrors(err))
		return
	}
	employeeID := req.EmployeeID
	if employeeID == "" {
		employeeID = middleware.EmployeeID(c)
	}
	a := middleware.MustPrincipal(c)
	e, err := h.svc.Enroll(req.ProgramID, employeeID, a.UserID, clientIP(c), userAgent(c))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.Created(c, "enrollment created", e)
}

func (h *TrainingHandler) UpdateEnrollment(c *gin.Context) {
	var req dto.EnrollmentUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.Error(c, 400, "validation failed", utils.ConvertValidationErrors(err))
		return
	}
	a := middleware.MustPrincipal(c)
	e, err := h.svc.UpdateEnrollment(c.Param("id"), req.Status, a.UserID, clientIP(c), userAgent(c))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "enrollment updated", e)
}

func (h *TrainingHandler) ListEnrollments(c *gin.Context) {
	p := utils.NewPagination(c.Query("page"), c.Query("page_size"))
	employeeID := c.Query("employee_id")
	if employeeID == "" {
		principal := middleware.GetPrincipal(c)
		if principal != nil && principal.Role == string(models.RoleEmployee) {
			employeeID = middleware.EmployeeID(c)
		}
	}
	items, total, err := h.svc.Enrollments(p, c.Query("program_id"), employeeID, c.Query("status"))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "enrollments", responses.List{Items: items, Total: total, Page: p.Page, PageSize: p.PageSize, TotalPages: p.TotalPages(total)})
}
