package services

import (
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/emplyra/backend/internal/auditmanager"
	"github.com/emplyra/backend/internal/models"
	"github.com/emplyra/backend/internal/repositories"
	"github.com/emplyra/backend/internal/utils"
)

type EmployeeService struct {
	employees    *repositories.EmployeeRepository
	departments  *repositories.DepartmentRepository
	designations *repositories.DesignationRepository
	audit        *auditmanager.Service
}

func NewEmployeeService(employees *repositories.EmployeeRepository, departments *repositories.DepartmentRepository, designations *repositories.DesignationRepository, audit *auditmanager.Service) *EmployeeService {
	return &EmployeeService{employees: employees, departments: departments, designations: designations, audit: audit}
}

type EmployeeInput struct {
	EmployeeCode     string
	FirstName        string
	LastName         string
	Email            string
	Phone            string
	DateOfBirth      string
	Gender           string
	Address          string
	EmergencyContact string
	JoiningDate      string
	EmploymentType   string
	Status           string
	DepartmentID     string
	DesignationID    string
	ManagerID        string
	UserID           string
}

func (s *EmployeeService) List(tenantID string, p utils.Pagination, params struct {
	Search, DepartmentID, DesignationID, ManagerID, Status, EmploymentType, SortBy, SortOrder string
}) ([]models.Employee, int64, error) {
	filter := func(q *gorm.DB) *gorm.DB {
		if params.Search != "" {
			like := "%" + strings.ToLower(params.Search) + "%"
			q = q.Where("LOWER(first_name) LIKE ? OR LOWER(last_name) LIKE ? OR LOWER(email) LIKE ? OR LOWER(employee_code) LIKE ?", like, like, like, like)
		}
		if params.DepartmentID != "" {
			q = q.Where("department_id = ?", params.DepartmentID)
		}
		if params.DesignationID != "" {
			q = q.Where("designation_id = ?", params.DesignationID)
		}
		if params.ManagerID != "" {
			q = q.Where("manager_id = ?", params.ManagerID)
		}
		if params.Status != "" {
			q = q.Where("status = ?", params.Status)
		}
		if params.EmploymentType != "" {
			q = q.Where("employment_type = ?", params.EmploymentType)
		}
		return q
	}
	return s.employees.List(tenantID, p, filter, params.SortBy, params.SortOrder)
}

func (s *EmployeeService) Get(tenantID, id string) (*models.Employee, error) {
	e, err := s.employees.FindByID(tenantID, id)
	if err != nil {
		return nil, ErrNotFound
	}
	return e, nil
}

func (s *EmployeeService) GetByUserID(userID string) (*models.Employee, error) {
	e, err := s.employees.FindByUserID(userID)
	if err != nil {
		return nil, ErrNotFound
	}
	return e, nil
}

func (s *EmployeeService) validateAssignment(tenantID string, fields map[string]interface{}, excludeID string) error {
	if code, ok := fields["employee_code"].(string); ok && code != "" {
		if s.employees.ExistsByCodeExcluding(tenantID, code, excludeID) {
			return ErrDuplicate
		}
	}
	if email, ok := fields["email"].(string); ok && email != "" {
		if s.employees.ExistsByEmailExcluding(tenantID, email, excludeID) {
			return ErrDuplicate
		}
	}
	if depID, ok := fields["department_id"].(string); ok && depID != "" {
		if _, err := s.departments.FindByID(tenantID, depID); err != nil {
			return ErrNotFound
		}
	}
	if desID, ok := fields["designation_id"].(string); ok && desID != "" {
		if _, err := s.designations.FindByID(tenantID, desID); err != nil {
			return ErrNotFound
		}
	}
	return nil
}

func (s *EmployeeService) Create(tenantID string, in EmployeeInput, actorID, ip, ua string) (*models.Employee, error) {
	fields, err := buildEmployeeFields(in, true)
	if err != nil {
		return nil, err
	}
	if err := s.validateAssignment(tenantID, fields, ""); err != nil {
		return nil, err
	}
	emp := &models.Employee{
		TenantID:         tenantID,
		EmployeeCode:     fields["employee_code"].(string),
		FirstName:        fields["first_name"].(string),
		LastName:         fields["last_name"].(string),
		Email:            strings.ToLower(strings.TrimSpace(in.Email)),
		Phone:            in.Phone,
		Address:          in.Address,
		EmergencyContact: in.EmergencyContact,
		EmploymentType:   models.EmploymentType(fields["employment_type"].(string)),
		Status:           models.EmployeeStatus(fields["status"].(string)),
	}
	if v, ok := fields["date_of_birth"].(*time.Time); ok {
		emp.DateOfBirth = v
	}
	if v, ok := fields["joining_date"].(*time.Time); ok {
		emp.JoiningDate = v
	}
	if in.Gender != "" {
		g := models.Gender(in.Gender)
		emp.Gender = &g
	}
	if in.DepartmentID != "" {
		emp.DepartmentID = &in.DepartmentID
	}
	if in.DesignationID != "" {
		emp.DesignationID = &in.DesignationID
	}
	if in.ManagerID != "" {
		emp.ManagerID = &in.ManagerID
	}
	if in.UserID != "" {
		emp.UserID = &in.UserID
	}
	if err := s.employees.Create(emp); err != nil {
		return nil, err
	}
	s.audit.Record(actorID, models.ActionCreate, "employee", emp.ID, ip, ua, map[string]string{"code": emp.EmployeeCode, "email": emp.Email})
	return emp, nil
}

func (s *EmployeeService) Update(tenantID, id string, in EmployeeInput, actorID, ip, ua string) (*models.Employee, error) {
	if _, err := s.employees.FindByID(tenantID, id); err != nil {
		return nil, ErrNotFound
	}
	fields, err := buildEmployeeFields(in, false)
	if err != nil {
		return nil, err
	}
	if in.Gender != "" {
		fields["gender"] = in.Gender
	}
	if len(fields) == 0 {
		return s.employees.FindByID(tenantID, id)
	}
	if err := s.validateAssignment(tenantID, fields, id); err != nil {
		return nil, err
	}
	if err := s.employees.Update(tenantID, id, fields); err != nil {
		return nil, err
	}
	s.audit.Record(actorID, models.ActionUpdate, "employee", id, ip, ua, nil)
	return s.employees.FindByID(tenantID, id)
}

func (s *EmployeeService) Delete(tenantID, id, actorID, ip, ua string) error {
	if _, err := s.employees.FindByID(tenantID, id); err != nil {
		return ErrNotFound
	}
	if err := s.employees.Delete(tenantID, id); err != nil {
		return err
	}
	s.audit.Record(actorID, models.ActionDelete, "employee", id, ip, ua, nil)
	return nil
}

func buildEmployeeFields(in EmployeeInput, isCreate bool) (map[string]interface{}, error) {
	fields := map[string]interface{}{}
	if isCreate || in.EmployeeCode != "" {
		fields["employee_code"] = strings.ToUpper(strings.TrimSpace(in.EmployeeCode))
	}
	if isCreate || in.FirstName != "" {
		fields["first_name"] = in.FirstName
	}
	if isCreate || in.LastName != "" {
		fields["last_name"] = in.LastName
	}
	if isCreate || in.Email != "" {
		fields["email"] = strings.ToLower(strings.TrimSpace(in.Email))
	}
	if in.Phone != "" {
		fields["phone"] = in.Phone
	}
	if in.DateOfBirth != "" {
		t, err := parseDate(in.DateOfBirth)
		if err != nil {
			return nil, err
		}
		fields["date_of_birth"] = t
	}
	if in.JoiningDate != "" {
		t, err := parseDate(in.JoiningDate)
		if err != nil {
			return nil, err
		}
		fields["joining_date"] = t
	}
	if in.Address != "" {
		fields["address"] = in.Address
	}
	if in.EmergencyContact != "" {
		fields["emergency_contact"] = in.EmergencyContact
	}
	if isCreate {
		fields["employment_type"] = orStr(in.EmploymentType, string(models.EmploymentFullTime))
		fields["status"] = orStr(in.Status, string(models.EmployeeActive))
	} else {
		if in.EmploymentType != "" {
			fields["employment_type"] = in.EmploymentType
		}
		if in.Status != "" {
			fields["status"] = in.Status
		}
	}
	if in.DepartmentID != "" {
		fields["department_id"] = in.DepartmentID
	}
	if in.DesignationID != "" {
		fields["designation_id"] = in.DesignationID
	}
	if in.ManagerID != "" {
		fields["manager_id"] = in.ManagerID
	}
	if in.UserID != "" {
		fields["user_id"] = in.UserID
	}
	return fields, nil
}

func orStr(s, def string) string {
	if s == "" {
		return def
	}
	return s
}
