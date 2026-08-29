package services

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/emplyra/backend/internal/auditmanager"
	"github.com/emplyra/backend/internal/models"
	"github.com/emplyra/backend/internal/repositories"
	"github.com/emplyra/backend/internal/utils"
)

var (
	ErrInvalidStatusTransition = errors.New("invalid payroll status transition")
	ErrPayrollExists           = errors.New("payroll already exists for this employee/month")
	ErrNoActiveStructure       = errors.New("no active salary structure for the employee")
)

type SalaryService struct {
	structures *repositories.SalaryStructureRepository
	employees  *repositories.EmployeeRepository
	audit      *auditmanager.Service
}

func NewSalaryService(structures *repositories.SalaryStructureRepository, employees *repositories.EmployeeRepository, audit *auditmanager.Service) *SalaryService {
	return &SalaryService{structures: structures, employees: employees, audit: audit}
}

func (s *SalaryService) CreateStructure(in struct {
	EmployeeID, BasicSalary, Allowances, Bonus, OvertimeRate, TaxRate, TaxAmount, Deductions, EffectiveFrom, Status string
}, actorID, ip, ua string) (*models.SalaryStructure, error) {
	if _, err := s.employees.FindByID(in.EmployeeID); err != nil {
		return nil, ErrNotFound
	}
	basic, err := models.FromDecimal(in.BasicSalary)
	if err != nil {
		return nil, err
	}
	eff, err := time.Parse("2006-01-02", in.EffectiveFrom)
	if err != nil {
		return nil, errors.New("invalid effective_from (expected YYYY-MM-DD)")
	}
	st := &models.SalaryStructure{
		EmployeeID:    in.EmployeeID,
		BasicSalary:   basic,
		Allowances:    decOrZero(in.Allowances),
		Bonus:         decOrZero(in.Bonus),
		OvertimeRate:  decOrZero(in.OvertimeRate),
		TaxRate:       decOrZero(in.TaxRate),
		TaxAmount:     decOrZero(in.TaxAmount),
		Deductions:    decOrZero(in.Deductions),
		EffectiveFrom: datatypes.Date(eff),
		Status:        models.SalaryStructureStatus(orStr(in.Status, string(models.SalaryActive))),
	}
	if err := s.structures.Create(st); err != nil {
		return nil, err
	}
	s.audit.Record(actorID, models.ActionCreate, "salary_structure", st.ID, ip, ua, map[string]string{"employee_id": in.EmployeeID})
	return st, nil
}

func (s *SalaryService) UpdateStructure(id string, in struct {
	BasicSalary, Allowances, Bonus, OvertimeRate, TaxRate, TaxAmount, Deductions, EffectiveFrom, EffectiveUntil, Status string
}, actorID, ip, ua string) (*models.SalaryStructure, error) {
	if _, err := s.structures.FindByID(id); err != nil {
		return nil, ErrNotFound
	}
	fields := map[string]interface{}{}
	setDec := func(key, val string) {
		if val == "" {
			return
		}
		if d, err := models.FromDecimal(val); err == nil {
			fields[key] = d
		}
	}
	setDec("basic_salary", in.BasicSalary)
	setDec("allowances", in.Allowances)
	setDec("bonus", in.Bonus)
	setDec("overtime_rate", in.OvertimeRate)
	setDec("tax_rate", in.TaxRate)
	setDec("tax_amount", in.TaxAmount)
	setDec("deductions", in.Deductions)
	if in.EffectiveFrom != "" {
		if eff, err := time.Parse("2006-01-02", in.EffectiveFrom); err == nil {
			fields["effective_from"] = datatypes.Date(eff)
		}
	}
	if in.EffectiveUntil != "" {
		if eff, err := time.Parse("2006-01-02", in.EffectiveUntil); err == nil {
			fields["effective_until"] = datatypes.Date(eff)
		}
	}
	if in.Status != "" {
		fields["status"] = in.Status
	}
	if len(fields) > 0 {
		if err := s.structures.Update(id, fields); err != nil {
			return nil, err
		}
	}
	s.audit.Record(actorID, models.ActionUpdate, "salary_structure", id, ip, ua, nil)
	return s.structures.FindByID(id)
}

func (s *SalaryService) DeleteStructure(id, actorID, ip, ua string) error {
	if _, err := s.structures.FindByID(id); err != nil {
		return ErrNotFound
	}
	if err := s.structures.Delete(id); err != nil {
		return err
	}
	s.audit.Record(actorID, models.ActionDelete, "salary_structure", id, ip, ua, nil)
	return nil
}

func (s *SalaryService) GetStructure(id string) (*models.SalaryStructure, error) {
	st, err := s.structures.FindByID(id)
	if err != nil {
		return nil, ErrNotFound
	}
	return st, nil
}

func (s *SalaryService) ListStructures(employeeID string) ([]models.SalaryStructure, error) {
	return s.structures.List(employeeID)
}

type PayrollService struct {
	payroll *repositories.PayrollRepository
	salary  *repositories.SalaryStructureRepository
	audit   *auditmanager.Service
}

func NewPayrollService(payroll *repositories.PayrollRepository, salary *repositories.SalaryStructureRepository, audit *auditmanager.Service) *PayrollService {
	return &PayrollService{payroll: payroll, salary: salary, audit: audit}
}

// Generate drafts payroll for every active employee with a salary structure.
func (s *PayrollService) Generate(month, year int, actorID, ip, ua string) (int, error) {
	on := lastDayOfMonth(month, year)
	employees, err := s.payroll.ActiveEmployeesWithSalary(datatypes.Date(on))
	if err != nil {
		return 0, err
	}
	created := 0
	err = s.payroll.Tx(func(tx *gorm.DB) error {
		for _, e := range employees {
			exists, err := s.payroll.ExistsForEmployee(e.ID, month, year)
			if err != nil || exists {
				continue
			}
			st, err := s.salary.EffectiveFor(e.ID, datatypes.Date(on))
			if err != nil {
				continue
			}
			p := models.Payroll{
				Month:             month,
				Year:              year,
				EmployeeID:        e.ID,
				SalaryStructureID: st.ID,
				BasicSalary:       st.BasicSalary,
				Allowances:        st.Allowances,
				Bonus:             st.Bonus,
				Overtime:          models.Decimal("0.00"),
				GrossSalary:       models.Decimal("0.00"),
				Tax:               models.Decimal("0.00"),
				Deductions:        st.Deductions,
				NetSalary:         models.Decimal("0.00"),
				Status:            models.PayrollDraft,
			}
			if err := tx.Create(&p).Error; err != nil {
				return err
			}
			created++
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	s.audit.Record(actorID, models.ActionCreate, "payroll", "", ip, ua, map[string]string{"month": itoa(month), "year": itoa(year), "created": itoa(created)})
	return created, nil
}

// Process computes final figures and marks a draft payroll as processed.
func (s *PayrollService) Process(id string, in struct{ Bonus, Overtime, Deductions, Notes string }, actorID, ip, ua string) (*models.Payroll, error) {
	p, err := s.payroll.FindByID(id)
	if err != nil {
		return nil, ErrNotFound
	}
	if in.Bonus != "" {
		if d, err := models.FromDecimal(in.Bonus); err == nil {
			p.Bonus = d
		}
	}
	if in.Deductions != "" {
		if d, err := models.FromDecimal(in.Deductions); err == nil {
			p.Deductions = d
		}
	}

	overtimeHours := 0.0
	if in.Overtime != "" {
		if d, err := models.FromDecimal(in.Overtime); err == nil {
			p.Overtime = d
		}
	} else {
		ot, err := s.payroll.EmployeeAttendanceOvertime(p.EmployeeID, p.Month, p.Year)
		if err == nil {
			overtimeHours = ot
		}
	}

	st, err := s.salary.EffectiveFor(p.EmployeeID, datatypes.Date(lastDayOfMonth(p.Month, p.Year)))
	if err == nil {
		if p.Overtime.IsZero() && overtimeHours > 0 {
			rate := st.OvertimeRate
			if rate.IsZero() {
				rate = models.MustDecimal("1.00")
			}
			p.Overtime = models.Mul(rate, models.MustDecimal(fmt.Sprintf("%.2f", overtimeHours)))
		}
		if p.Tax.IsZero() {
			taxAmount := st.TaxAmount
			if !taxAmount.IsZero() {
				p.Tax = taxAmount
			} else {
				p.Tax = models.PercentOf(payrollGross(p), st.TaxRate)
			}
		}
		if p.Deductions.IsZero() {
			p.Deductions = st.Deductions
		}
	}

	p.GrossSalary = payrollGross(p)
	p.NetSalary = models.Sub(p.GrossSalary, models.Add(p.Tax, p.Deductions))

	if p.Status != models.PayrollDraft && p.Status != models.PayrollProcessing {
		return nil, ErrInvalidStatusTransition
	}
	if p.Status == models.PayrollDraft {
		p.Status = models.PayrollProcessing
	}

	now := time.Now().UTC()
	fields := map[string]interface{}{
		"bonus": p.Bonus, "overtime": p.Overtime, "deductions": p.Deductions,
		"gross_salary": p.GrossSalary, "tax": p.Tax, "net_salary": p.NetSalary,
		"status": p.Status, "processed_by": actorID, "processed_at": &now, "notes": in.Notes,
	}
	if err := s.payroll.Update(id, fields); err != nil {
		return nil, err
	}
	s.audit.Record(actorID, models.ActionPayrollProcess, "payroll", id, ip, ua, map[string]string{"month": itoa(p.Month), "year": itoa(p.Year), "net": p.NetSalary.FloatString()})
	return s.payroll.FindByID(id)
}

func (s *PayrollService) MarkPaid(id, paymentRef, notes string, actorID, ip, ua string) (*models.Payroll, error) {
	p, err := s.payroll.FindByID(id)
	if err != nil {
		return nil, ErrNotFound
	}
	if p.Status != models.PayrollProcessed && p.Status != models.PayrollProcessing {
		return nil, ErrInvalidStatusTransition
	}
	now := time.Now().UTC()
	fields := map[string]interface{}{"status": models.PayrollPaid, "paid_on": &now, "payment_ref": paymentRef, "notes": notes}
	if err := s.payroll.Update(id, fields); err != nil {
		return nil, err
	}
	s.audit.Record(actorID, models.ActionUpdate, "payroll", id, ip, ua, map[string]string{"action": "mark_paid"})
	return s.payroll.FindByID(id)
}

func (s *PayrollService) Cancel(id, notes string, actorID, ip, ua string) (*models.Payroll, error) {
	p, err := s.payroll.FindByID(id)
	if err != nil {
		return nil, ErrNotFound
	}
	if p.Status == models.PayrollPaid {
		return nil, ErrInvalidStatusTransition
	}
	fields := map[string]interface{}{"status": models.PayrollCancelled, "notes": notes}
	if err := s.payroll.Update(id, fields); err != nil {
		return nil, err
	}
	s.audit.Record(actorID, models.ActionUpdate, "payroll", id, ip, ua, map[string]string{"action": "cancel"})
	return s.payroll.FindByID(id)
}

func (s *PayrollService) Get(id string) (*models.Payroll, error) {
	p, err := s.payroll.FindByID(id)
	if err != nil {
		return nil, ErrNotFound
	}
	return p, nil
}

func (s *PayrollService) Payslip(id string) (*models.Payroll, error) {
	return s.Get(id)
}

func (s *PayrollService) List(p utils.Pagination, month, year int, employeeID, status, departmentID string) ([]models.Payroll, int64, error) {
	return s.payroll.List(p, month, year, employeeID, status, departmentID)
}

func payrollGross(p *models.Payroll) models.Decimal {
	return models.Add(models.Add(p.BasicSalary, p.Allowances), models.Add(p.Bonus, p.Overtime))
}

func decOrZero(s string) models.Decimal {
	if s == "" {
		return models.Decimal("0.00")
	}
	d, err := models.FromDecimal(s)
	if err != nil {
		return models.Decimal("0.00")
	}
	return d
}

func lastDayOfMonth(month, year int) time.Time {
	first := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	return first.AddDate(0, 1, -1)
}

func itoa(n int) string {
	return fmt.Sprintf("%d", n)
}
