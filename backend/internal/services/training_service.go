package services

import (
	"errors"
	"time"

	"gorm.io/datatypes"

	"github.com/emplyra/backend/internal/auditmanager"
	"github.com/emplyra/backend/internal/models"
	"github.com/emplyra/backend/internal/notifications"
	"github.com/emplyra/backend/internal/repositories"
	"github.com/emplyra/backend/internal/utils"
)

type TrainingService struct {
	programs *repositories.TrainingRepository
	scheds   *repositories.TrainingScheduleRepository
	enroll   *repositories.EnrollmentRepository
	emp      *repositories.EmployeeRepository
	notify   *notifications.Service
	audit    *auditmanager.Service
}

func NewTrainingService(programs *repositories.TrainingRepository, scheds *repositories.TrainingScheduleRepository, enroll *repositories.EnrollmentRepository, emp *repositories.EmployeeRepository, notify *notifications.Service, audit *auditmanager.Service) *TrainingService {
	return &TrainingService{programs: programs, scheds: scheds, enroll: enroll, emp: emp, notify: notify, audit: audit}
}

type ProgramInput struct {
	Title, Description, Provider, StartDate, EndDate, Location, Status string
	MaxSeats                                                           int
}

func (s *TrainingService) CreateProgram(tenantID string, in ProgramInput, actorID, ip, ua string) (*models.TrainingProgram, error) {
	start, err := time.Parse("2006-01-02", in.StartDate)
	if err != nil {
		return nil, errors.New("invalid start_date (expected YYYY-MM-DD)")
	}
	end, err := time.Parse("2006-01-02", in.EndDate)
	if err != nil {
		return nil, errors.New("invalid end_date (expected YYYY-MM-DD)")
	}
	p := &models.TrainingProgram{
		TenantID: tenantID,
		Title:    in.Title, Description: in.Description, Provider: in.Provider,
		StartDate: datatypes.Date(start), EndDate: datatypes.Date(end),
		Location: in.Location, MaxSeats: in.MaxSeats,
		Status: models.TrainingStatus(orStr(in.Status, string(models.TrainingScheduled))),
	}
	if err := s.programs.Create(p); err != nil {
		return nil, err
	}
	s.audit.Record(actorID, models.ActionCreate, "training_program", p.ID, ip, ua, map[string]string{"title": p.Title})
	return p, nil
}

func (s *TrainingService) UpdateProgram(tenantID, id string, in ProgramInput, actorID, ip, ua string) (*models.TrainingProgram, error) {
	if _, err := s.programs.FindByID(tenantID, id); err != nil {
		return nil, ErrNotFound
	}
	fields := map[string]interface{}{}
	if in.Title != "" {
		fields["title"] = in.Title
	}
	if in.Description != "" {
		fields["description"] = in.Description
	}
	if in.Provider != "" {
		fields["provider"] = in.Provider
	}
	if in.Location != "" {
		fields["location"] = in.Location
	}
	if in.Status != "" {
		fields["status"] = in.Status
	}
	if in.MaxSeats != 0 {
		fields["max_seats"] = in.MaxSeats
	}
	set := func(key, val string) error {
		if val == "" {
			return nil
		}
		d, err := time.Parse("2006-01-02", val)
		if err != nil {
			return errors.New("invalid date (expected YYYY-MM-DD)")
		}
		fields[key] = datatypes.Date(d)
		return nil
	}
	if err := set("start_date", in.StartDate); err != nil {
		return nil, err
	}
	if err := set("end_date", in.EndDate); err != nil {
		return nil, err
	}
	if err := s.programs.Update(tenantID, id, fields); err != nil {
		return nil, err
	}
	s.audit.Record(actorID, models.ActionUpdate, "training_program", id, ip, ua, nil)
	return s.programs.FindByID(tenantID, id)
}

func (s *TrainingService) DeleteProgram(tenantID, id, actorID, ip, ua string) error {
	if _, err := s.programs.FindByID(tenantID, id); err != nil {
		return ErrNotFound
	}
	if err := s.programs.Delete(tenantID, id); err != nil {
		return err
	}
	s.audit.Record(actorID, models.ActionDelete, "training_program", id, ip, ua, nil)
	return nil
}

func (s *TrainingService) Program(tenantID, id string) (*models.TrainingProgram, error) {
	p, err := s.programs.FindByID(tenantID, id)
	if err != nil {
		return nil, ErrNotFound
	}
	return p, nil
}

func (s *TrainingService) Programs(tenantID string, p utils.Pagination, status string) ([]models.TrainingProgram, int64, error) {
	return s.programs.List(tenantID, p, status)
}

// --- Schedules ---

func (s *TrainingService) CreateSchedule(tenantID string, in struct {
	ProgramID, Date, StartTime, EndTime, Trainer, Location string
	MaxSeats                                               int
}, actorID, ip, ua string) (*models.TrainingSchedule, error) {
	if _, err := s.programs.FindByID(tenantID, in.ProgramID); err != nil {
		return nil, ErrNotFound
	}
	d, err := time.Parse("2006-01-02", in.Date)
	if err != nil {
		return nil, errors.New("invalid date (expected YYYY-MM-DD)")
	}
	sc := &models.TrainingSchedule{
		TenantID:  tenantID,
		ProgramID: in.ProgramID, Date: datatypes.Date(d),
		StartTime: in.StartTime, EndTime: in.EndTime,
		Trainer: in.Trainer, Location: in.Location, MaxSeats: in.MaxSeats,
	}
	if err := s.scheds.Create(sc); err != nil {
		return nil, err
	}
	s.audit.Record(actorID, models.ActionCreate, "training_schedule", sc.ID, ip, ua, nil)
	return sc, nil
}

func (s *TrainingService) UpdateSchedule(tenantID, id string, in struct {
	ProgramID, Date, StartTime, EndTime, Trainer, Location string
	MaxSeats                                               int
}, actorID, ip, ua string) (*models.TrainingSchedule, error) {
	if _, err := s.scheds.FindByID(tenantID, id); err != nil {
		return nil, ErrNotFound
	}
	fields := map[string]interface{}{}
	if in.ProgramID != "" {
		fields["program_id"] = in.ProgramID
	}
	if in.Date != "" {
		if d, err := time.Parse("2006-01-02", in.Date); err == nil {
			fields["date"] = datatypes.Date(d)
		}
	}
	if in.StartTime != "" {
		fields["start_time"] = in.StartTime
	}
	if in.EndTime != "" {
		fields["end_time"] = in.EndTime
	}
	if in.Trainer != "" {
		fields["trainer"] = in.Trainer
	}
	if in.Location != "" {
		fields["location"] = in.Location
	}
	if in.MaxSeats != 0 {
		fields["max_seats"] = in.MaxSeats
	}
	if err := s.scheds.Update(tenantID, id, fields); err != nil {
		return nil, err
	}
	s.audit.Record(actorID, models.ActionUpdate, "training_schedule", id, ip, ua, nil)
	return s.scheds.FindByID(tenantID, id)
}

func (s *TrainingService) DeleteSchedule(tenantID, id, actorID, ip, ua string) error {
	if _, err := s.scheds.FindByID(tenantID, id); err != nil {
		return ErrNotFound
	}
	if err := s.scheds.Delete(tenantID, id); err != nil {
		return err
	}
	s.audit.Record(actorID, models.ActionDelete, "training_schedule", id, ip, ua, nil)
	return nil
}

func (s *TrainingService) Schedules(tenantID, programID string) ([]models.TrainingSchedule, error) {
	return s.scheds.ListByProgram(tenantID, programID)
}

// --- Enrollments ---

func (s *TrainingService) Enroll(tenantID, programID, employeeID string, actorID, ip, ua string) (*models.TrainingEnrollment, error) {
	if _, err := s.programs.FindByID(tenantID, programID); err != nil {
		return nil, ErrNotFound
	}
	if _, err := s.emp.FindByID(tenantID, employeeID); err != nil {
		return nil, ErrNotFound
	}
	exists, err := s.enroll.Exists(tenantID, programID, employeeID)
	if err != nil || exists {
		return nil, ErrEnrollmentDuplicate
	}
	e := &models.TrainingEnrollment{
		TenantID:  tenantID,
		ProgramID: programID, EmployeeID: employeeID,
		Status: models.EnrollmentEnrolled, EnrolledAt: time.Now(),
	}
	if err := s.enroll.Create(e); err != nil {
		return nil, err
	}
	if ep, err := s.emp.FindByID(tenantID, employeeID); err == nil && ep.UserID != nil {
		_ = s.notify.Notify(*ep.UserID, "Training enrollment",
			"You have been enrolled in a training program.", models.NotifyTraining, "/training", nil)
	}
	s.audit.Record(actorID, models.ActionCreate, "training_enrollment", e.ID, ip, ua, map[string]string{"program_id": programID, "employee_id": employeeID})
	return e, nil
}

func (s *TrainingService) UpdateEnrollment(tenantID, id, status string, actorID, ip, ua string) (*models.TrainingEnrollment, error) {
	if _, err := s.enroll.FindByID(tenantID, id); err != nil {
		return nil, ErrNotFound
	}
	fields := map[string]interface{}{"status": status}
	if status == string(models.EnrollmentCompleted) {
		now := time.Now()
		fields["completed_at"] = &now
	}
	if err := s.enroll.Update(tenantID, id, fields); err != nil {
		return nil, err
	}
	s.audit.Record(actorID, models.ActionUpdate, "training_enrollment", id, ip, ua, map[string]string{"status": status})
	return s.enroll.FindByID(tenantID, id)
}

func (s *TrainingService) Enrollments(tenantID string, p utils.Pagination, programID, employeeID, status string) ([]models.TrainingEnrollment, int64, error) {
	return s.enroll.List(tenantID, p, programID, employeeID, status)
}
