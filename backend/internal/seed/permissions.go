package seed

import "github.com/emplyra/backend/internal/models"

type permissionDef struct {
	Name        string
	Module      string
	Description string
}

var catalog = []permissionDef{
	{"auth:login", "auth", "Login"},
	{"auth:refresh", "auth", "Refresh access token"},
	{"auth:logout", "auth", "Logout"},

	{"user:create", "user", "Create users"},
	{"user:read", "user", "Read users"},
	{"user:update", "user", "Update users"},
	{"user:delete", "user", "Delete users"},

	{"role:create", "role", "Create roles"},
	{"role:read", "role", "Read roles"},
	{"role:update", "role", "Update roles"},
	{"role:delete", "role", "Delete roles"},
	{"role:assign", "role", "Assign roles to users"},
	{"permission:read", "role", "Read permission catalog"},

	{"employee:create", "employee", "Create employees"},
	{"employee:read", "employee", "Read employees"},
	{"employee:update", "employee", "Update employees"},
	{"employee:delete", "employee", "Delete employees"},

	{"department:create", "department", "Create departments"},
	{"department:read", "department", "Read departments"},
	{"department:update", "department", "Update departments"},
	{"department:delete", "department", "Delete departments"},

	{"designation:create", "designation", "Create designations"},
	{"designation:read", "designation", "Read designations"},
	{"designation:update", "designation", "Update designations"},
	{"designation:delete", "designation", "Delete designations"},

	{"attendance:create", "attendance", "Check in/out"},
	{"attendance:read", "attendance", "Read all attendance"},
	{"attendance:update", "attendance", "Update attendance"},

	{"leave:create", "leave", "Apply for leave"},
	{"leave:read", "leave", "Read all leave requests"},
	{"leave:approve", "leave", "Approve leave"},
	{"leave:reject", "leave", "Reject leave"},

	{"holiday:create", "holiday", "Create holidays"},
	{"holiday:read", "holiday", "Read holidays"},
	{"holiday:update", "holiday", "Update holidays"},
	{"holiday:delete", "holiday", "Delete holidays"},

	{"salary:create", "salary", "Create salary structures"},
	{"salary:read", "salary", "Read salary structures"},
	{"salary:update", "salary", "Update salary structures"},
	{"salary:delete", "salary", "Delete salary structures"},

	{"payroll:create", "payroll", "Create payroll entries"},
	{"payroll:read", "payroll", "Read payroll"},
	{"payroll:process", "payroll", "Process payroll"},
	{"payroll:pay", "payroll", "Mark payroll as paid"},
	{"payroll:cancel", "payroll", "Cancel payroll"},
	{"payroll:payslip", "payroll", "Generate payslips"},

	{"job:create", "recruitment", "Create job posts"},
	{"job:read", "recruitment", "Read job posts"},
	{"job:update", "recruitment", "Update job posts"},
	{"job:delete", "recruitment", "Delete job posts"},

	{"candidate:create", "recruitment", "Create candidates"},
	{"candidate:read", "recruitment", "Read candidates"},
	{"candidate:update", "recruitment", "Update candidates"},
	{"candidate:delete", "recruitment", "Delete candidates"},

	{"application:create", "recruitment", "Create applications"},
	{"application:read", "recruitment", "Read applications"},
	{"application:update", "recruitment", "Update applications"},

	{"interview:create", "recruitment", "Schedule interviews"},
	{"interview:read", "recruitment", "Read interviews"},
	{"interview:update", "recruitment", "Complete interviews"},

	{"onboarding:create", "onboarding", "Create onboarding"},
	{"onboarding:read", "onboarding", "Read onboarding"},
	{"onboarding:update", "onboarding", "Update onboarding"},

	{"goal:create", "performance", "Create goals"},
	{"goal:read", "performance", "Read goals"},
	{"goal:update", "performance", "Update goals"},
	{"goal:delete", "performance", "Delete goals"},

	{"kpi:create", "performance", "Create KPIs"},
	{"kpi:read", "performance", "Read KPIs"},
	{"kpi:update", "performance", "Update KPIs"},
	{"kpi:delete", "performance", "Delete KPIs"},

	{"review:create", "performance", "Create performance reviews"},
	{"review:read", "performance", "Read performance reviews"},
	{"review:update", "performance", "Update performance reviews"},
	{"review:submit", "performance", "Submit self/manager evaluation"},

	{"training:create", "training", "Create training programs"},
	{"training:read", "training", "Read training programs"},
	{"training:update", "training", "Update training programs"},
	{"training:delete", "training", "Delete training programs"},
	{"enrollment:create", "training", "Enroll in training"},
	{"enrollment:read", "training", "Read enrollments"},
	{"enrollment:update", "training", "Update enrollments"},

	{"document:create", "document", "Upload documents"},
	{"document:read", "document", "Read documents"},
	{"document:delete", "document", "Delete documents"},
	{"document:download", "document", "Download documents"},

	{"notification:read", "notification", "Read notifications"},
	{"notification:update", "notification", "Mark notifications read"},

	{"report:read", "report", "Read reports"},
	{"dashboard:read", "dashboard", "Read dashboard"},
	{"audit:read", "audit", "Read audit logs"},
}

func permissionsCatalog() []models.Permission {
	out := make([]models.Permission, 0, len(catalog))
	for _, p := range catalog {
		out = append(out, models.Permission{
			Name:        p.Name,
			Module:      p.Module,
			Description: p.Description,
		})
	}
	return out
}

// AllPermissionNames returns every permission definition name (for SUPER_ADMIN).
func AllPermissionNames() []string {
	out := make([]string, 0, len(catalog))
	for _, p := range catalog {
		out = append(out, p.Name)
	}
	return out
}

// PermissionByName indexes the catalog by name.
func PermissionByName() map[string]string {
	out := make(map[string]string, len(catalog))
	for _, p := range catalog {
		out[p.Name] = p.Module
	}
	return out
}

// RolePermissions maps every system role to its granted permissions.
func RolePermissions() map[models.RoleName][]string {
	return map[models.RoleName][]string{
		models.RoleSuperAdmin: AllPermissionNames(),
		models.RoleHRAdmin: {
			"user:create", "user:read", "user:update", "role:read", "role:assign",
			"employee:create", "employee:read", "employee:update", "employee:delete",
			"department:create", "department:read", "department:update", "department:delete",
			"designation:create", "designation:read", "designation:update", "designation:delete",
			"attendance:create", "attendance:read", "attendance:update",
			"leave:create", "leave:read", "leave:approve", "leave:reject",
			"holiday:create", "holiday:read", "holiday:update", "holiday:delete",
			"salary:create", "salary:read", "salary:update", "salary:delete",
			"payroll:create", "payroll:read", "payroll:process", "payroll:pay", "payroll:cancel", "payroll:payslip",
			"job:create", "job:read", "job:update", "job:delete",
			"candidate:create", "candidate:read", "candidate:update", "candidate:delete",
			"application:create", "application:read", "application:update",
			"interview:create", "interview:read", "interview:update",
			"onboarding:create", "onboarding:read", "onboarding:update",
			"goal:create", "goal:read", "goal:update", "goal:delete",
			"kpi:create", "kpi:read", "kpi:update", "kpi:delete",
			"review:create", "review:read", "review:update",
			"training:create", "training:read", "training:update", "training:delete",
			"enrollment:create", "enrollment:read", "enrollment:update",
			"document:create", "document:read", "document:delete", "document:download",
			"notification:read", "notification:update",
			"report:read", "dashboard:read", "audit:read",
		},
		models.RoleManager: {
			"user:read", "employee:read", "employee:update",
			"department:read",
			"attendance:create", "attendance:read",
			"leave:create", "leave:read", "leave:approve", "leave:reject",
			"holiday:read",
			"goal:create", "goal:read", "goal:update",
			"kpi:create", "kpi:read", "kpi:update",
			"review:create", "review:read", "review:update", "review:submit",
			"training:read", "enrollment:read", "enrollment:create",
			"document:create", "document:read",
			"notification:read", "notification:update",
			"report:read", "dashboard:read",
		},
		models.RoleEmployee: {
			"auth:login", "auth:refresh", "auth:logout",
			"employee:read", "attendance:create", "attendance:read",
			"leave:create", "leave:read",
			"holiday:read",
			"goal:create", "goal:read", "goal:update",
			"kpi:read", "review:read", "review:submit",
			"training:read", "enrollment:create", "enrollment:read",
			"document:create", "document:read", "document:download",
			"notification:read", "notification:update",
			"dashboard:read",
		},
		models.RoleRecruiter: {
			"user:read", "employee:read",
			"department:read",
			"job:create", "job:read", "job:update", "job:delete",
			"candidate:create", "candidate:read", "candidate:update", "candidate:delete",
			"application:create", "application:read", "application:update",
			"interview:create", "interview:read", "interview:update",
			"onboarding:create", "onboarding:read", "onboarding:update",
			"training:read",
			"document:create", "document:read",
			"notification:read", "notification:update",
			"report:read", "dashboard:read",
		},
		models.RoleAccountant: {
			"user:read", "employee:read",
			"department:read",
			"salary:create", "salary:read", "salary:update",
			"payroll:create", "payroll:read", "payroll:process", "payroll:pay", "payroll:cancel", "payroll:payslip",
			"attendance:read",
			"notification:read", "notification:update",
			"report:read", "dashboard:read",
		},
	}
}
