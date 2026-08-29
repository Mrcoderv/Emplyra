package services

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/emplyra/backend/internal/auditmanager"
	"github.com/emplyra/backend/internal/models"
	"github.com/emplyra/backend/internal/notifications"
	"github.com/emplyra/backend/internal/repositories"
	"github.com/emplyra/backend/internal/utils"
)

var ErrLeaveAlreadyDecided = errors.New("leave request already decided")

type LeaveService struct {
	leaves   *repositories.LeaveRepository
	types    *repositories.LeaveTypeRepository
	balances *repositories.LeaveBalanceRepository
	emp      *repositories.EmployeeRepository
	notify   *notifications.Service
	audit    *auditmanager.Service
}

func NewLeaveService(leaves *repositories.LeaveRepository, types *repositories.LeaveTypeRepository, balances *repositories.LeaveBalanceRepository, emp *repositories.EmployeeRepository, notify *notifications.Service, audit *auditmanager.Service) *LeaveService {
	return &LeaveService{leaves: leaves, types: types, balances: balances, emp: emp, notify: notify, audit: audit}
}

func (s *LeaveService) Create(tenantID, employeeID, leaveTypeID, start, end, reason string, actorID, ip, ua string) (*models.Leave, error) {
	if _, err := s.emp.FindByID(tenantID, employeeID); err != nil {
		return nil, ErrNotFound
	}
	lt, err := s.types.FindByID(tenantID, leaveTypeID)
	if err != nil {
		return nil, ErrNotFound
	}
	st, err := parseDate(start)
	if err != nil {
		return nil, err
	}
	en, err := parseDate(end)
	if err != nil {
		return nil, err
	}
	if st.After(*en) {
		return nil, errors.New("start date cannot be after end date")
	}
	dstart := datatypes.Date(*st)
	dend := datatypes.Date(*en)
	overlaps, err := s.leaves.Overlaps(tenantID, employeeID, dstart, dend, "")
	if err == nil && len(overlaps) > 0 {
		return nil, ErrLeaveOverlap
	}
	days := businessDays(*st, *en)
	if days < 1 {
		return nil, errors.New("leave must include at least one working day")
	}
	if err := s.checkBalance(tenantID, employeeID, leaveTypeID, st.Year(), days); err != nil {
		return nil, err
	}
	leave := &models.Leave{
		TenantID:    tenantID,
		EmployeeID:  employeeID,
		LeaveTypeID: leaveTypeID,
		StartDate:   dstart,
		EndDate:     dend,
		Days:        days,
		Reason:      reason,
		Status:      models.LeavePending,
	}
	if err := s.leaves.Create(leave); err != nil {
		return nil, err
	}
	s.audit.Record(actorID, models.ActionCreate, "leave", leave.ID, ip, ua, map[string]string{"employee_id": employeeID, "days": fmt.Sprintf("%d", days)})
	s.notifyManager(tenantID, employeeID, lt)
	return leave, nil
}

func (s *LeaveService) checkBalance(tenantID, employeeID, leaveTypeID string, year, days int) error {
	bal, err := s.balances.Find(tenantID, employeeID, leaveTypeID, year)
	if err != nil {
		return nil // no configured balance -> allowed
	}
	remaining := bal.Entitlement - bal.Used
	if remaining < days {
		return ErrInsufficientLeaveBalance
	}
	return nil
}

func (s *LeaveService) List(tenantID string, p utils.Pagination, f repositories.LeaveFilter) ([]models.Leave, int64, error) {
	return s.leaves.List(tenantID, p, f)
}

func (s *LeaveService) Get(tenantID, id string) (*models.Leave, error) {
	l, err := s.leaves.FindByID(tenantID, id)
	if err != nil {
		return nil, ErrNotFound
	}
	return l, nil
}

func (s *LeaveService) Approve(tenantID, id, note string, actorID, ip, ua string) (*models.Leave, error) {
	return s.decide(tenantID, id, models.LeaveApproved, note, actorID, ip, ua)
}

func (s *LeaveService) Reject(tenantID, id, note string, actorID, ip, ua string) (*models.Leave, error) {
	return s.decide(tenantID, id, models.LeaveRejected, note, actorID, ip, ua)
}

func (s *LeaveService) decide(tenantID, id string, status models.LeaveStatus, note string, actorID, ip, ua string) (*models.Leave, error) {
	leave, err := s.leaves.FindByID(tenantID, id)
	if err != nil {
		return nil, ErrNotFound
	}
	if leave.Status != models.LeavePending {
		return nil, ErrLeaveAlreadyDecided
	}
	now := time.Now().UTC()
	fields := map[string]interface{}{
		"status":      status,
		"reviewer_id": actorID,
		"reviewed_at": &now,
		"review_note": note,
	}
	action := models.ActionReject
	if status == models.LeaveApproved {
		action = models.ActionApprove
	}
	err = s.leaves.Tx(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Leave{}).Scopes(repositories.TenantScope(tenantID)).Where("id = ?", id).Updates(fields).Error; err != nil {
			return err
		}
		if status == models.LeaveApproved && leave.LeaveType != nil && leave.LeaveType.IsPaid {
			return tx.Model(&models.LeaveBalance{}).
				Scopes(repositories.TenantScope(tenantID)).
				Where("employee_id = ? AND leave_type_id = ? AND year = ?", leave.EmployeeID, leave.LeaveTypeID, time.Time(leave.StartDate).Year()).
				UpdateColumn("used", gorm.Expr("used + ?", leave.Days)).Error
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.audit.Record(actorID, action, "leave", leave.ID, ip, ua, map[string]string{"employee_id": leave.EmployeeID})
	s.notifyEmployee(tenantID, leave, status, id)
	return s.leaves.FindByID(tenantID, id)
}

func (s *LeaveService) LeaveTypes(tenantID string) ([]models.LeaveType, error) {
	return s.types.List(tenantID)
}

func (s *LeaveService) SetBalance(tenantID, employeeID, leaveTypeID string, year, entitlement int, actorID, ip, ua string) (*models.LeaveBalance, error) {
	if year == 0 {
		year = time.Now().Year()
	}
	if err := s.balances.InitializeIfMissing(tenantID, employeeID, leaveTypeID, year, entitlement); err != nil {
		return nil, err
	}
	bal, err := s.balances.Find(tenantID, employeeID, leaveTypeID, year)
	if err != nil {
		return nil, ErrNotFound
	}
	bal.Entitlement = entitlement
	if err := s.balances.Upsert(bal); err != nil {
		return nil, err
	}
	s.audit.Record(actorID, models.ActionUpdate, "leave_balance", bal.ID, ip, ua, map[string]string{"entitlement": fmt.Sprintf("%d", entitlement)})
	return bal, nil
}

func (s *LeaveService) Balances(tenantID, employeeID string, year int) ([]models.LeaveBalance, error) {
	return s.balances.List(tenantID, employeeID, year)
}

func (s *LeaveService) notifyManager(tenantID, employeeID string, lt *models.LeaveType) {
	emp, err := s.emp.FindByID(tenantID, employeeID)
	if err != nil || emp.Manager == nil {
		return
	}
	if emp.Manager.UserID != nil {
		name := ltName(lt)
		_ = s.notify.Notify(*emp.Manager.UserID,
			"New leave request",
			fmt.Sprintf("%s applied for %s", emp.FullName(), name),
			models.NotifyLeaveRequest, "/leaves", nil)
	}
}

func (s *LeaveService) notifyEmployee(tenantID string, leave *models.Leave, status models.LeaveStatus, id string) {
	emp, err := s.emp.FindByID(tenantID, leave.EmployeeID)
	if err != nil || emp.User == nil || emp.UserID == nil {
		return
	}
	ntype := models.NotifyLeaveApproval
	title := "Leave approved"
	msg := fmt.Sprintf("Your leave (%s) has been approved.", leaveTypeDisplayName(leave))
	if status == models.LeaveRejected {
		ntype = models.NotifyLeaveRejection
		title = "Leave rejected"
		msg = fmt.Sprintf("Your leave (%s) has been rejected.", leaveTypeDisplayName(leave))
	}
	_ = s.notify.Notify(*emp.UserID, title, msg, ntype, "/leaves/"+id, nil)
}

func leaveTypeDisplayName(l *models.Leave) string {
	if l.LeaveType != nil {
		return l.LeaveType.Name
	}
	return "leave"
}

func ltName(lt *models.LeaveType) string {
	if lt == nil {
		return "leave"
	}
	return lt.Name
}

func businessDays(start, end time.Time) int {
	days := 0
	d := start
	for !d.After(end) {
		if wd := d.Weekday(); wd != time.Saturday && wd != time.Sunday {
			days++
		}
		d = d.AddDate(0, 0, 1)
	}
	return days
}
