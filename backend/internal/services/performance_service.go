package services

import (
	"time"

	"gorm.io/datatypes"

	"github.com/emplyra/backend/internal/auditmanager"
	"github.com/emplyra/backend/internal/models"
	"github.com/emplyra/backend/internal/repositories"
	"github.com/emplyra/backend/internal/utils"
)

type PerformanceService struct {
	goals   *repositories.GoalRepository
	kpis    *repositories.KpiRepository
	reviews *repositories.ReviewRepository
	emp     *repositories.EmployeeRepository
	audit   *auditmanager.Service
}

func NewPerformanceService(goals *repositories.GoalRepository, kpis *repositories.KpiRepository, reviews *repositories.ReviewRepository, emp *repositories.EmployeeRepository, audit *auditmanager.Service) *PerformanceService {
	return &PerformanceService{goals: goals, kpis: kpis, reviews: reviews, emp: emp, audit: audit}
}

// --- Goals ---

type GoalInput struct {
	EmployeeID, Title, Description, TargetDate, Status string
	Weight                                             int
}

func (s *PerformanceService) CreateGoal(tenantID string, in GoalInput, actorID, ip, ua string) (*models.Goal, error) {
	if _, err := s.emp.FindByID(tenantID, in.EmployeeID); err != nil {
		return nil, ErrNotFound
	}
	g := &models.Goal{
		TenantID:   tenantID,
		EmployeeID: in.EmployeeID, Title: in.Title, Description: in.Description,
		Status: models.GoalStatus(orStr(in.Status, string(models.GoalNotStarted))),
		Weight: in.Weight,
	}
	if in.TargetDate != "" {
		if d, err := time.Parse("2006-01-02", in.TargetDate); err == nil {
			dd := datatypes.Date(d)
			g.TargetDate = &dd
		}
	}
	if err := s.goals.Create(g); err != nil {
		return nil, err
	}
	s.audit.Record(actorID, models.ActionCreate, "goal", g.ID, ip, ua, map[string]string{"employee_id": in.EmployeeID})
	return g, nil
}

func (s *PerformanceService) UpdateGoal(tenantID, id string, in GoalInput, actorID, ip, ua string) (*models.Goal, error) {
	if _, err := s.goals.FindByID(tenantID, id); err != nil {
		return nil, ErrNotFound
	}
	fields := map[string]interface{}{}
	if in.Title != "" {
		fields["title"] = in.Title
	}
	if in.Description != "" {
		fields["description"] = in.Description
	}
	if in.Status != "" {
		fields["status"] = in.Status
	}
	if in.Weight != 0 {
		fields["weight"] = in.Weight
	}
	if in.TargetDate != "" {
		if d, err := time.Parse("2006-01-02", in.TargetDate); err == nil {
			fields["target_date"] = datatypes.Date(d)
		}
	}
	if err := s.goals.Update(tenantID, id, fields); err != nil {
		return nil, err
	}
	s.audit.Record(actorID, models.ActionUpdate, "goal", id, ip, ua, nil)
	return s.goals.FindByID(tenantID, id)
}

func (s *PerformanceService) DeleteGoal(tenantID, id, actorID, ip, ua string) error {
	if _, err := s.goals.FindByID(tenantID, id); err != nil {
		return ErrNotFound
	}
	if err := s.goals.Delete(tenantID, id); err != nil {
		return err
	}
	s.audit.Record(actorID, models.ActionDelete, "goal", id, ip, ua, nil)
	return nil
}

func (s *PerformanceService) Goal(tenantID, id string) (*models.Goal, error) {
	g, err := s.goals.FindByID(tenantID, id)
	if err != nil {
		return nil, ErrNotFound
	}
	return g, nil
}

func (s *PerformanceService) Goals(tenantID string, p utils.Pagination, employeeIDs []string, status string) ([]models.Goal, int64, error) {
	return s.goals.List(tenantID, p, employeeIDs, status)
}

// --- KPIs ---

type KPIInput struct {
	EmployeeID, Name, Description, Target, Actual, Unit, Period string
	Weight                                                      int
	Score                                                       *float64
}

func (s *PerformanceService) CreateKPI(tenantID string, in KPIInput, actorID, ip, ua string) (*models.KPI, error) {
	if _, err := s.emp.FindByID(tenantID, in.EmployeeID); err != nil {
		return nil, ErrNotFound
	}
	k := &models.KPI{
		TenantID:   tenantID,
		EmployeeID: in.EmployeeID, Name: in.Name, Description: in.Description,
		Target: in.Target, Actual: in.Actual, Unit: in.Unit, Weight: in.Weight,
		Period: in.Period, Score: in.Score,
	}
	if err := s.kpis.Create(k); err != nil {
		return nil, err
	}
	s.audit.Record(actorID, models.ActionCreate, "kpi", k.ID, ip, ua, map[string]string{"employee_id": in.EmployeeID})
	return k, nil
}

func (s *PerformanceService) UpdateKPI(tenantID, id string, in KPIInput, actorID, ip, ua string) (*models.KPI, error) {
	if _, err := s.kpis.FindByID(tenantID, id); err != nil {
		return nil, ErrNotFound
	}
	fields := map[string]interface{}{}
	if in.Name != "" {
		fields["name"] = in.Name
	}
	if in.Description != "" {
		fields["description"] = in.Description
	}
	if in.Target != "" {
		fields["target"] = in.Target
	}
	if in.Actual != "" {
		fields["actual"] = in.Actual
	}
	if in.Unit != "" {
		fields["unit"] = in.Unit
	}
	if in.Weight != 0 {
		fields["weight"] = in.Weight
	}
	if in.Period != "" {
		fields["period"] = in.Period
	}
	if in.Score != nil {
		fields["score"] = *in.Score
	}
	if err := s.kpis.Update(tenantID, id, fields); err != nil {
		return nil, err
	}
	s.audit.Record(actorID, models.ActionUpdate, "kpi", id, ip, ua, nil)
	return s.kpis.FindByID(tenantID, id)
}

func (s *PerformanceService) DeleteKPI(tenantID, id, actorID, ip, ua string) error {
	if _, err := s.kpis.FindByID(tenantID, id); err != nil {
		return ErrNotFound
	}
	if err := s.kpis.Delete(tenantID, id); err != nil {
		return err
	}
	s.audit.Record(actorID, models.ActionDelete, "kpi", id, ip, ua, nil)
	return nil
}

func (s *PerformanceService) KPI(tenantID, id string) (*models.KPI, error) {
	k, err := s.kpis.FindByID(tenantID, id)
	if err != nil {
		return nil, ErrNotFound
	}
	return k, nil
}

func (s *PerformanceService) KPIs(tenantID string, p utils.Pagination, employeeIDs []string, period string) ([]models.KPI, int64, error) {
	return s.kpis.List(tenantID, p, employeeIDs, period)
}

// --- Reviews ---

type ReviewInput struct {
	EmployeeID, ReviewerID, Period, DueDate string
}

func (s *PerformanceService) CreateReview(tenantID string, in ReviewInput, actorID, ip, ua string) (*models.PerformanceReview, error) {
	if _, err := s.emp.FindByID(tenantID, in.EmployeeID); err != nil {
		return nil, ErrNotFound
	}
	r := &models.PerformanceReview{
		TenantID:   tenantID,
		EmployeeID: in.EmployeeID, Period: in.Period,
		Status: models.ReviewPending,
	}
	if in.ReviewerID != "" {
		r.ReviewerID = stringPtr(in.ReviewerID)
	}
	if in.DueDate != "" {
		if d, err := time.Parse("2006-01-02", in.DueDate); err == nil {
			dd := datatypes.Date(d)
			r.DueDate = &dd
		}
	}
	if err := s.reviews.Create(r); err != nil {
		return nil, err
	}
	s.audit.Record(actorID, models.ActionCreate, "performance_review", r.ID, ip, ua, map[string]string{"employee_id": in.EmployeeID})
	return r, nil
}

func (s *PerformanceService) SubmitReview(tenantID, id, selfEval, managerFeedback, status string, score *float64, actorID, ip, ua string) (*models.PerformanceReview, error) {
	r, err := s.reviews.FindByID(tenantID, id)
	if err != nil {
		return nil, ErrNotFound
	}
	fields := map[string]interface{}{}
	if selfEval != "" {
		fields["self_evaluation"] = selfEval
		if r.Status == models.ReviewPending {
			fields["status"] = models.ReviewSelfSubmitted
		}
	}
	if managerFeedback != "" {
		fields["manager_feedback"] = managerFeedback
		fields["status"] = models.ReviewManagerDone
		if r.SelfEvaluation != "" {
			fields["status"] = models.ReviewCompleted
		}
	}
	if score != nil {
		fields["score"] = *score
		if r.SelfEvaluation != "" || managerFeedback != "" {
			fields["status"] = models.ReviewCompleted
		}
	}
	if status != "" {
		fields["status"] = status
	}
	if len(fields) == 0 {
		return r, nil
	}
	if err := s.reviews.Update(tenantID, id, fields); err != nil {
		return nil, err
	}
	now := time.Now()
	if ns, ok := fields["status"]; ok && ns == models.ReviewCompleted {
		_ = s.reviews.Update(tenantID, id, map[string]interface{}{"completed_at": &now})
	}
	s.audit.Record(actorID, models.ActionUpdate, "performance_review", id, ip, ua, nil)
	return s.reviews.FindByID(tenantID, id)
}

func (s *PerformanceService) Review(tenantID, id string) (*models.PerformanceReview, error) {
	r, err := s.reviews.FindByID(tenantID, id)
	if err != nil {
		return nil, ErrNotFound
	}
	return r, nil
}

func (s *PerformanceService) Reviews(tenantID string, p utils.Pagination, employeeIDs []string, reviewerID, status string) ([]models.PerformanceReview, int64, error) {
	return s.reviews.List(tenantID, p, employeeIDs, reviewerID, status)
}
