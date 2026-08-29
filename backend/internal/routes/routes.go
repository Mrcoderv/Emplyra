package routes

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"github.com/emplyra/backend/internal/auth"
	"github.com/emplyra/backend/internal/handlers"
	"github.com/emplyra/backend/internal/middleware"
	"github.com/emplyra/backend/internal/notifications"
	"github.com/emplyra/backend/internal/repositories"
	"github.com/emplyra/backend/internal/services"
)

// Deps carries every dependency the router needs. Wiring happens once in main.
type Deps struct {
	AllowedOrigins []string
	JWT          *auth.JWTManager
	UserRepo     *repositories.UserRepository
	EmployeeRepo *repositories.EmployeeRepository
	TenantRepo   *repositories.TenantRepository

	AccessSvc      *services.AuthService
	UserSvc        *services.UserService
	RoleSvc        *services.RoleService
	DepartmentSvc  *services.DepartmentService
	DesignationSvc *services.DesignationService
	EmployeeSvc    *services.EmployeeService
	AttendanceSvc  *services.AttendanceService
	HolidaySvc     *services.HolidayService
	LeaveSvc       *services.LeaveService
	SalarySvc      *services.SalaryService
	PayrollSvc     *services.PayrollService
	RecruitSvc     *services.RecruitmentService
	GoogleFormSvc  *services.GoogleFormService
	PerformanceSvc *services.PerformanceService
	TrainingSvc    *services.TrainingService
	DocumentSvc    *services.DocumentService
	ReportSvc      *services.ReportService
	AuditLogs      *repositories.AuditLogRepository
	Notify         *notifications.Service
	TenantSvc      *services.TenantService
}

func NewRouter(d Deps) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestLogger())
	r.Use(middleware.SecurityHeaders())
		origins := d.AllowedOrigins
		if len(origins) == 0 {
			origins = []string{"http://localhost:3000"}
		}
		r.Use(cors.New(cors.Config{
			AllowOrigins:     origins,
			AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Tenant-ID"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))
	r.Use(middleware.RequestSizeLimit(50 << 20))
	r.Use(middleware.NewRateLimiter(rate.Limit(20), 60).Middleware())
	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })

	// Auth layer.
	authMW := middleware.Auth(d.JWT)
	resolver := middleware.NewPermissionResolver(d.UserRepo, 30*time.Second)
	rbac := func(perm string) gin.HandlerFunc { return middleware.RBAC(resolver, perm) }

	api := r.Group("/api/v1")

	// --- Public auth endpoints (self-serve) ---
	authHandler := handlers.NewAuthHandler(d.AccessSvc)
	api.POST("/auth/login", authHandler.Login)
	api.POST("/auth/refresh", authHandler.Refresh)

	// --- Google OAuth callback (browser redirect; state-validated) ---
	gfHandler := handlers.NewGoogleFormHandler(d.GoogleFormSvc)
	api.GET("/integrations/google/oauth/callback", gfHandler.OAuthCallback)

	protected := api.Group("", authMW, middleware.EmployeeScope(d.EmployeeRepo), middleware.TenantScope(d.TenantRepo, resolver, "platform:tenant-access"))
	{
		// --- Auth self-service (logged in) ---
		authGroup := protected.Group("/auth")
		authGroup.POST("/logout", authHandler.Logout)
		authGroup.GET("/me", authHandler.Me)

		// --- Users ---
		userHandler := handlers.NewUserHandler(d.UserSvc)
		users := protected.Group("/users", rbac("user:read"))
		users.GET("", userHandler.List)
		users.POST("", rbac("user:create"), userHandler.Create)
		users.GET("/:id", userHandler.Get)
		users.PUT("/:id", rbac("user:update"), userHandler.Update)
		users.DELETE("/:id", rbac("user:delete"), userHandler.Delete)

		// --- Roles ---
		roleHandler := handlers.NewRoleHandler(d.RoleSvc)
		roles := protected.Group("/roles", rbac("role:read"))
		roles.GET("", roleHandler.List)
		roles.GET("/permissions", rbac("permission:read"), roleHandler.Permissions)
		roles.POST("", rbac("role:create"), roleHandler.Create)
		roles.GET("/:id", roleHandler.Get)
		roles.PUT("/:id", rbac("role:update"), roleHandler.Update)
		roles.DELETE("/:id", rbac("role:delete"), roleHandler.Delete)

		// --- Department / designation / employees ---
		deptHandler := handlers.NewDepartmentHandler(d.DepartmentSvc)
		departments := protected.Group("/departments")
		departments.GET("", rbac("department:read"), deptHandler.List)
		departments.POST("", rbac("department:create"), deptHandler.Create)
		departments.GET("/:id", rbac("department:read"), deptHandler.Get)
		departments.PUT("/:id", rbac("department:update"), deptHandler.Update)
		departments.DELETE("/:id", rbac("department:delete"), deptHandler.Delete)

		desigHandler := handlers.NewDesignationHandler(d.DesignationSvc)
		designations := protected.Group("/designations")
		designations.GET("", rbac("designation:read"), desigHandler.List)
		designations.POST("", rbac("designation:create"), desigHandler.Create)
		designations.GET("/:id", rbac("designation:read"), desigHandler.Get)
		designations.PUT("/:id", rbac("designation:update"), desigHandler.Update)
		designations.DELETE("/:id", rbac("designation:delete"), desigHandler.Delete)

		empHandler := handlers.NewEmployeeHandler(d.EmployeeSvc)
		employees := protected.Group("/employees")
		employees.GET("", rbac("employee:read"), empHandler.List)
		employees.POST("", rbac("employee:create"), empHandler.Create)
		employees.GET("/me", empHandler.MyProfile)
		employees.GET("/:id", rbac("employee:read"), empHandler.Get)
		employees.PUT("/:id", rbac("employee:update"), empHandler.Update)
		employees.DELETE("/:id", rbac("employee:delete"), empHandler.Delete)

		// --- Attendance ---
		attHandler := handlers.NewAttendanceHandler(d.AttendanceSvc)
		attendance := protected.Group("/attendance")
		attendance.POST("/check-in", rbac("attendance:create"), attHandler.CheckIn)
		attendance.POST("/check-out", rbac("attendance:create"), attHandler.CheckOut)
		attendance.GET("", rbac("attendance:read"), attHandler.List)
		attendance.GET("/:id", rbac("attendance:read"), attHandler.Get)
		attendance.PUT("/:id", rbac("attendance:update"), attHandler.Update)

		// --- Holidays ---
		holHandler := handlers.NewHolidayHandler(d.HolidaySvc)
		holidays := protected.Group("/holidays")
		holidays.GET("", rbac("holiday:read"), holHandler.List)
		holidays.POST("", rbac("holiday:create"), holHandler.Create)
		holidays.GET("/:id", rbac("holiday:read"), holHandler.Get)
		holidays.PUT("/:id", rbac("holiday:update"), holHandler.Update)
		holidays.DELETE("/:id", rbac("holiday:delete"), holHandler.Delete)

		// --- Leaves ---
		leaveHandler := handlers.NewLeaveHandler(d.LeaveSvc)
		leaves := protected.Group("/leaves")
		leaves.GET("/types", rbac("leave:read"), leaveHandler.Types)
		leaves.GET("/balances", rbac("leave:read"), leaveHandler.Balances)
		leaves.POST("/balances/set", rbac("leave:create"), leaveHandler.SetBalance)
		leaves.POST("/balance", rbac("leave:create"), leaveHandler.SetBalance)
		leaves.GET("", rbac("leave:read"), leaveHandler.List)
		leaves.POST("", rbac("leave:create"), leaveHandler.Create)
		leaves.GET("/:id", rbac("leave:read"), leaveHandler.Get)
		leaves.PUT("/:id/approve", rbac("leave:approve"), leaveHandler.Approve)
		leaves.PUT("/:id/reject", rbac("leave:reject"), leaveHandler.Reject)

		// --- Salary structures ---
		payHandler := handlers.NewPayrollHandler(d.SalarySvc, d.PayrollSvc)
		salary := protected.Group("/salary")
		salary.GET("", rbac("salary:read"), payHandler.ListStructures)
		salary.POST("", rbac("salary:create"), payHandler.CreateStructure)
		salary.GET("/:id", rbac("salary:read"), payHandler.GetStructure)
		salary.PUT("/:id", rbac("salary:update"), payHandler.UpdateStructure)
		salary.DELETE("/:id", rbac("salary:delete"), payHandler.DeleteStructure)

		// --- Payroll ---
		payroll := protected.Group("/payroll")
		payroll.GET("", rbac("payroll:read"), payHandler.List)
		payroll.POST("/generate", rbac("payroll:create"), payHandler.Generate)
		payroll.POST("/process", rbac("payroll:process"), payHandler.Process)
		payroll.POST("/mark-paid", rbac("payroll:pay"), payHandler.MarkPaid)
		payroll.POST("/cancel", rbac("payroll:cancel"), payHandler.Cancel)
		payroll.GET("/:id", rbac("payroll:read"), payHandler.Get)
		payroll.GET("/:id/payslip", rbac("payroll:payslip"), payHandler.Payslip)

		// --- Recruitment ---
		recruitHandler := handlers.NewRecruitmentHandler(d.RecruitSvc)
		jobs := protected.Group("/recruitment/jobs")
		jobs.GET("", rbac("job:read"), recruitHandler.ListJobs)
		jobs.POST("", rbac("job:create"), recruitHandler.CreateJob)
		jobs.GET("/:id", rbac("job:read"), recruitHandler.GetJob)
		jobs.PUT("/:id", rbac("job:update"), recruitHandler.UpdateJob)
		jobs.DELETE("/:id", rbac("job:delete"), recruitHandler.DeleteJob)

		// --- Google Forms integration (recruitment) ---
		jobs.POST("/:id/google-form/connect", rbac("googleform:connect"), gfHandler.Connect)
		jobs.GET("/:id/google-form", rbac("googleform:read"), gfHandler.Get)
		jobs.PUT("/:id/google-form", rbac("googleform:connect"), gfHandler.Update)
		jobs.DELETE("/:id/google-form", rbac("googleform:connect"), gfHandler.Disconnect)
		jobs.POST("/:id/google-form/sync", rbac("googleform:sync"), gfHandler.Sync)
		jobs.GET("/:id/google-form/sync-status", rbac("googleform:read"), gfHandler.SyncStatus)
		jobs.GET("/:id/google-form/responses", rbac("googleform:read"), gfHandler.Responses)

		// --- Google OAuth (admin-only account linking) ---
		oauth := protected.Group("/integrations/google/oauth")
		oauth.POST("/authorize", rbac("googleform:connect"), gfHandler.OAuthAuthorize)

		candidates := protected.Group("/recruitment/candidates")
		candidates.GET("", rbac("candidate:read"), recruitHandler.ListCandidates)
		candidates.POST("", rbac("candidate:create"), recruitHandler.CreateCandidate)
		candidates.GET("/:id", rbac("candidate:read"), recruitHandler.GetCandidate)
		candidates.PUT("/:id", rbac("candidate:update"), recruitHandler.UpdateCandidate)
		candidates.DELETE("/:id", rbac("candidate:delete"), recruitHandler.DeleteCandidate)

		applications := protected.Group("/recruitment/applications")
		applications.GET("", rbac("application:read"), recruitHandler.ListApplications)
		applications.POST("", rbac("application:create"), recruitHandler.CreateApplication)
		applications.GET("/:id", rbac("application:read"), recruitHandler.GetApplication)
		applications.PUT("/:id/status", rbac("application:update"), recruitHandler.UpdateApplicationStatus)

		interviews := protected.Group("/recruitment/interviews")
		interviews.GET("", rbac("interview:read"), recruitHandler.ListInterviews)
		interviews.POST("", rbac("interview:create"), recruitHandler.ScheduleInterview)
		interviews.GET("/:id", rbac("interview:read"), recruitHandler.GetInterview)
		interviews.PUT("/:id", rbac("interview:update"), recruitHandler.CompleteInterview)

		onboarding := protected.Group("/recruitment/onboarding")
		onboarding.GET("", rbac("onboarding:read"), recruitHandler.ListOnboardings)
		onboarding.POST("", rbac("onboarding:create"), recruitHandler.CreateOnboarding)
		onboarding.GET("/:id", rbac("onboarding:read"), recruitHandler.GetOnboarding)
		onboarding.PUT("/:id", rbac("onboarding:update"), recruitHandler.UpdateOnboarding)
		onboarding.POST("/hire", rbac("onboarding:create"), recruitHandler.Hire)

		// --- Performance ---
		perfHandler := handlers.NewPerformanceHandler(d.PerformanceSvc, d.EmployeeRepo)
		performance := protected.Group("/performance")
		performance.GET("/goals", rbac("goal:read"), perfHandler.ListGoals)
		performance.POST("/goals", rbac("goal:create"), perfHandler.CreateGoal)
		performance.GET("/goals/:id", rbac("goal:read"), perfHandler.GetGoal)
		performance.PUT("/goals/:id", rbac("goal:update"), perfHandler.UpdateGoal)
		performance.DELETE("/goals/:id", rbac("goal:delete"), perfHandler.DeleteGoal)

		performance.GET("/kpis", rbac("kpi:read"), perfHandler.ListKPIs)
		performance.POST("/kpis", rbac("kpi:create"), perfHandler.CreateKPI)
		performance.GET("/kpis/:id", rbac("kpi:read"), perfHandler.GetKPI)
		performance.PUT("/kpis/:id", rbac("kpi:update"), perfHandler.UpdateKPI)
		performance.DELETE("/kpis/:id", rbac("kpi:delete"), perfHandler.DeleteKPI)

		performance.GET("/reviews", rbac("review:read"), perfHandler.ListReviews)
		performance.POST("/reviews", rbac("review:create"), perfHandler.CreateReview)
		performance.GET("/reviews/:id", rbac("review:read"), perfHandler.GetReview)
		performance.PUT("/reviews/:id/submit", rbac("review:submit"), perfHandler.SubmitReview)

		// --- Training ---
		trainHandler := handlers.NewTrainingHandler(d.TrainingSvc)
		training := protected.Group("/training")
		training.GET("/programs", rbac("training:read"), trainHandler.ListPrograms)
		training.POST("/programs", rbac("training:create"), trainHandler.CreateProgram)
		training.GET("/programs/:id", rbac("training:read"), trainHandler.GetProgram)
		training.PUT("/programs/:id", rbac("training:update"), trainHandler.UpdateProgram)
		training.DELETE("/programs/:id", rbac("training:delete"), trainHandler.DeleteProgram)

		training.GET("/schedules", rbac("training:read"), trainHandler.ListSchedules)
		training.POST("/schedules", rbac("training:create"), trainHandler.CreateSchedule)
		training.PUT("/schedules/:id", rbac("training:update"), trainHandler.UpdateSchedule)
		training.DELETE("/schedules/:id", rbac("training:delete"), trainHandler.DeleteSchedule)

		training.GET("/enrollments", rbac("enrollment:read"), trainHandler.ListEnrollments)
		training.POST("/enrollments", rbac("enrollment:create"), trainHandler.Enroll)
		training.PUT("/enrollments/:id", rbac("enrollment:update"), trainHandler.UpdateEnrollment)

		// --- Documents ---
		docHandler := handlers.NewDocumentHandler(d.DocumentSvc, d.EmployeeRepo)
		documents := protected.Group("/documents")
		documents.GET("", rbac("document:read"), docHandler.List)
		documents.POST("", rbac("document:create"), docHandler.Upload)
		documents.GET("/:id", rbac("document:read"), docHandler.Get)
		documents.GET("/:id/download", rbac("document:download"), docHandler.Download)
		documents.DELETE("/:id", rbac("document:delete"), docHandler.Delete)

		// --- Notifications (own scope) ---
		notifHandler := handlers.NewNotificationHandler(d.Notify)
		notifications := protected.Group("/notifications")
		notifications.GET("", rbac("notification:read"), notifHandler.List)
		notifications.GET("/unread-count", rbac("notification:read"), notifHandler.UnreadCount)
		notifications.PUT("/read-all", rbac("notification:update"), notifHandler.MarkAllRead)
		notifications.PUT("/:id/read", rbac("notification:update"), notifHandler.MarkRead)

		// --- Reports & dashboard ---
		reportHandler := handlers.NewReportHandler(d.ReportSvc, d.EmployeeRepo)
		reports := protected.Group("/reports", rbac("report:read"))
		reports.GET("/headcount", reportHandler.Headcount)
		reports.GET("/attendance", reportHandler.Attendance)
		reports.GET("/leaves", reportHandler.Leaves)
		reports.GET("/payroll", reportHandler.Payroll)
		reports.GET("/recruitment", reportHandler.Recruitment)
		reports.GET("/holidays", reportHandler.Holidays)

		dashboard := protected.Group("/dashboard")
		dashboard.GET("/summary", rbac("dashboard:read"), reportHandler.Dashboard)

		// --- Audit logs ---
		auditHandler := handlers.NewAuditHandler(d.AuditLogs)
		audit := protected.Group("/audit", rbac("audit:read"))
		audit.GET("/logs", auditHandler.List)
	}

	// --- Platform admin / tenant management (platform operators only) ---
	if d.TenantSvc != nil {
		tenantHandler := handlers.NewTenantHandler(d.TenantSvc)
		admin := api.Group("/admin", authMW)
		{
			admin.GET("/dashboard", rbac("platform:dashboard:read"), tenantHandler.Dashboard)

			tenants := admin.Group("/tenants")
			tenants.GET("", rbac("tenant:read"), tenantHandler.List)
			tenants.POST("", rbac("tenant:create"), tenantHandler.Create)
			tenants.GET("/:id", rbac("tenant:read"), tenantHandler.Get)
			tenants.PUT("/:id", rbac("tenant:update"), tenantHandler.Update)
			tenants.POST("/:id/activate", rbac("tenant:update"), tenantHandler.Activate)
			tenants.POST("/:id/suspend", rbac("tenant:update"), tenantHandler.Suspend)
			tenants.POST("/:id/owners", rbac("tenant:update"), tenantHandler.CreateOwner)

			pUsers := admin.Group("/users")
			pUsers.GET("", rbac("platform:user:read"), tenantHandler.PlatformUsers)
			pUsers.POST("", rbac("platform:user:create"), tenantHandler.CreatePlatformUser)
			pUsers.PUT("/:id", rbac("platform:user:update"), tenantHandler.UpdatePlatformUser)
			pUsers.DELETE("/:id", rbac("platform:user:delete"), tenantHandler.DeletePlatformUser)
		}

		// --- Self-serve tenant registration (public, rate-limited) ---
		api.POST("/public/tenants", tenantHandler.Register)
	}

	return r
}
