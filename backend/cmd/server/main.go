package main

import (
	"log/slog"
	"os"

	"github.com/emplyra/backend/internal/auditmanager"
	"github.com/emplyra/backend/internal/auth"
	"github.com/emplyra/backend/internal/config"
	"github.com/emplyra/backend/internal/database"
	"github.com/emplyra/backend/internal/google"
	"github.com/emplyra/backend/internal/notifications"
	"github.com/emplyra/backend/internal/repositories"
	"github.com/emplyra/backend/internal/routes"
	"github.com/emplyra/backend/internal/seed"
	"github.com/emplyra/backend/internal/services"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config", "err", err)
		os.Exit(1)
	}

	db, err := database.Connect(cfg)
	if err != nil {
		slog.Error("database", "err", err)
		os.Exit(1)
	}

	if err := database.Migrate(db); err != nil {
		slog.Error("migrate", "err", err)
		os.Exit(1)
	}

	if err := seed.Run(db, cfg); err != nil {
		slog.Error("seed", "err", err)
		os.Exit(1)
	}

	if cfg.UploadDir != "" {
		if err := os.MkdirAll(cfg.UploadDir, 0o755); err != nil {
			slog.Warn("upload dir", "err", err)
		}
	}

	// --- Infrastructure ---
	audit := auditmanager.New(db)
	notify := notifications.New(db)
	jwt := auth.NewJWTManager(cfg.JWTSecret, cfg.AccessTokenTTL)

	// --- Google integration ---
	googleCfg := google.ConfigFromEnv()
	if googleCfg.TokenKey == "" {
		googleCfg.TokenKey = cfg.JWTSecret
	}
	oauthTokenRepo := repositories.NewGoogleOAuthTokenRepository(db)
	tokenManager := google.NewTokenManager(googleCfg, oauthTokenRepo)
	sheetsClient := google.NewSheetsClient(tokenManager)
	googleFormRepo := repositories.NewGoogleFormIntegrationRepository(db)
	googleFormRespRepo := repositories.NewGoogleFormResponseRepository(db)

	// --- Repositories ---
	userRepo := repositories.NewUserRepository(db)
	roleRepo := repositories.NewRoleRepository(db)
	permRepo := repositories.NewPermissionRepository(db)
	tokenRepo := repositories.NewTokenRepository(db)
	deptRepo := repositories.NewDepartmentRepository(db)
	desigRepo := repositories.NewDesignationRepository(db)
	empRepo := repositories.NewEmployeeRepository(db)
	attRepo := repositories.NewAttendanceRepository(db)
	leaveRepo := repositories.NewLeaveRepository(db)
	leaveTypeRepo := repositories.NewLeaveTypeRepository(db)
	leaveBalRepo := repositories.NewLeaveBalanceRepository(db)
	holidayRepo := repositories.NewHolidayRepository(db)
	structureRepo := repositories.NewSalaryStructureRepository(db)
	payrollRepo := repositories.NewPayrollRepository(db)
	jobRepo := repositories.NewJobPostRepository(db)
	candRepo := repositories.NewCandidateRepository(db)
	appRepo := repositories.NewApplicationRepository(db)
	interviewRepo := repositories.NewInterviewRepository(db)
	onboardRepo := repositories.NewOnboardingRepository(db)
	goalRepo := repositories.NewGoalRepository(db)
	kpiRepo := repositories.NewKpiRepository(db)
	reviewRepo := repositories.NewReviewRepository(db)
	trainRepo := repositories.NewTrainingRepository(db)
	schedRepo := repositories.NewTrainingScheduleRepository(db)
	enrollRepo := repositories.NewEnrollmentRepository(db)
	docRepo := repositories.NewDocumentRepository(db)
	auditLogRepo := repositories.NewAuditLogRepository(db)
	tenantRepo := repositories.NewTenantRepository(db)

	// --- Services ---
	authSvc := services.NewAuthService(userRepo, tokenRepo, tenantRepo, jwt, cfg, audit)
	userSvc := services.NewUserService(userRepo, roleRepo, audit)
	roleSvc := services.NewRoleService(roleRepo, permRepo, audit)
	empSvc := services.NewEmployeeService(empRepo, deptRepo, desigRepo, audit)
	deptSvc := services.NewDepartmentService(deptRepo, empRepo, audit)
	desigSvc := services.NewDesignationService(desigRepo, audit)
	attSvc := services.NewAttendanceService(attRepo, empRepo, audit)
	holidaySvc := services.NewHolidayService(holidayRepo, audit)
	leaveSvc := services.NewLeaveService(leaveRepo, leaveTypeRepo, leaveBalRepo, empRepo, notify, audit)
	salarySvc := services.NewSalaryService(structureRepo, empRepo, audit)
	payrollSvc := services.NewPayrollService(payrollRepo, structureRepo, audit)
	recruitSvc := services.NewRecruitmentService(jobRepo, candRepo, appRepo, interviewRepo, onboardRepo, empSvc, userRepo, roleRepo, notify, audit)
	perfSvc := services.NewPerformanceService(goalRepo, kpiRepo, reviewRepo, empRepo, audit)
	trainSvc := services.NewTrainingService(trainRepo, schedRepo, enrollRepo, empRepo, notify, audit)
	docSvc := services.NewDocumentService(docRepo, empRepo, cfg, audit)
	reportSvc := services.NewReportService(db)
	tenantSvc := services.NewTenantService(db, tenantRepo, userRepo, roleRepo, audit)
	googleFormSvc := services.NewGoogleFormService(googleFormRepo, googleFormRespRepo, oauthTokenRepo, sheetsClient, tokenManager, googleCfg.SuccessRedirect, recruitSvc, notify, audit)

	router := routes.NewRouter(routes.Deps{
		JWT:          jwt,
		UserRepo:     userRepo,
		EmployeeRepo: empRepo,
		TenantRepo:   tenantRepo,

		AccessSvc:      authSvc,
		UserSvc:        userSvc,
		RoleSvc:        roleSvc,
		DepartmentSvc:  deptSvc,
		DesignationSvc: desigSvc,
		EmployeeSvc:    empSvc,
		AttendanceSvc:  attSvc,
		HolidaySvc:     holidaySvc,
		LeaveSvc:       leaveSvc,
		SalarySvc:      salarySvc,
		PayrollSvc:     payrollSvc,
		RecruitSvc:     recruitSvc,
		GoogleFormSvc:  googleFormSvc,
		PerformanceSvc: perfSvc,
		TrainingSvc:    trainSvc,
		DocumentSvc:    docSvc,
		ReportSvc:      reportSvc,
		AuditLogs:      auditLogRepo,
		Notify:         notify,
		TenantSvc:      tenantSvc,
	})

	slog.Info("server starting", "port", cfg.Port, "env", cfg.Environment)
	if err := router.Run(":" + cfg.Port); err != nil {
		slog.Error("server", "err", err)
		os.Exit(1)
	}
}
