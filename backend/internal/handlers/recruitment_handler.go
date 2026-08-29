package handlers

import (
	"github.com/gin-gonic/gin"

	"github.com/emplyra/backend/internal/dto"
	"github.com/emplyra/backend/internal/middleware"
	"github.com/emplyra/backend/internal/responses"
	"github.com/emplyra/backend/internal/services"
	"github.com/emplyra/backend/internal/utils"
)

type RecruitmentHandler struct {
	svc *services.RecruitmentService
}

func NewRecruitmentHandler(svc *services.RecruitmentService) *RecruitmentHandler {
	return &RecruitmentHandler{svc: svc}
}

// --- Job posts ---

func (h *RecruitmentHandler) CreateJob(c *gin.Context) {
	var req dto.JobPostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.Error(c, 400, "validation failed", utils.ConvertValidationErrors(err))
		return
	}
	a := middleware.MustPrincipal(c)
	j, err := h.svc.CreateJob(middleware.TenantID(c), services.JobPostInput{Title: req.Title, DepartmentID: req.DepartmentID, Description: req.Description, Requirements: req.Requirements, Status: req.Status, Deadline: req.Deadline, Vacancies: req.Vacancies}, a.UserID, clientIP(c), userAgent(c))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.Created(c, "job post created", j)
}

func (h *RecruitmentHandler) UpdateJob(c *gin.Context) {
	var req dto.JobPostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.Error(c, 400, "validation failed", utils.ConvertValidationErrors(err))
		return
	}
	a := middleware.MustPrincipal(c)
	j, err := h.svc.UpdateJob(middleware.TenantID(c), c.Param("id"), services.JobPostInput{Title: req.Title, DepartmentID: req.DepartmentID, Description: req.Description, Requirements: req.Requirements, Status: req.Status, Deadline: req.Deadline, Vacancies: req.Vacancies}, a.UserID, clientIP(c), userAgent(c))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "job post updated", j)
}

func (h *RecruitmentHandler) DeleteJob(c *gin.Context) {
	a := middleware.MustPrincipal(c)
	if err := h.svc.DeleteJob(middleware.TenantID(c), c.Param("id"), a.UserID, clientIP(c), userAgent(c)); err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "job post deleted", nil)
}

func (h *RecruitmentHandler) GetJob(c *gin.Context) {
	j, err := h.svc.Job(middleware.TenantID(c), c.Param("id"))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "job post", j)
}

func (h *RecruitmentHandler) ListJobs(c *gin.Context) {
	p := utils.NewPagination(c.Query("page"), c.Query("page_size"))
	items, total, err := h.svc.Jobs(middleware.TenantID(c), p, c.Query("department_id"), c.Query("status"), c.Query("search"))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "job posts", responses.List{Items: items, Total: total, Page: p.Page, PageSize: p.PageSize, TotalPages: p.TotalPages(total)})
}

// --- Candidates ---

func (h *RecruitmentHandler) CreateCandidate(c *gin.Context) {
	var req dto.CandidateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.Error(c, 400, "validation failed", utils.ConvertValidationErrors(err))
		return
	}
	a := middleware.MustPrincipal(c)
	cand, err := h.svc.CreateCandidate(middleware.TenantID(c), services.CandidateInput{FirstName: req.FirstName, LastName: req.LastName, Email: req.Email, Phone: req.Phone, Source: req.Source, Status: req.Status, Notes: req.Notes, ResumePath: req.ResumePath, Address: req.Address, DateOfBirth: req.DateOfBirth, Education: req.Education, Experience: req.Experience, Skills: req.Skills}, a.UserID, clientIP(c), userAgent(c))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.Created(c, "candidate created", cand)
}

func (h *RecruitmentHandler) UpdateCandidate(c *gin.Context) {
	var req dto.CandidateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.Error(c, 400, "validation failed", utils.ConvertValidationErrors(err))
		return
	}
	a := middleware.MustPrincipal(c)
	cand, err := h.svc.UpdateCandidate(middleware.TenantID(c), c.Param("id"), services.CandidateInput{FirstName: req.FirstName, LastName: req.LastName, Email: req.Email, Phone: req.Phone, Source: req.Source, Status: req.Status, Notes: req.Notes, ResumePath: req.ResumePath, Address: req.Address, DateOfBirth: req.DateOfBirth, Education: req.Education, Experience: req.Experience, Skills: req.Skills}, a.UserID, clientIP(c), userAgent(c))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "candidate updated", cand)
}

func (h *RecruitmentHandler) DeleteCandidate(c *gin.Context) {
	a := middleware.MustPrincipal(c)
	if err := h.svc.DeleteCandidate(middleware.TenantID(c), c.Param("id"), a.UserID, clientIP(c), userAgent(c)); err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "candidate deleted", nil)
}

func (h *RecruitmentHandler) GetCandidate(c *gin.Context) {
	cand, err := h.svc.Candidate(middleware.TenantID(c), c.Param("id"))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "candidate", cand)
}

func (h *RecruitmentHandler) ListCandidates(c *gin.Context) {
	p := utils.NewPagination(c.Query("page"), c.Query("page_size"))
	items, total, err := h.svc.Candidates(middleware.TenantID(c), p, c.Query("status"), c.Query("search"))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "candidates", responses.List{Items: items, Total: total, Page: p.Page, PageSize: p.PageSize, TotalPages: p.TotalPages(total)})
}

// --- Applications ---

func (h *RecruitmentHandler) CreateApplication(c *gin.Context) {
	var req dto.ApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.Error(c, 400, "validation failed", utils.ConvertValidationErrors(err))
		return
	}
	a := middleware.MustPrincipal(c)
	app, err := h.svc.CreateApplication(middleware.TenantID(c), services.ApplicationInput{JobPostID: req.JobPostID, CandidateID: req.CandidateID, AppliedDate: req.AppliedDate, CoverLetter: req.CoverLetter}, a.UserID, clientIP(c), userAgent(c))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.Created(c, "application created", app)
}

func (h *RecruitmentHandler) UpdateApplicationStatus(c *gin.Context) {
	var req dto.ApplicationStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.Error(c, 400, "validation failed", utils.ConvertValidationErrors(err))
		return
	}
	a := middleware.MustPrincipal(c)
	app, err := h.svc.UpdateApplicationStatus(middleware.TenantID(c), c.Param("id"), req.Status, req.Note, a.UserID, clientIP(c), userAgent(c))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "application updated", app)
}

func (h *RecruitmentHandler) GetApplication(c *gin.Context) {
	app, err := h.svc.Application(middleware.TenantID(c), c.Param("id"))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "application", app)
}

func (h *RecruitmentHandler) ListApplications(c *gin.Context) {
	p := utils.NewPagination(c.Query("page"), c.Query("page_size"))
	items, total, err := h.svc.Applications(middleware.TenantID(c), p, c.Query("job_post_id"), c.Query("candidate_id"), c.Query("status"))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "applications", responses.List{Items: items, Total: total, Page: p.Page, PageSize: p.PageSize, TotalPages: p.TotalPages(total)})
}

// --- Interviews ---

func (h *RecruitmentHandler) ScheduleInterview(c *gin.Context) {
	var req dto.InterviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.Error(c, 400, "validation failed", utils.ConvertValidationErrors(err))
		return
	}
	a := middleware.MustPrincipal(c)
	i, err := h.svc.ScheduleInterview(middleware.TenantID(c), services.InterviewInput{ApplicationID: req.ApplicationID, InterviewerID: req.InterviewerID, ScheduledAt: req.ScheduledAt, Type: req.Type, DurationMin: req.DurationMin}, a.UserID, clientIP(c), userAgent(c))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.Created(c, "interview scheduled", i)
}

func (h *RecruitmentHandler) CompleteInterview(c *gin.Context) {
	var req dto.InterviewFeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.Error(c, 400, "validation failed", utils.ConvertValidationErrors(err))
		return
	}
	a := middleware.MustPrincipal(c)
	i, err := h.svc.CompleteInterview(middleware.TenantID(c), c.Param("id"), req.Status, req.Feedback, req.Score, a.UserID, clientIP(c), userAgent(c))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "interview updated", i)
}

func (h *RecruitmentHandler) GetInterview(c *gin.Context) {
	i, err := h.svc.Interview(middleware.TenantID(c), c.Param("id"))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "interview", i)
}

func (h *RecruitmentHandler) ListInterviews(c *gin.Context) {
	p := utils.NewPagination(c.Query("page"), c.Query("page_size"))
	items, total, err := h.svc.Interviews(middleware.TenantID(c), p, c.Query("application_id"), c.Query("interviewer_id"), c.Query("status"), c.Query("from"), c.Query("to"))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "interviews", responses.List{Items: items, Total: total, Page: p.Page, PageSize: p.PageSize, TotalPages: p.TotalPages(total)})
}

// --- Onboarding ---

func (h *RecruitmentHandler) CreateOnboarding(c *gin.Context) {
	var req dto.OnboardingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.Error(c, 400, "validation failed", utils.ConvertValidationErrors(err))
		return
	}
	a := middleware.MustPrincipal(c)
	o, err := h.svc.CreateOnboarding(middleware.TenantID(c), struct {
		EmployeeID, CandidateID, StartDate, Notes string
		Tasks                                     []string
	}{req.EmployeeID, req.CandidateID, req.StartDate, req.Notes, req.Tasks}, a.UserID, clientIP(c), userAgent(c))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.Created(c, "onboarding created", o)
}

func (h *RecruitmentHandler) UpdateOnboarding(c *gin.Context) {
	var req dto.OnboardingUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.Error(c, 400, "validation failed", utils.ConvertValidationErrors(err))
		return
	}
	a := middleware.MustPrincipal(c)
	o, err := h.svc.UpdateOnboarding(middleware.TenantID(c), c.Param("id"), req.Status, req.Notes, req.Tasks, a.UserID, clientIP(c), userAgent(c))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "onboarding updated", o)
}

func (h *RecruitmentHandler) GetOnboarding(c *gin.Context) {
	o, err := h.svc.Onboarding(middleware.TenantID(c), c.Param("id"))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "onboarding", o)
}

func (h *RecruitmentHandler) ListOnboardings(c *gin.Context) {
	p := utils.NewPagination(c.Query("page"), c.Query("page_size"))
	items, total, err := h.svc.Onboardings(middleware.TenantID(c), p, c.Query("employee_id"), c.Query("status"))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.OK(c, "onboardings", responses.List{Items: items, Total: total, Page: p.Page, PageSize: p.PageSize, TotalPages: p.TotalPages(total)})
}

// --- Hire ---

func (h *RecruitmentHandler) Hire(c *gin.Context) {
	var req dto.HireCandidateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.Error(c, 400, "validation failed", utils.ConvertValidationErrors(err))
		return
	}
	a := middleware.MustPrincipal(c)
	emp, err := h.svc.Hire(middleware.TenantID(c), services.HireInput{ApplicationID: req.ApplicationID, EmployeeCode: req.EmployeeCode, DepartmentID: req.DepartmentID, DesignationID: req.DesignationID, ManagerID: req.ManagerID, JoiningDate: req.JoiningDate, CreateUser: req.CreateUser}, a.UserID, clientIP(c), userAgent(c))
	if err != nil {
		mapServiceError(c, err)
		return
	}
	responses.Created(c, "candidate hired and employee created", emp)
}
