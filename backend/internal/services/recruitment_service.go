package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gorm.io/datatypes"

	"github.com/emplyra/backend/internal/auditmanager"
	"github.com/emplyra/backend/internal/models"
	"github.com/emplyra/backend/internal/notifications"
	"github.com/emplyra/backend/internal/repositories"
	"github.com/emplyra/backend/internal/utils"
)

type RecruitmentService struct {
	jobs       *repositories.JobPostRepository
	candidates *repositories.CandidateRepository
	apps       *repositories.ApplicationRepository
	interviews *repositories.InterviewRepository
	onboard    *repositories.OnboardingRepository
	employees  *EmployeeService
	users      *repositories.UserRepository
	roles      *repositories.RoleRepository
	notify     *notifications.Service
	audit      *auditmanager.Service
}

func NewRecruitmentService(
	jobs *repositories.JobPostRepository,
	candidates *repositories.CandidateRepository,
	apps *repositories.ApplicationRepository,
	interviews *repositories.InterviewRepository,
	onboard *repositories.OnboardingRepository,
	employees *EmployeeService,
	users *repositories.UserRepository,
	roles *repositories.RoleRepository,
	notify *notifications.Service,
	audit *auditmanager.Service,
) *RecruitmentService {
	return &RecruitmentService{
		jobs: jobs, candidates: candidates, apps: apps, interviews: interviews, onboard: onboard,
		employees: employees, users: users, roles: roles, notify: notify, audit: audit,
	}
}

// --- Job posts ---

type JobPostInput struct {
	Title, DepartmentID, Description, Requirements, Status, Deadline string
	Vacancies                                                        int
}

func (s *RecruitmentService) CreateJob(tenantID string, in JobPostInput, actorID, ip, ua string) (*models.JobPost, error) {
	j := &models.JobPost{
		TenantID: tenantID, Title: in.Title, DepartmentID: stringPtr(in.DepartmentID),
		Description: in.Description, Requirements: in.Requirements,
		Vacancies: in.Vacancies, Status: models.JobPostStatus(orStr(in.Status, string(models.JobOpen))),
		PostedBy: &actorID,
	}
	if in.Deadline != "" {
		if d, err := time.Parse("2006-01-02", in.Deadline); err == nil {
			dd := datatypes.Date(d)
			j.Deadline = &dd
		}
	}
	if j.Vacancies == 0 {
		j.Vacancies = 1
	}
	if err := s.jobs.Create(j); err != nil {
		return nil, err
	}
	s.audit.Record(actorID, models.ActionCreate, "job_post", j.ID, ip, ua, map[string]string{"title": j.Title})
	return j, nil
}

func (s *RecruitmentService) UpdateJob(tenantID, id string, in JobPostInput, actorID, ip, ua string) (*models.JobPost, error) {
	if _, err := s.jobs.FindByID(tenantID, id); err != nil {
		return nil, ErrNotFound
	}
	fields := map[string]interface{}{}
	if in.Title != "" {
		fields["title"] = in.Title
	}
	if in.DepartmentID != "" {
		fields["department_id"] = in.DepartmentID
	}
	if in.Description != "" {
		fields["description"] = in.Description
	}
	if in.Requirements != "" {
		fields["requirements"] = in.Requirements
	}
	if in.Status != "" {
		fields["status"] = in.Status
	}
	if in.Vacancies != 0 {
		fields["vacancies"] = in.Vacancies
	}
	if in.Deadline != "" {
		if d, err := time.Parse("2006-01-02", in.Deadline); err == nil {
			fields["deadline"] = datatypes.Date(d)
		}
	}
	if err := s.jobs.Update(tenantID, id, fields); err != nil {
		return nil, err
	}
	s.audit.Record(actorID, models.ActionUpdate, "job_post", id, ip, ua, nil)
	return s.jobs.FindByID(tenantID, id)
}

func (s *RecruitmentService) DeleteJob(tenantID, id, actorID, ip, ua string) error {
	if _, err := s.jobs.FindByID(tenantID, id); err != nil {
		return ErrNotFound
	}
	if err := s.jobs.Delete(tenantID, id); err != nil {
		return err
	}
	s.audit.Record(actorID, models.ActionDelete, "job_post", id, ip, ua, nil)
	return nil
}

func (s *RecruitmentService) Job(tenantID, id string) (*models.JobPost, error) {
	j, err := s.jobs.FindByID(tenantID, id)
	if err != nil {
		return nil, ErrNotFound
	}
	return j, nil
}

func (s *RecruitmentService) Jobs(tenantID string, p utils.Pagination, departmentID, status, search string) ([]models.JobPost, int64, error) {
	return s.jobs.List(tenantID, p, departmentID, status, search)
}

// --- Candidates ---

type CandidateInput struct {
	FirstName, LastName, Email, Phone, Source, Status, Notes, ResumePath string
	Address, DateOfBirth, Education, Experience, Skills                  string
}

func (s *RecruitmentService) CreateCandidate(tenantID string, in CandidateInput, actorID, ip, ua string) (*models.Candidate, error) {
	if s.candidates.AlreadyEmployeeEmail(tenantID, in.Email) {
		return nil, ErrCandidateAlreadyEmployee
	}
	if existing, _ := s.candidates.FindByEmail(tenantID, in.Email); existing != nil {
		return existing, nil
	}
	c := &models.Candidate{
		TenantID:  tenantID,
		FirstName: in.FirstName, LastName: in.LastName, Email: in.Email, Phone: in.Phone,
		Source: in.Source, Status: models.CandidateStatus(orStr(in.Status, string(models.CandidateNew))),
		Notes: in.Notes, ResumePath: in.ResumePath,
		Address: in.Address, DateOfBirth: in.DateOfBirth,
		Education: in.Education, Experience: in.Experience, Skills: in.Skills,
	}
	if err := s.candidates.Create(c); err != nil {
		return nil, err
	}
	s.audit.Record(actorID, models.ActionCreate, "candidate", c.ID, ip, ua, map[string]string{"email": c.Email})
	return c, nil
}

func (s *RecruitmentService) UpdateCandidate(tenantID, id string, in CandidateInput, actorID, ip, ua string) (*models.Candidate, error) {
	ex, err := s.candidates.FindByID(tenantID, id)
	if err != nil {
		return nil, ErrNotFound
	}
	if in.Email != "" && in.Email != ex.Email && s.candidates.AlreadyEmployeeEmail(tenantID, in.Email) {
		return nil, ErrCandidateAlreadyEmployee
	}
	fields := map[string]interface{}{}
	if in.FirstName != "" {
		fields["first_name"] = in.FirstName
	}
	if in.LastName != "" {
		fields["last_name"] = in.LastName
	}
	if in.Email != "" {
		fields["email"] = in.Email
	}
	if in.Phone != "" {
		fields["phone"] = in.Phone
	}
	if in.Source != "" {
		fields["source"] = in.Source
	}
	if in.Status != "" {
		fields["status"] = in.Status
	}
	if in.Notes != "" {
		fields["notes"] = in.Notes
	}
	if in.ResumePath != "" {
		fields["resume_path"] = in.ResumePath
	}
	if in.Address != "" {
		fields["address"] = in.Address
	}
	if in.DateOfBirth != "" {
		fields["date_of_birth"] = in.DateOfBirth
	}
	if in.Education != "" {
		fields["education"] = in.Education
	}
	if in.Experience != "" {
		fields["experience"] = in.Experience
	}
	if in.Skills != "" {
		fields["skills"] = in.Skills
	}
	if err := s.candidates.Update(tenantID, id, fields); err != nil {
		return nil, err
	}
	s.audit.Record(actorID, models.ActionUpdate, "candidate", id, ip, ua, nil)
	return s.candidates.FindByID(tenantID, id)
}

func (s *RecruitmentService) DeleteCandidate(tenantID, id, actorID, ip, ua string) error {
	if _, err := s.candidates.FindByID(tenantID, id); err != nil {
		return ErrNotFound
	}
	if err := s.candidates.Delete(tenantID, id); err != nil {
		return err
	}
	s.audit.Record(actorID, models.ActionDelete, "candidate", id, ip, ua, nil)
	return nil
}

func (s *RecruitmentService) Candidate(tenantID, id string) (*models.Candidate, error) {
	c, err := s.candidates.FindByID(tenantID, id)
	if err != nil {
		return nil, ErrNotFound
	}
	return c, nil
}

func (s *RecruitmentService) Candidates(tenantID string, p utils.Pagination, status, search string) ([]models.Candidate, int64, error) {
	return s.candidates.List(tenantID, p, status, search)
}

// --- Applications ---

type ApplicationInput struct {
	JobPostID, CandidateID, AppliedDate, CoverLetter string
}

func (s *RecruitmentService) CreateApplication(tenantID string, in ApplicationInput, actorID, ip, ua string) (*models.Application, error) {
	if _, err := s.jobs.FindByID(tenantID, in.JobPostID); err != nil {
		return nil, ErrNotFound
	}
	if _, err := s.candidates.FindByID(tenantID, in.CandidateID); err != nil {
		return nil, ErrNotFound
	}
	exists, err := s.apps.Exists(tenantID, in.CandidateID, in.JobPostID)
	if err != nil || exists {
		return nil, ErrDuplicateApplication
	}
	date := time.Now()
	if in.AppliedDate != "" {
		if d, err := time.Parse("2006-01-02", in.AppliedDate); err == nil {
			date = d
		}
	}
	a := &models.Application{
		TenantID:  tenantID,
		JobPostID: in.JobPostID, CandidateID: in.CandidateID,
		AppliedDate: datatypes.Date(date), CoverLetter: in.CoverLetter,
		Status: models.AppApplied,
	}
	if err := s.apps.Create(a); err != nil {
		return nil, err
	}
	s.audit.Record(actorID, models.ActionCreate, "application", a.ID, ip, ua, map[string]string{"candidate_id": in.CandidateID, "job_post_id": in.JobPostID})
	return a, nil
}

func (s *RecruitmentService) UpdateApplicationStatus(tenantID, id, status, note string, actorID, ip, ua string) (*models.Application, error) {
	app, err := s.apps.FindByID(tenantID, id)
	if err != nil {
		return nil, ErrNotFound
	}
	fields := map[string]interface{}{"status": status}
	if note != "" {
		if app.JobPost != nil && app.Candidate != nil {
			s.notify.NotifyByEmail(app.Candidate.Email, app.Candidate.FirstName, "Application "+status, fmt.Sprintf("Your application for '%s' has been updated to %s.", app.JobPost.Title, status), models.NotifyRecruitment, "")
		}
	}
	if err := s.apps.Update(tenantID, id, fields); err != nil {
		return nil, err
	}
	s.audit.Record(actorID, models.ActionUpdate, "application", id, ip, ua, map[string]string{"status": status})
	return s.apps.FindByID(tenantID, id)
}

func (s *RecruitmentService) Application(tenantID, id string) (*models.Application, error) {
	a, err := s.apps.FindByID(tenantID, id)
	if err != nil {
		return nil, ErrNotFound
	}
	return a, nil
}

func (s *RecruitmentService) Applications(tenantID string, p utils.Pagination, jobID, candidateID, status string) ([]models.Application, int64, error) {
	return s.apps.List(tenantID, p, jobID, candidateID, status)
}

// --- Interviews ---

type InterviewInput struct {
	ApplicationID, InterviewerID, ScheduledAt, Type string
	DurationMin                                     int
}

func (s *RecruitmentService) ScheduleInterview(tenantID string, in InterviewInput, actorID, ip, ua string) (*models.Interview, error) {
	if _, err := s.apps.FindByID(tenantID, in.ApplicationID); err != nil {
		return nil, ErrNotFound
	}
	at, err := time.Parse(time.RFC3339, in.ScheduledAt)
	if err != nil {
		return nil, errors.New("invalid scheduled_at (expected RFC3339)")
	}
	i := &models.Interview{
		TenantID:      tenantID,
		ApplicationID: in.ApplicationID, InterviewerID: stringPtr(in.InterviewerID),
		ScheduledAt: at, DurationMin: in.DurationMin,
		Type:   models.InterviewType(orStr(in.Type, string(models.InterviewInPerson))),
		Status: models.InterviewScheduled,
	}
	if err := s.interviews.Create(i); err != nil {
		return nil, err
	}
	s.audit.Record(actorID, models.ActionCreate, "interview", i.ID, ip, ua, map[string]string{"application_id": in.ApplicationID})
	return i, nil
}

func (s *RecruitmentService) CompleteInterview(tenantID, id, status, feedback string, score *float64, actorID, ip, ua string) (*models.Interview, error) {
	if _, err := s.interviews.FindByID(tenantID, id); err != nil {
		return nil, ErrNotFound
	}
	fields := map[string]interface{}{"feedback": feedback, "status": status}
	if score != nil {
		fields["score"] = *score
	}
	if err := s.interviews.Update(tenantID, id, fields); err != nil {
		return nil, err
	}
	s.audit.Record(actorID, models.ActionUpdate, "interview", id, ip, ua, map[string]string{"status": status})
	return s.interviews.FindByID(tenantID, id)
}

func (s *RecruitmentService) Interview(tenantID, id string) (*models.Interview, error) {
	i, err := s.interviews.FindByID(tenantID, id)
	if err != nil {
		return nil, ErrNotFound
	}
	return i, nil
}

func (s *RecruitmentService) Interviews(tenantID string, p utils.Pagination, applicationID, interviewerID, status, fromStr, toStr string) ([]models.Interview, int64, error) {
	var from, to *time.Time
	if f, err := time.Parse("2006-01-02", fromStr); err == nil {
		from = &f
	}
	if t, err := time.Parse("2006-01-02", toStr); err == nil {
		to = &t
	}
	return s.interviews.List(tenantID, p, applicationID, interviewerID, status, from, to)
}

// --- Onboarding ---

func (s *RecruitmentService) CreateOnboarding(tenantID string, in struct {
	EmployeeID, CandidateID, StartDate, Notes string
	Tasks                                     []string
}, actorID, ip, ua string) (*models.Onboarding, error) {
	if _, err := s.employees.Get(tenantID, in.EmployeeID); err != nil {
		return nil, ErrNotFound
	}
	start, err := time.Parse("2006-01-02", in.StartDate)
	if err != nil {
		return nil, errors.New("invalid start_date (expected YYYY-MM-DD)")
	}
	if in.CandidateID != "" {
		if _, err := s.candidates.FindByID(tenantID, in.CandidateID); err != nil {
			return nil, ErrNotFound
		}
	}
	o := &models.Onboarding{
		TenantID:    tenantID,
		EmployeeID:  in.EmployeeID,
		CandidateID: stringPtr(in.CandidateID),
		StartDate:   datatypes.Date(start),
		Status:      models.OnboardingPending,
		Tasks:       tasksJSON(in.Tasks),
		Notes:       in.Notes,
	}
	if err := s.onboard.Create(o); err != nil {
		return nil, err
	}
	s.audit.Record(actorID, models.ActionCreate, "onboarding", o.ID, ip, ua, map[string]string{"employee_id": in.EmployeeID})
	return o, nil
}

func (s *RecruitmentService) UpdateOnboarding(tenantID, id, status, notes string, tasks []string, actorID, ip, ua string) (*models.Onboarding, error) {
	if _, err := s.onboard.FindByID(tenantID, id); err != nil {
		return nil, ErrNotFound
	}
	fields := map[string]interface{}{}
	if status != "" {
		fields["status"] = status
	}
	if notes != "" {
		fields["notes"] = notes
	}
	if tasks != nil {
		fields["tasks"] = tasksJSON(tasks)
	}
	if err := s.onboard.Update(tenantID, id, fields); err != nil {
		return nil, err
	}
	s.audit.Record(actorID, models.ActionUpdate, "onboarding", id, ip, ua, map[string]string{"status": status})
	return s.onboard.FindByID(tenantID, id)
}

func (s *RecruitmentService) Onboarding(tenantID, id string) (*models.Onboarding, error) {
	o, err := s.onboard.FindByID(tenantID, id)
	if err != nil {
		return nil, ErrNotFound
	}
	return o, nil
}

func (s *RecruitmentService) Onboardings(tenantID string, p utils.Pagination, employeeID, status string) ([]models.Onboarding, int64, error) {
	return s.onboard.List(tenantID, p, employeeID, status)
}

// --- Hiring ---

type HireInput struct {
	ApplicationID, EmployeeCode, DepartmentID, DesignationID, ManagerID, JoiningDate string
	CreateUser                                                                       bool
}

func (s *RecruitmentService) Hire(tenantID string, in HireInput, actorID, ip, ua string) (*models.Employee, error) {
	app, err := s.apps.FindByID(tenantID, in.ApplicationID)
	if err != nil || app.Candidate == nil {
		return nil, ErrNotFound
	}
	cand := app.Candidate
	if s.candidates.AlreadyEmployeeEmail(tenantID, cand.Email) {
		return nil, ErrCandidateAlreadyEmployee
	}
	code := in.EmployeeCode
	if code == "" {
		code = s.nextEmployeeCode()
	}
	emp, err := s.employees.Create(tenantID, EmployeeInput{
		EmployeeCode:  code,
		FirstName:     cand.FirstName,
		LastName:      cand.LastName,
		Email:         cand.Email,
		Phone:         cand.Phone,
		JoiningDate:   in.JoiningDate,
		DepartmentID:  in.DepartmentID,
		DesignationID: in.DesignationID,
		ManagerID:     in.ManagerID,
		Status:        string(models.EmployeeOnboardingProceed),
	}, actorID, ip, ua)
	if err != nil {
		return nil, err
	}
	_ = s.apps.Update(tenantID, app.ID, map[string]interface{}{"status": models.AppHired})
	_ = s.candidates.Update(tenantID, cand.ID, map[string]interface{}{"status": models.CandidateHired})
	start := time.Now()
	if in.JoiningDate != "" {
		if d, err := time.Parse("2006-01-02", in.JoiningDate); err == nil {
			start = d
		}
	}
	o := &models.Onboarding{
		TenantID:    tenantID,
		EmployeeID:  emp.ID,
		CandidateID: &cand.ID,
		StartDate:   datatypes.Date(start),
		Status:      models.OnboardingPending,
	}
	if err := s.onboard.Create(o); err != nil {
		return nil, err
	}
	if app.JobPost != nil && app.JobPost.Vacancies > 0 {
		_ = s.jobs.Update(tenantID, app.JobPost.ID, map[string]interface{}{"vacancies": app.JobPost.Vacancies - 1})
	}
	s.audit.Record(actorID, models.ActionCreate, "employee", emp.ID, ip, ua, map[string]string{"hired_from_candidate": cand.ID})
	s.audit.Record(actorID, models.ActionUpdate, "candidate", cand.ID, ip, ua, map[string]string{"status": "HIRED"})
	return emp, nil
}

func (s *RecruitmentService) nextEmployeeCode() string {
	prefix := "EMP"
	t := time.Now()
	return fmt.Sprintf("%s%d", prefix, t.UnixNano()%1000000)
}

func tasksJSON(tasks []string) datatypes.JSON {
	b, _ := json.Marshal(tasks)
	return datatypes.JSON(b)
}
