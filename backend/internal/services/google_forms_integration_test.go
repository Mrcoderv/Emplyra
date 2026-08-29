package services

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/emplyra/backend/internal/auditmanager"
	"github.com/emplyra/backend/internal/google"
	"github.com/emplyra/backend/internal/models"
	"github.com/emplyra/backend/internal/notifications"
	"github.com/emplyra/backend/internal/repositories"
	"github.com/emplyra/backend/internal/utils"
)

type fakeSheets struct {
	values [][]string
	names  []string
	err    error
}

func (f *fakeSheets) ListSheets(string) ([]string, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.names, nil
}

func (f *fakeSheets) GetValues(string, string) ([][]string, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.values, nil
}

func newGoogleFormHarness(t *testing.T, sheets google.SheetsReader) (*gorm.DB, *RecruitmentService, *GoogleFormService) {
	t.Helper()
	db := integrationDB(t)

	jobRepo := repositories.NewJobPostRepository(db)
	candRepo := repositories.NewCandidateRepository(db)
	appRepo := repositories.NewApplicationRepository(db)
	interviewRepo := repositories.NewInterviewRepository(db)
	onboardRepo := repositories.NewOnboardingRepository(db)
	empRepo := repositories.NewEmployeeRepository(db)
	userRepo := repositories.NewUserRepository(db)
	roleRepo := repositories.NewRoleRepository(db)
	audit := auditmanager.New(db)
	notify := notifications.New(db)

	recruit := NewRecruitmentService(jobRepo, candRepo, appRepo, interviewRepo, onboardRepo,
		NewEmployeeService(empRepo, repositories.NewDepartmentRepository(db), repositories.NewDesignationRepository(db), audit),
		userRepo, roleRepo, notify, audit)

	g := NewGoogleFormService(
		repositories.NewGoogleFormIntegrationRepository(db),
		repositories.NewGoogleFormResponseRepository(db),
		repositories.NewGoogleOAuthTokenRepository(db),
		sheets, nil, "/", recruit, notify, audit)

	return db, recruit, g
}

func TestGoogleFormSyncFlowIntegration(t *testing.T) {
	db, recruit, g := newGoogleFormHarness(t, &fakeSheets{
		names: []string{"Form Responses 1"},
		values: [][]string{
			{"Timestamp", "Full Name", "Email", "Phone", "Skills", "Cover Letter"},
			{"8/29/2026 10:00:00", "Jane Doe", "jane@example.com", "555-1234", "Go, SQL", "hire me"},
			{"8/29/2026 11:00:00", "John Smith", "john@example.com", "555-9999", "Python", ""},
			{"", "No Mail", "", "555-0000", "", ""},
		},
	})
	actor := uuid.NewString()

	job, err := recruit.CreateJob(models.DefaultTenantID, JobPostInput{Title: "Backend Engineer", Vacancies: 2}, actor, "", "")
	if err != nil {
		t.Fatalf("create job: %v", err)
	}

	if _, err := g.Connect(models.DefaultTenantID, job.ID, GoogleFormInput{
		FormURL: "https://docs.google.com/forms/d/e/FORMID/viewform",
		SheetID: "sheets-id-1",
	}, actor, "", ""); err != nil {
		t.Fatalf("connect: %v", err)
	}

	result, err := g.Sync(models.DefaultTenantID, job.ID, "incremental", actor, "", "")
	if err != nil {
		t.Fatalf("sync: %v", err)
	}
	if result.Imported != 2 {
		t.Fatalf("expected 2 imported, got %+v", result)
	}
	if result.Failed != 1 {
		t.Fatalf("expected 1 failed (missing email), got %+v", result)
	}

	_, total, err := recruit.Applications(models.DefaultTenantID, utils.NewPagination("1", "100"), job.ID, "", "")
	if err != nil || total != 2 {
		t.Fatalf("expected 2 applications, got %d err=%v", total, err)
	}

	_, counts, err := g.SyncStatus(models.DefaultTenantID, job.ID)
	if err != nil {
		t.Fatalf("sync status: %v", err)
	}
	if counts.Imported != 2 || counts.Failed != 1 {
		t.Fatalf("unexpected counters: %+v", counts)
	}

	result, err = g.Sync(models.DefaultTenantID, job.ID, "full", actor, "", "")
	if err != nil {
		t.Fatalf("resync: %v", err)
	}
	// All three responses (2 imported + 1 recorded error) already exist in the
	// ledger, so a full re-scan imports nothing and skips everything.
	if result.Imported != 0 || result.Duplicates != 3 {
		t.Fatalf("expected 0 imported and 3 duplicates, got %+v", result)
	}

	_, total, err = recruit.Applications(models.DefaultTenantID, utils.NewPagination("1", "100"), job.ID, "", "")
	if err != nil || total != 2 {
		t.Fatalf("applications must stay at 2, got %d err=%v", total, err)
	}

	// Incremental sync after a completed run must not touch the sheet at all.
	result, err = g.Sync(models.DefaultTenantID, job.ID, "incremental", actor, "", "")
	if err != nil {
		t.Fatalf("incremental resync: %v", err)
	}
	if result.Imported != 0 || result.Duplicates != 0 {
		t.Fatalf("expected no-op incremental sync, got %+v", result)
	}

	cleanupGoogleFormData(t, db, job.ID)
}

func TestGoogleFormIncrementalAppendIntegration(t *testing.T) {
	sheets := &fakeSheets{
		names: []string{"Form Responses 1"},
		values: [][]string{
			{"Timestamp", "Full Name", "Email", "Skills"},
			{"8/29/2026 09:00:00", "Amy Pond", "amy@example.com", "Go"},
		},
	}
	db, recruit, g := newGoogleFormHarness(t, sheets)
	actor := uuid.NewString()

	job, err := recruit.CreateJob(models.DefaultTenantID, JobPostInput{Title: "SRE", Vacancies: 2}, actor, "", "")
	if err != nil {
		t.Fatalf("create job: %v", err)
	}
	if _, err := g.Connect(models.DefaultTenantID, job.ID, GoogleFormInput{
		FormURL: "https://docs.google.com/forms/d/e/APPEND/viewform",
		SheetID: "sheets-append",
	}, actor, "", ""); err != nil {
		t.Fatalf("connect: %v", err)
	}

	if res, err := g.Sync(models.DefaultTenantID, job.ID, "incremental", actor, "", ""); err != nil || res.Imported != 1 {
		t.Fatalf("initial sync: %+v err=%v", res, err)
	}

	sheets.values = append(sheets.values, []string{"8/29/2026 12:00:00", "River Song", "river@example.com", "K8s"})

	res, err := g.Sync(models.DefaultTenantID, job.ID, "incremental", actor, "", "")
	if err != nil {
		t.Fatalf("append sync: %v", err)
	}
	if res.Imported != 1 {
		t.Fatalf("expected exactly the new row imported, got %+v", res)
	}

	_, total, err := recruit.Applications(models.DefaultTenantID, utils.NewPagination("1", "100"), job.ID, "", "")
	if err != nil || total != 2 {
		t.Fatalf("expected 2 applications after append, got %d err=%v", total, err)
	}

	cleanupGoogleFormData(t, db, job.ID)
}

func TestGoogleFormSyncSheetErrorSurfacesIntegration(t *testing.T) {
	sheets := &fakeSheets{err: google.ErrInvalidSpreadsheet}
	db, recruit, g := newGoogleFormHarness(t, sheets)
	actor := uuid.NewString()

	job, err := recruit.CreateJob(models.DefaultTenantID, JobPostInput{Title: "Ops", Vacancies: 1}, actor, "", "")
	if err != nil {
		t.Fatalf("create job: %v", err)
	}
	if _, err := g.Connect(models.DefaultTenantID, job.ID, GoogleFormInput{
		FormURL: "https://docs.google.com/forms/d/e/OTHER/viewform",
		SheetID: "bad-sheet-id",
	}, actor, "", ""); err != nil {
		t.Fatalf("connect: %v", err)
	}

	_, err = g.Sync(models.DefaultTenantID, job.ID, "incremental", actor, "", "")
	if !errors.Is(err, ErrGoogleInvalidSpreadsheet) {
		t.Fatalf("expected ErrGoogleInvalidSpreadsheet, got %v", err)
	}

	integ, err := g.Get(models.DefaultTenantID, job.ID)
	if err != nil {
		t.Fatalf("get integration: %v", err)
	}
	if integ.SyncError == "" {
		t.Fatal("expected integration.sync_error to be recorded")
	}
	if integ.Status != models.GoogleFormStatusError {
		t.Fatalf("expected ERROR status, got %q", integ.Status)
	}

	cleanupGoogleFormData(t, db, job.ID)
}

func cleanupGoogleFormData(t *testing.T, db *gorm.DB, jobID string) {
	t.Helper()
	_ = db.Exec("DELETE FROM google_form_responses WHERE integration_id IN (SELECT id FROM google_form_integrations WHERE job_id = ?)", jobID).Error
	_ = db.Delete(&models.GoogleFormIntegration{}, "job_id = ?", jobID).Error
	_ = db.Exec("DELETE FROM candidates WHERE source = 'Google Forms'").Error
	_ = db.Delete(&models.JobPost{}, "id = ?", jobID).Error
}
