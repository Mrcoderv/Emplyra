package services

import (
	"errors"
	"time"

	"gorm.io/datatypes"

	"github.com/emplyra/backend/internal/auditmanager"
	"github.com/emplyra/backend/internal/models"
	"github.com/emplyra/backend/internal/repositories"
	"github.com/emplyra/backend/internal/utils"
)

const (
	workStartTime = 9 * time.Hour // 09:00
	lateGrace     = 15 * time.Minute
	officeHours   = 9.0 // paid working hours per day used for overtime calc
)

type AttendanceService struct {
	att   *repositories.AttendanceRepository
	emp   *repositories.EmployeeRepository
	audit *auditmanager.Service
}

func NewAttendanceService(att *repositories.AttendanceRepository, emp *repositories.EmployeeRepository, audit *auditmanager.Service) *AttendanceService {
	return &AttendanceService{att: att, emp: emp, audit: audit}
}

func (s *AttendanceService) CheckIn(employeeID, remarks string, actorID, ip, ua string) (*models.Attendance, error) {
	if err := s.employeesExist(employeeID); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	today := datatypes.Date(now)
	existing, err := s.att.FindToday(employeeID, now)
	if err == nil && existing != nil {
		return nil, ErrAlreadyCheckedIn
	}
	late, lateMin := computeLate(now)
	status := models.AttendancePresent
	if late {
		status = models.AttendanceLate
	}
	rec := &models.Attendance{
		EmployeeID:   employeeID,
		Date:         today,
		CheckIn:      &now,
		Status:       status,
		LateMinutes:  lateMin,
		WorkingHours: 0,
		Remarks:      remarks,
	}
	if err := s.att.Create(rec); err != nil {
		return nil, err
	}
	s.audit.Record(actorID, models.ActionCreate, "attendance", rec.ID, "", "", map[string]string{"employee_id": employeeID, "action": "check_in"})
	return rec, nil
}

func (s *AttendanceService) CheckOut(employeeID, remarks string, overtime float64, actorID, ip, ua string) (*models.Attendance, error) {
	now := time.Now().UTC()
	existing, err := s.att.FindToday(employeeID, now)
	if err != nil {
		return nil, ErrNoCheckIn
	}
	if existing.CheckOut != nil {
		return nil, errors.New("already checked out today")
	}
	checkOut := now
	hours := checkOut.Sub(*existing.CheckIn).Hours()
	if hours < 0 {
		return nil, errors.New("check-out time precedes check-in")
	}
	fields := map[string]interface{}{
		"check_out":     &checkOut,
		"working_hours": round2(hours),
		"remarks":       remarks,
	}
	if overtime == 0 {
		ot := hours - officeHours
		if ot > 0 {
			fields["overtime"] = round2(ot)
		}
	} else {
		fields["overtime"] = overtime
	}
	if existing.Status == models.AttendancePresent || existing.Status == models.AttendanceLate {
		if existing.CheckIn.Hour() >= 14 {
			fields["status"] = models.AttendanceHalfDay
		}
	}
	if err := s.att.Update(existing.ID, fields); err != nil {
		return nil, err
	}
	s.audit.Record(actorID, models.ActionUpdate, "attendance", existing.ID, "", "", map[string]string{"employee_id": employeeID, "action": "check_out"})
	return s.att.FindByID(existing.ID)
}

func (s *AttendanceService) Update(id string, in struct {
	CheckOut    *string
	CheckIn     *string
	Status      string
	LateMinutes int
	Overtime    float64
	Remarks     string
}, actorID, ip, ua string) (*models.Attendance, error) {
	existing, err := s.att.FindByID(id)
	if err != nil {
		return nil, ErrNotFound
	}
	fields := map[string]interface{}{}
	if in.CheckIn != nil {
		t, err := time.Parse(time.RFC3339, *in.CheckIn)
		if err != nil {
			return nil, errors.New("invalid check_in time (expected RFC3339)")
		}
		fields["check_in"] = &t
	}
	if in.CheckOut != nil {
		t, err := time.Parse(time.RFC3339, *in.CheckOut)
		if err != nil {
			return nil, errors.New("invalid check_out time (expected RFC3339)")
		}
		fields["check_out"] = &t
	}
	if in.Status != "" {
		fields["status"] = in.Status
	}
	if in.Remarks != "" {
		fields["remarks"] = in.Remarks
	}
	if in.LateMinutes != 0 {
		fields["late_minutes"] = in.LateMinutes
	}
	if in.Overtime != 0 {
		fields["overtime"] = in.Overtime
	}
	if ci, ok := fields["check_in"]; ok {
		co, ok2 := fields["check_out"]
		if ok2 {
			hours := co.(*time.Time).Sub(*ci.(*time.Time)).Hours()
			if hours < 0 {
				return nil, errors.New("check-out precedes check-in")
			}
			fields["working_hours"] = round2(hours)
		}
	}
	if len(fields) == 0 {
		return existing, nil
	}
	if err := s.att.Update(id, fields); err != nil {
		return nil, err
	}
	s.audit.Record(actorID, models.ActionUpdate, "attendance", id, ip, ua, nil)
	return s.att.FindByID(id)
}

func (s *AttendanceService) List(p utils.Pagination, employeeID, from, to string, status string) ([]models.Attendance, int64, error) {
	f, err := parseDate(from)
	if err != nil {
		return nil, 0, err
	}
	t, err := parseDate(to)
	if err != nil {
		return nil, 0, err
	}
	var dfrom, dto datatypes.Date
	if f != nil {
		dfrom = datatypes.Date(*f)
	}
	if t != nil {
		dto = datatypes.Date(*t)
	}
	return s.att.List(p, employeeID, dfrom, dto, status)
}

func (s *AttendanceService) Get(id string) (*models.Attendance, error) {
	a, err := s.att.FindByID(id)
	if err != nil {
		return nil, ErrNotFound
	}
	return a, nil
}

func (s *AttendanceService) employeesExist(id string) error {
	if _, err := s.emp.FindByID(id); err != nil {
		return errors.New("employee not found")
	}
	return nil
}

func (s *AttendanceService) EmployeeIDForUser(userID string) (string, error) {
	emp, err := s.emp.FindByUserID(userID)
	if err != nil {
		return "", errors.New("no employee profile linked to this user")
	}
	return emp.ID, nil
}

func computeLate(t time.Time) (bool, int) {
	local := t.Local()
	start := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, local.Location()).Add(workStartTime)
	grace := start.Add(lateGrace)
	if local.After(grace) {
		return true, int(local.Sub(start).Minutes())
	}
	return false, 0
}

func round2(f float64) float64 {
	return float64(int(f*100+0.5)) / 100
}
