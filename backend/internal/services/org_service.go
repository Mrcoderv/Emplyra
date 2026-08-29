package services

import (
	"github.com/emplyra/backend/internal/auditmanager"
	"github.com/emplyra/backend/internal/models"
	"github.com/emplyra/backend/internal/repositories"
)

type DepartmentService struct {
	repos *repositories.DepartmentRepository
	emp   *repositories.EmployeeRepository
	audit *auditmanager.Service
}

func NewDepartmentService(repos *repositories.DepartmentRepository, emp *repositories.EmployeeRepository, audit *auditmanager.Service) *DepartmentService {
	return &DepartmentService{repos: repos, emp: emp, audit: audit}
}

func (s *DepartmentService) validate(dept *models.Department) error {
	if existing, _ := s.repos.FindByCode(dept.Code); existing != nil && existing.ID != dept.ID {
		return ErrDuplicate
	}
	return nil
}

func (s *DepartmentService) List() ([]models.Department, error) {
	return s.repos.List(nil)
}

func (s *DepartmentService) Get(id string) (*models.Department, error) {
	d, err := s.repos.FindByID(id)
	if err != nil {
		return nil, ErrNotFound
	}
	return d, nil
}

func (s *DepartmentService) Create(in struct{ Name, Code, Description, ManagerID, Status string }, actorID, ip, ua string) (*models.Department, error) {
	dept := &models.Department{
		Name:        in.Name,
		Code:        in.Code,
		Description: in.Description,
		ManagerID:   stringPtr(in.ManagerID),
		Status:      models.OrgEntityStatus(in.Status),
	}
	if dept.Status == "" {
		dept.Status = models.OrgStatusActive
	}
	if err := s.validate(dept); err != nil {
		return nil, err
	}
	if err := s.repos.Create(dept); err != nil {
		return nil, err
	}
	s.audit.Record(actorID, models.ActionCreate, "department", dept.ID, ip, ua, map[string]string{"code": dept.Code})
	return dept, nil
}

func (s *DepartmentService) Update(id string, in struct{ Name, Code, Description, ManagerID, Status string }, actorID, ip, ua string) (*models.Department, error) {
	existing, err := s.repos.FindByID(id)
	if err != nil {
		return nil, ErrNotFound
	}
	fields := map[string]interface{}{}
	if in.Name != "" {
		fields["name"] = in.Name
	}
	if in.Code != "" {
		fields["code"] = in.Code
	}
	if in.Description != "" {
		fields["description"] = in.Description
	}
	if in.ManagerID != "" {
		fields["manager_id"] = in.ManagerID
	}
	if in.Status != "" {
		fields["status"] = in.Status
	}
	if code, ok := fields["code"]; ok {
		if dup, _ := s.repos.FindByCode(code.(string)); dup != nil && dup.ID != id {
			return nil, ErrDuplicate
		}
	}
	if len(fields) == 0 {
		return existing, nil
	}
	if err := s.repos.Update(id, fields); err != nil {
		return nil, err
	}
	s.audit.Record(actorID, models.ActionUpdate, "department", id, ip, ua, nil)
	return s.repos.FindByID(id)
}

func (s *DepartmentService) Delete(id, actorID, ip, ua string) error {
	if _, err := s.repos.FindByID(id); err != nil {
		return ErrNotFound
	}
	if err := s.repos.Delete(id); err != nil {
		return err
	}
	s.audit.Record(actorID, models.ActionDelete, "department", id, ip, ua, nil)
	return nil
}

type DesignationService struct {
	repos *repositories.DesignationRepository
	audit *auditmanager.Service
}

func NewDesignationService(repos *repositories.DesignationRepository, audit *auditmanager.Service) *DesignationService {
	return &DesignationService{repos: repos, audit: audit}
}

func (s *DesignationService) List() ([]models.Designation, error) {
	return s.repos.List()
}

func (s *DesignationService) Get(id string) (*models.Designation, error) {
	d, err := s.repos.FindByID(id)
	if err != nil {
		return nil, ErrNotFound
	}
	return d, nil
}

func (s *DesignationService) Create(in struct {
	Name, Description, DepartmentID, Status string
	Level                                   int
}, actorID, ip, ua string) (*models.Designation, error) {
	d := &models.Designation{
		Name:         in.Name,
		Description:  in.Description,
		DepartmentID: stringPtr(in.DepartmentID),
		Level:        in.Level,
		Status:       models.OrgEntityStatus(in.Status),
	}
	if d.Level == 0 {
		d.Level = 1
	}
	if d.Status == "" {
		d.Status = models.OrgStatusActive
	}
	if err := s.repos.Create(d); err != nil {
		return nil, err
	}
	s.audit.Record(actorID, models.ActionCreate, "designation", d.ID, ip, ua, map[string]string{"name": d.Name})
	return d, nil
}

func (s *DesignationService) Update(id string, in struct {
	Name, Description, DepartmentID, Status string
	Level                                   int
}, actorID, ip, ua string) (*models.Designation, error) {
	if _, err := s.repos.FindByID(id); err != nil {
		return nil, ErrNotFound
	}
	fields := map[string]interface{}{}
	if in.Name != "" {
		fields["name"] = in.Name
	}
	if in.Description != "" {
		fields["description"] = in.Description
	}
	if in.DepartmentID != "" {
		fields["department_id"] = in.DepartmentID
	}
	if in.Level != 0 {
		fields["level"] = in.Level
	}
	if in.Status != "" {
		fields["status"] = in.Status
	}
	if err := s.repos.Update(id, fields); err != nil {
		return nil, err
	}
	s.audit.Record(actorID, models.ActionUpdate, "designation", id, ip, ua, nil)
	return s.repos.FindByID(id)
}

func (s *DesignationService) Delete(id, actorID, ip, ua string) error {
	if _, err := s.repos.FindByID(id); err != nil {
		return ErrNotFound
	}
	if err := s.repos.Delete(id); err != nil {
		return err
	}
	s.audit.Record(actorID, models.ActionDelete, "designation", id, ip, ua, nil)
	return nil
}
