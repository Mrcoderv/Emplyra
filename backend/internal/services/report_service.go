package services

import (
	"time"

	"gorm.io/gorm"

	"github.com/emplyra/backend/internal/models"
)

// ReportService runs read-only analytical queries across tables. It is the
// single entry point for reports and the dashboard summary.
type ReportService struct {
	db *gorm.DB
}

func NewReportService(db *gorm.DB) *ReportService { return &ReportService{db: db} }

type HeadcountRow struct {
	DepartmentID string `json:"department_id"`
	Department   string `json:"department"`
	Total        int64  `json:"total"`
	Active       int64  `json:"active"`
	Inactive     int64  `json:"inactive"`
	OnLeave      int64  `json:"on_leave"`
	Terminated   int64  `json:"terminated"`
}

// HeadcountByDepartment returns employee counts grouped by department.
func (s *ReportService) HeadcountByDepartment() ([]HeadcountRow, error) {
	rows := make([]HeadcountRow, 0)
	err := s.db.Model(&models.Employee{}).
		Select(`COALESCE(departments.id,'') AS department_id,
		        COALESCE(departments.name,'Unassigned') AS department,
		        COUNT(*) AS total,
		        COUNT(*) FILTER (WHERE employees.status = 'ACTIVE') AS active,
		        COUNT(*) FILTER (WHERE employees.status = 'INACTIVE') AS inactive,
		        COUNT(*) FILTER (WHERE employees.status = 'ON_LEAVE') AS on_leave,
		        COUNT(*) FILTER (WHERE employees.status = 'TERMINATED') AS terminated`).
		Joins("LEFT JOIN departments ON departments.id = employees.department_id").
		Group("departments.id, departments.name").
		Order("department ASC").
		Scan(&rows).Error
	return rows, err
}

type AttendanceSummary struct {
	Status    string   `json:"status"`
	Count     int64    `json:"count"`
	LateCount int64    `json:"late_count"`
	AvgHours  *float64 `json:"avg_hours,omitempty"`
}

func (s *ReportService) AttendanceSummary(from, to string) ([]AttendanceSummary, float64, error) {
	fromT, err := parseDate(from)
	if err != nil {
		if from != "" {
			return nil, 0, err
		}
	}
	toT, err := parseDate(to)
	if err != nil {
		if to != "" {
			return nil, 0, err
		}
	}
	q := s.db.Model(&models.Attendance{})
	if fromT != nil {
		q = q.Where("date >= ?", fromT.Format("2006-01-02"))
	}
	if toT != nil {
		q = q.Where("date <= ?", toT.Format("2006-01-02"))
	}
	var rows []struct {
		Status    string
		Count     int64
		LateCount int64
	}
	if err := q.Select(`status,
               COUNT(*) AS count,
               COUNT(*) FILTER (WHERE late_minutes > 0) AS late_count`).
		Group("status").Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	out := make([]AttendanceSummary, 0, len(rows))
	var totalPresent int64
	for _, r := range rows {
		out = append(out, AttendanceSummary{Status: r.Status, Count: r.Count, LateCount: r.LateCount})
		if r.Status == string(models.AttendancePresent) || r.Status == string(models.AttendanceLate) {
			totalPresent += r.Count
		}
	}
	var avgHours *float64
	var hr struct{ H float64 }
	if err := s.db.Model(&models.Attendance{}).
		Select("COALESCE(AVG(working_hours),0) AS h").Scan(&hr).Error; err == nil {
		if f := hr.H; f > 0 {
			avgHours = &f
		}
	}
	if totalPresent > 0 {
		avgHours = &hr.H
	}
	return out, *sumIf(avgHours), nil
}

func sumIf(f *float64) *float64 {
	if f == nil {
		return floatPtr(0)
	}
	return f
}

func floatPtr(f float64) *float64 { return &f }

type LeaveSummary struct {
	Pending    int64         `json:"pending"`
	Approved   int64         `json:"approved"`
	Rejected   int64         `json:"rejected"`
	DaysByType []TypeDaysRow `json:"days_by_type"`
}

type TypeDaysRow struct {
	LeaveTypeID string `json:"leave_type_id"`
	LeaveType   string `json:"leave_type"`
	Days        int64  `json:"days"`
}

func (s *ReportService) LeaveSummary() (*LeaveSummary, error) {
	out := &LeaveSummary{}
	if err := s.db.Model(&models.Leave{}).
		Select(`COUNT(*) FILTER (WHERE status = 'PENDING') AS pending,
		        COUNT(*) FILTER (WHERE status = 'APPROVED') AS approved,
		        COUNT(*) FILTER (WHERE status = 'REJECTED') AS rejected`).
		Scan(out).Error; err != nil {
		return nil, err
	}
	var typeRows []struct {
		LeaveTypeID string
		LeaveType   string
		Days        int64
	}
	if err := s.db.Model(&models.Leave{}).
		Select(`leaves.leave_type_id,
		        COALESCE(leave_types.name,'') AS leave_type,
		        COALESCE(SUM(leaves.days),0) AS days`).
		Joins("LEFT JOIN leave_types ON leave_types.id = leaves.leave_type_id").
		Where("leaves.status = ?", string(models.LeaveApproved)).
		Group("leaves.leave_type_id, leave_types.name").
		Scan(&typeRows).Error; err != nil {
		return nil, err
	}
	for _, t := range typeRows {
		out.DaysByType = append(out.DaysByType, TypeDaysRow{LeaveTypeID: t.LeaveTypeID, LeaveType: t.LeaveType, Days: t.Days})
	}
	return out, nil
}

type PayrollSummary struct {
	Gross      models.Decimal `json:"gross_salary"`
	Net        models.Decimal `json:"net_salary"`
	Tax        models.Decimal `json:"tax"`
	Deductions models.Decimal `json:"deductions"`
	Entries    int64          `json:"entries"`
	Processed  int64          `json:"processed"`
	Paid       int64          `json:"paid"`
	Draft      int64          `json:"draft"`
}

func (s *ReportService) PayrollSummary(month, year int) (*PayrollSummary, error) {
	out := &PayrollSummary{}
	q := s.db.Model(&models.Payroll{}).
		Where("status <> ?", string(models.PayrollCancelled))
	if month > 0 {
		q = q.Where("month = ?", month)
	}
	if year > 0 {
		q = q.Where("year = ?", year)
	}
	if err := q.Select(`COALESCE(SUM(gross_salary),0)::text AS gross_salary,
	        COALESCE(SUM(net_salary),0)::text AS net_salary,
	        COALESCE(SUM(tax),0)::text AS tax,
	        COALESCE(SUM(deductions),0)::text AS deductions,
	        COUNT(*) AS entries,
	        COUNT(*) FILTER (WHERE status = 'PROCESSED') AS processed,
	        COUNT(*) FILTER (WHERE status = 'PAID') AS paid,
	        COUNT(*) FILTER (WHERE status = 'DRAFT') AS draft`).
		Scan(out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

type FunnelRow struct {
	Status string `json:"status"`
	Count  int64  `json:"count"`
}

func (s *ReportService) RecruitmentFunnel() ([]FunnelRow, error) {
	var rows []struct {
		Status string
		Count  int64
	}
	if err := s.db.Model(&models.Application{}).
		Select("status, COUNT(*) AS count").
		Group("status").Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]FunnelRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, FunnelRow{Status: r.Status, Count: r.Count})
	}
	return out, nil
}

func (s *ReportService) UpcomingHolidays(limit int) ([]models.Holiday, error) {
	if limit <= 0 || limit > 50 {
		limit = 5
	}
	var items []models.Holiday
	err := s.db.Where("status = ? AND date >= ?", string(models.HolidayActive), time.Now().Format("2006-01-02")).
		Order("date ASC").Limit(limit).Find(&items).Error
	return items, err
}

// --- Dashboard ---

type DashboardSummary struct {
	Employees struct {
		Total      int64 `json:"total"`
		Active     int64 `json:"active"`
		OnLeave    int64 `json:"on_leave"`
		Onboarding int64 `json:"onboarding"`
	} `json:"employees"`
	Departments        int64          `json:"departments"`
	OpenJobs           int64          `json:"open_jobs"`
	Candidates         int64          `json:"candidates"`
	PendingLeaves      int64          `json:"pending_leaves"`
	PresentToday       int64          `json:"present_today"`
	MonthPayroll       models.Decimal `json:"month_payroll_net"`
	UpcomingHolidays   int            `json:"upcoming_holidays"`
	PendingReviews     int64          `json:"pending_reviews"`
	TrainingsScheduled int64          `json:"trainings_scheduled"`
}

func (s *ReportService) Summary(employeeID string) (*DashboardSummary, error) {
	out := &DashboardSummary{}
	d := s.db.Model(&models.Employee{})
	if err := d.Select(`COUNT(*) AS total,
	    COUNT(*) FILTER (WHERE status = 'ACTIVE') AS active,
	    COUNT(*) FILTER (WHERE status = 'ON_LEAVE') AS on_leave,
	    COUNT(*) FILTER (WHERE status = 'ONBOARDING') AS onboarding`).
		Scan(&out.Employees).Error; err != nil {
		return nil, err
	}
	_ = s.db.Model(&models.Department{}).Count(&out.Departments).Error
	_ = s.db.Model(&models.JobPost{}).Where("status = ?", string(models.JobOpen)).Count(&out.OpenJobs).Error
	_ = s.db.Model(&models.Candidate{}).Count(&out.Candidates).Error
	_ = s.db.Model(&models.Leave{}).Where("status = ?", string(models.LeavePending)).Count(&out.PendingLeaves).Error
	_ = s.db.Model(&models.Attendance{}).Where("date = ?", time.Now().Format("2006-01-02")).
		Where("status IN ?", []string{string(models.AttendancePresent), string(models.AttendanceLate)}).
		Count(&out.PresentToday).Error

	now := time.Now()
	var monthNet struct{ M models.Decimal }
	if err := s.db.Model(&models.Payroll{}).
		Select("COALESCE(SUM(net_salary),0)::text AS m").
		Where("month = ? AND year = ? AND status IN ?", int(now.Month()), now.Year(),
			[]string{string(models.PayrollProcessed), string(models.PayrollPaid)}).
		Scan(&monthNet).Error; err == nil {
		out.MonthPayroll = monthNet.M
	}

	var upcoming int64
	_ = s.db.Model(&models.Holiday{}).
		Where("status = ? AND date >= ?", string(models.HolidayActive), now.Format("2006-01-02")).
		Count(&upcoming).Error
	out.UpcomingHolidays = int(upcoming)

	_ = s.db.Model(&models.PerformanceReview{}).
		Where("status IN ?", []string{string(models.ReviewSelfSubmitted), string(models.ReviewManagerDone)}).
		Count(&out.PendingReviews).Error

	_ = s.db.Model(&models.TrainingProgram{}).
		Where("status = ?", string(models.TrainingScheduled)).Count(&out.TrainingsScheduled).Error

	if employeeID != "" {
		var own struct {
			PendingLeaves int64
			PresentDays   int64
		}
		_ = s.db.Model(&models.Leave{}).Where("employee_id = ? AND status = ?", employeeID, string(models.LeavePending)).Count(&own.PendingLeaves).Error
		out.PendingLeaves = own.PendingLeaves
	}
	return out, nil
}
