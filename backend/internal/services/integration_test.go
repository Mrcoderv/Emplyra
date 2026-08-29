package services

import (
	"errors"
	"os"
	"testing"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/emplyra/backend/internal/auditmanager"
	"github.com/emplyra/backend/internal/database"
	"github.com/emplyra/backend/internal/models"
	"github.com/emplyra/backend/internal/notifications"
	"github.com/emplyra/backend/internal/repositories"
)

func integrationDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DB_URL")
	if dsn == "" {
		t.Skip("TEST_DB_URL not set; skipping integration test")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("connect test db: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}
	return db
}

func mustEmployee(t *testing.T, db *gorm.DB, code string) *models.Employee {
	t.Helper()
	e := &models.Employee{
		EmployeeCode: code,
		FirstName:    "Test",
		LastName:     "User",
		Email:        code + "@example.com",
		Status:       models.EmployeeActive,
	}
	if err := repositories.NewEmployeeRepository(db).Create(e); err != nil {
		t.Fatalf("create employee: %v", err)
	}
	return e
}

func TestLeaveApproveFlowIntegration(t *testing.T) {
	db := integrationDB(t)
	actor := uuid.NewString()

	empRepo := repositories.NewEmployeeRepository(db)
	ltRepo := repositories.NewLeaveTypeRepository(db)
	balRepo := repositories.NewLeaveBalanceRepository(db)
	leaveRepo := repositories.NewLeaveRepository(db)
	audit := auditmanager.New(db)
	notify := notifications.New(db)
	svc := NewLeaveService(leaveRepo, ltRepo, balRepo, empRepo, notify, audit)

	emp := mustEmployee(t, db, "LEAVEIT01")
	var lt models.LeaveType
	if err := db.Create(&models.LeaveType{Name: "Annual Leave", Code: "ANNUAL_IT", IsPaid: true}).Scan(&lt).Error; err != nil {
		t.Fatalf("create leave type: %v", err)
	}

	if _, err := svc.SetBalance(models.DefaultTenantID, emp.ID, lt.ID, 2026, 10, actor, "", ""); err != nil {
		t.Fatalf("set balance: %v", err)
	}
	bal, err := balRepo.Find(models.DefaultTenantID, emp.ID, lt.ID, 2026)
	if err != nil || bal.Entitlement != 10 {
		t.Fatalf("expected entitlement 10, got %+v err=%v", bal, err)
	}

	// Mon-Fri range spanning a weekend should compute business days only.
	l, err := svc.Create(models.DefaultTenantID, emp.ID, lt.ID, "2026-09-07", "2026-09-13", "vacation", actor, "", "")
	if err != nil {
		t.Fatalf("create leave: %v", err)
	}
	if l.Days != 5 {
		t.Fatalf("expected 5 business days, got %d", l.Days)
	}
	if l.Status != models.LeavePending {
		t.Fatalf("expected PENDING, got %s", l.Status)
	}

	approved, err := svc.Approve(models.DefaultTenantID, l.ID, "approved", actor, "", "")
	if err != nil {
		t.Fatalf("approve: %v", err)
	}
	if approved.Status != models.LeaveApproved {
		t.Fatalf("expected APPROVED, got %s", approved.Status)
	}
	bal, err = balRepo.Find(models.DefaultTenantID, emp.ID, lt.ID, 2026)
	if err != nil || bal.Used != 5 {
		t.Fatalf("expected used=5, got %+v err=%v", bal, err)
	}

	// Overlapping leave must be rejected during create.
	if _, err := svc.Create(models.DefaultTenantID, emp.ID, lt.ID, "2026-09-08", "2026-09-09", "overlap", actor, "", ""); !errors.Is(err, ErrLeaveOverlap) {
		t.Fatalf("expected ErrLeaveOverlap, got %v", err)
	}

	// Exceeding balance must be rejected.
	if _, err := svc.Create(models.DefaultTenantID, emp.ID, lt.ID, "2026-10-01", "2026-10-15", "too much", actor, "", ""); !errors.Is(err, ErrInsufficientLeaveBalance) {
		t.Fatalf("expected ErrInsufficientLeaveBalance, got %v", err)
	}

	_ = db.Delete(&models.Leave{}, "employee_id = ?", emp.ID)
	_ = db.Delete(&models.LeaveBalance{}, "employee_id = ?", emp.ID)
	_ = db.Delete(&models.LeaveType{}, "id = ?", lt.ID)
	_ = db.Delete(&models.Employee{}, "id = ?", emp.ID)
}

func TestAttendanceCheckInOutIntegration(t *testing.T) {
	db := integrationDB(t)
	actor := uuid.NewString()

	empRepo := repositories.NewEmployeeRepository(db)
	attRepo := repositories.NewAttendanceRepository(db)
	audit := auditmanager.New(db)
	svc := NewAttendanceService(attRepo, empRepo, audit)

	emp := mustEmployee(t, db, "ATTIT01")

	a, err := svc.CheckIn(models.DefaultTenantID, emp.ID, "in", actor, "", "")
	if err != nil {
		t.Fatalf("check-in: %v", err)
	}
	if a.ID == "" {
		t.Fatal("expected attendance id")
	}
	if a.Status == "" {
		t.Fatal("expected attendance status")
	}

	if _, err := svc.CheckIn(models.DefaultTenantID, emp.ID, "again", actor, "", ""); !errors.Is(err, ErrAlreadyCheckedIn) {
		t.Fatalf("expected ErrAlreadyCheckedIn on duplicate, got %v", err)
	}

	out, err := svc.CheckOut(models.DefaultTenantID, emp.ID, "out", 0, actor, "", "")
	if err != nil {
		t.Fatalf("check-out: %v", err)
	}
	if out.CheckOut == nil || out.WorkingHours < 0 {
		t.Fatalf("unexpected check-out record: %+v", out)
	}

	// Unauthenticated role fallback: employee id lookup by user id must fail cleanly.
	if _, err := svc.EmployeeIDForUser(uuid.NewString()); err == nil {
		t.Fatal("expected error for unknown user")
	}

	_ = db.Delete(&models.Attendance{}, "employee_id = ?", emp.ID)
	_ = db.Delete(&models.Employee{}, "id = ?", emp.ID)
}
