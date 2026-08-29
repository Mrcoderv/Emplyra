package services

import (
	"errors"
	"time"

	"gorm.io/datatypes"

	"github.com/emplyra/backend/internal/auditmanager"
	"github.com/emplyra/backend/internal/models"
	"github.com/emplyra/backend/internal/repositories"
)

type HolidayService struct {
	holidays *repositories.HolidayRepository
	audit    *auditmanager.Service
}

func NewHolidayService(holidays *repositories.HolidayRepository, audit *auditmanager.Service) *HolidayService {
	return &HolidayService{holidays: holidays, audit: audit}
}

func (s *HolidayService) Create(tenantID string, in struct{ Name, Date, Description, Type, Status string }, actorID, ip, ua string) (*models.Holiday, error) {
	d, err := time.Parse("2006-01-02", in.Date)
	if err != nil {
		return nil, errors.New("invalid date (expected YYYY-MM-DD)")
	}
	h := &models.Holiday{
		TenantID:    tenantID,
		Name:        in.Name,
		Date:        datatypes.Date(d),
		Description: in.Description,
		Type:        in.Type,
		Status:      models.HolidayStatus(orStr(in.Status, string(models.HolidayActive))),
	}
	if err := s.holidays.Create(h); err != nil {
		return nil, err
	}
	s.audit.Record(actorID, models.ActionCreate, "holiday", h.ID, ip, ua, map[string]string{"name": h.Name, "date": in.Date})
	return h, nil
}

func (s *HolidayService) Update(tenantID, id string, in struct{ Name, Date, Description, Type, Status string }, actorID, ip, ua string) (*models.Holiday, error) {
	if _, err := s.holidays.FindByID(tenantID, id); err != nil {
		return nil, ErrNotFound
	}
	fields := map[string]interface{}{}
	if in.Name != "" {
		fields["name"] = in.Name
	}
	if in.Date != "" {
		d, err := time.Parse("2006-01-02", in.Date)
		if err != nil {
			return nil, errors.New("invalid date (expected YYYY-MM-DD)")
		}
		fields["date"] = datatypes.Date(d)
	}
	if in.Description != "" {
		fields["description"] = in.Description
	}
	if in.Type != "" {
		fields["type"] = in.Type
	}
	if in.Status != "" {
		fields["status"] = in.Status
	}
	if err := s.holidays.Update(tenantID, id, fields); err != nil {
		return nil, err
	}
	s.audit.Record(actorID, models.ActionUpdate, "holiday", id, ip, ua, nil)
	return s.holidays.FindByID(tenantID, id)
}

func (s *HolidayService) Get(tenantID, id string) (*models.Holiday, error) {
	h, err := s.holidays.FindByID(tenantID, id)
	if err != nil {
		return nil, ErrNotFound
	}
	return h, nil
}

func (s *HolidayService) Delete(tenantID, id, actorID, ip, ua string) error {
	if _, err := s.holidays.FindByID(tenantID, id); err != nil {
		return ErrNotFound
	}
	if err := s.holidays.Delete(tenantID, id); err != nil {
		return err
	}
	s.audit.Record(actorID, models.ActionDelete, "holiday", id, ip, ua, nil)
	return nil
}

func (s *HolidayService) List(tenantID, from, to, status string) ([]models.Holiday, error) {
	f, err := parseDate(from)
	if err != nil {
		return nil, err
	}
	t, err := parseDate(to)
	if err != nil {
		return nil, err
	}
	var df, dt datatypes.Date
	if f != nil {
		df = datatypes.Date(*f)
	}
	if t != nil {
		dt = datatypes.Date(*t)
	}
	return s.holidays.List(tenantID, df, dt, status)
}
