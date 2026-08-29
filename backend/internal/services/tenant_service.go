package services

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/emplyra/backend/internal/auditmanager"
	"github.com/emplyra/backend/internal/auth"
	"github.com/emplyra/backend/internal/models"
	"github.com/emplyra/backend/internal/repositories"
	"github.com/emplyra/backend/internal/utils"
)

var (
	ErrTenantNotFound    = errors.New("tenant not found")
	ErrTenantSlugTaken   = errors.New("tenant slug is already in use")
	ErrRoleScopeMismatch = errors.New("role is not valid for this scope")
	ErrTrialDaysInvalid  = errors.New("trial_days must be positive when plan is TRIAL")
)

type TenantService struct {
	db      *gorm.DB
	tenants *repositories.TenantRepository
	users   *repositories.UserRepository
	roles   *repositories.RoleRepository
	audit   *auditmanager.Service
}

func NewTenantService(db *gorm.DB, tenants *repositories.TenantRepository, users *repositories.UserRepository, roles *repositories.RoleRepository, audit *auditmanager.Service) *TenantService {
	return &TenantService{db: db, tenants: tenants, users: users, roles: roles, audit: audit}
}

var slugRe = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func (s *TenantService) List(p utils.Pagination, search, status string) ([]models.Tenant, int64, error) {
	return s.tenants.List(p, search, status)
}

func (s *TenantService) Get(id string) (*models.Tenant, models.TenantUsage, error) {
	t, err := s.tenants.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, models.TenantUsage{}, ErrTenantNotFound
		}
		return nil, models.TenantUsage{}, err
	}
	usage, err := s.tenants.Usage(id)
	if err != nil {
		return nil, models.TenantUsage{}, err
	}
	return t, usage, nil
}

// Register creates a TRIAL tenant with an owner account, self-serve and public.
func (s *TenantService) Register(in struct {
	Name, Email, Password, FirstName, LastName string
	TrialDays                                  int
}, ip, ua string) (*models.Tenant, error) {
	slug := slugify(in.Name)
	if slug == "" {
		slug = randomSlug()
	}
	if in.TrialDays <= 0 {
		in.TrialDays = 14
	}
	return s.Create(struct {
		Name, Slug, Plan, Industry                                              string
		TrialDays                                                               int
		OwnerEmail, OwnerPassword, OwnerFirstName, OwnerLastName, OwnerUsername string
	}{
		Name: in.Name, Slug: slug, Plan: string(models.PlanFree),
		TrialDays:  in.TrialDays,
		OwnerEmail: in.Email, OwnerPassword: in.Password,
		OwnerFirstName: in.FirstName, OwnerLastName: in.LastName,
	}, "", ip, ua)
}

func (s *TenantService) Create(in struct {
	Name, Slug, Plan, Industry                                              string
	TrialDays                                                               int
	OwnerEmail, OwnerPassword, OwnerFirstName, OwnerLastName, OwnerUsername string
}, actorID, ip, ua string) (*models.Tenant, error) {
	if !slugRe.MatchString(in.Slug) {
		return nil, ErrTenantSlugTaken
	}
	if _, err := s.tenants.FindBySlug(in.Slug); err == nil {
		return nil, ErrTenantSlugTaken
	}
	plan := models.TenantPlan(strings.ToUpper(strings.TrimSpace(in.Plan)))
	switch plan {
	case "":
		plan = models.PlanFree
	case models.PlanFree, models.PlanProfessional, models.PlanEnterprise:
	default:
		return nil, ErrRoleScopeMismatch
	}
	status := models.TenantActive
	var trialEndsAt *time.Time
	if plan == models.PlanFree && in.TrialDays > 0 {
		status = models.TenantTrial
		t := time.Now().AddDate(0, 0, in.TrialDays)
		trialEndsAt = &t
	}

	txErr := s.db.Transaction(func(tx *gorm.DB) error {
		tTenants := repositories.NewTenantRepository(tx)
		tUsers := repositories.NewUserRepository(tx)
		tRoles := repositories.NewRoleRepository(tx)
		tenant := &models.Tenant{
			Name:        in.Name,
			Slug:        in.Slug,
			Status:      status,
			Plan:        plan,
			TrialEndsAt: trialEndsAt,
			Industry:    in.Industry,
		}
		if actorID != "" {
			tenant.CreatedBy = &actorID
		}
		if err := tTenants.Create(tenant); err != nil {
			return err
		}
		if in.OwnerEmail != "" || in.OwnerPassword != "" {
			if _, err := s.createTenantUser(tUsers, tRoles, tenant.ID, in.OwnerUsername, in.OwnerEmail, in.OwnerPassword, in.OwnerFirstName, in.OwnerLastName, models.RoleTenantOwner); err != nil {
				return err
			}
		}
		return nil
	})
	if txErr != nil {
		return nil, txErr
	}
	created, _ := s.tenants.FindBySlug(in.Slug)
	if actorID != "" {
		s.audit.Record(actorID, models.ActionCreate, "tenant", created.ID, ip, ua, map[string]string{"slug": created.Slug})
	}
	return created, nil
}

func (s *TenantService) createTenantUser(tUsers *repositories.UserRepository, tRoles *repositories.RoleRepository, tenantID string, username, email, password, firstName, lastName string, roleName models.RoleName) (*models.User, error) {
	email = auth.NormalizeEmail(email)
	if email == "" {
		return nil, ErrTenantNotFound
	}
	if username == "" {
		username = userUsernameFromEmail(email)
	}
	role, err := tRoles.FindByName(string(roleName))
	if err != nil {
		return nil, ErrRoleScopeMismatch
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return nil, err
	}
	u := &models.User{
		Username:     username,
		Email:        email,
		PasswordHash: hash,
		FirstName:    firstName,
		LastName:     lastName,
		Status:       models.UserStatusActive,
		TenantID:     &tenantID,
		RoleID:       role.ID,
	}
	if err := tUsers.Create(u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *TenantService) CreateOwner(tenantID string, in struct {
	Email, Password, FirstName, LastName string
}, actorID, ip, ua string) (*models.User, error) {
	if _, err := s.tenants.FindByID(tenantID); err != nil {
		return nil, ErrTenantNotFound
	}
	var owner *models.User
	err := s.db.Transaction(func(tx *gorm.DB) error {
		tUsers := repositories.NewUserRepository(tx)
		tRoles := repositories.NewRoleRepository(tx)
		o, err := s.createTenantUser(tUsers, tRoles, tenantID, "", in.Email, in.Password, in.FirstName, in.LastName, models.RoleTenantOwner)
		if err != nil {
			return err
		}
		owner = o
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.audit.Record(actorID, models.ActionCreate, "tenant_owner", tenantID, ip, ua, map[string]string{"email": owner.Email})
	return owner, nil
}

func (s *TenantService) Update(id string, in struct {
	Name, Plan, Industry string
}, actorID, ip, ua string) (*models.Tenant, error) {
	t, err := s.tenants.FindByID(id)
	if err != nil {
		return nil, ErrTenantNotFound
	}
	fields := map[string]interface{}{}
	if in.Name != "" {
		fields["name"] = in.Name
	}
	if in.Industry != "" {
		fields["industry"] = in.Industry
	}
	if in.Plan != "" {
		plan := models.TenantPlan(strings.ToUpper(strings.TrimSpace(in.Plan)))
		switch plan {
		case models.PlanFree, models.PlanProfessional, models.PlanEnterprise:
			fields["plan"] = plan
		default:
			return nil, ErrRoleScopeMismatch
		}
	}
	if len(fields) == 0 {
		return t, nil
	}
	if err := s.tenants.Update(id, fields); err != nil {
		return nil, err
	}
	s.audit.Record(actorID, models.ActionUpdate, "tenant", id, ip, ua, nil)
	return s.tenants.FindByID(id)
}

func (s *TenantService) SetStatus(id string, status models.TenantStatus, actorID, ip, ua string) (*models.Tenant, error) {
	if _, err := s.tenants.FindByID(id); err != nil {
		return nil, ErrTenantNotFound
	}
	if err := s.tenants.Update(id, map[string]interface{}{"status": status}); err != nil {
		return nil, err
	}
	act := models.ActionUpdate
	if status == models.TenantSuspended {
		act = models.ActionOther
	}
	s.audit.Record(actorID, act, "tenant", id, ip, ua, map[string]string{"status": string(status)})
	return s.tenants.FindByID(id)
}

func (s *TenantService) Activate(id, actorID, ip, ua string) (*models.Tenant, error) {
	return s.SetStatus(id, models.TenantActive, actorID, ip, ua)
}

func (s *TenantService) Suspend(id, actorID, ip, ua string) (*models.Tenant, error) {
	return s.SetStatus(id, models.TenantSuspended, actorID, ip, ua)
}

func (s *TenantService) PlatformUsersList(p utils.Pagination, search string) ([]models.User, int64, error) {
	paginate := func(q *gorm.DB) *gorm.DB { return q.Offset(p.Offset).Limit(p.Limit) }
	filter := func(q *gorm.DB) *gorm.DB {
		if search != "" {
			like := "%" + strings.ToLower(search) + "%"
			return q.Where("LOWER(username) LIKE ? OR LOWER(email) LIKE ? OR LOWER(first_name) LIKE ? OR LOWER(last_name) LIKE ?", like, like, like, like)
		}
		return q
	}
	return s.users.ListPlatform(paginate, filter)
}

func (s *TenantService) resolveRole(ref string) (*models.Role, error) {
	role, err := s.roles.FindByID(ref)
	if err != nil {
		role, err = s.roles.FindByName(ref)
		if err != nil {
			return nil, err
		}
	}
	return role, nil
}

func (s *TenantService) CreatePlatformUser(in struct {
	Username, Email, Password, FirstName, LastName, Role, Status string
}, actorID, ip, ua string) (*models.User, error) {
	role, err := s.resolveRole(in.Role)
	if err != nil {
		return nil, ErrRoleScopeMismatch
	}
	if !isPlatformRole(role) {
		return nil, ErrRoleScopeMismatch
	}
	email := auth.NormalizeEmail(in.Email)
	if u, _ := s.users.FindByEmail(email); u != nil {
		return nil, ErrDuplicate
	}
	hash, err := auth.HashPassword(in.Password)
	if err != nil {
		return nil, err
	}
	status := models.UserStatus(in.Status)
	if status == "" {
		status = models.UserStatusActive
	}
	user := &models.User{
		Username:     in.Username,
		Email:        email,
		PasswordHash: hash,
		FirstName:    in.FirstName,
		LastName:     in.LastName,
		RoleID:       role.ID,
		Status:       status,
	}
	if err := s.users.Create(user); err != nil {
		return nil, err
	}
	s.audit.Record(actorID, models.ActionCreate, "platform_user", user.ID, ip, ua, map[string]string{"username": user.Username})
	return user, nil
}

func (s *TenantService) UpdatePlatformUser(id string, in struct {
	FirstName, LastName, Role, Status string
}, actorID, ip, ua string) (*models.User, error) {
	user, err := s.users.FindByID(id)
	if err != nil {
		return nil, ErrNotFound
	}
	if user.TenantID != nil {
		return nil, ErrNotFound
	}
	fields := map[string]interface{}{}
	if in.FirstName != "" {
		fields["first_name"] = in.FirstName
	}
	if in.LastName != "" {
		fields["last_name"] = in.LastName
	}
	if in.Role != "" {
		role, err := s.resolveRole(in.Role)
		if err != nil || !isPlatformRole(role) {
			return nil, ErrRoleScopeMismatch
		}
		fields["role_id"] = role.ID
	}
	if in.Status != "" {
		fields["status"] = in.Status
	}
	if len(fields) > 0 {
		if err := s.users.Update(id, fields); err != nil {
			return nil, err
		}
		s.audit.Record(actorID, models.ActionUpdate, "platform_user", id, ip, ua, nil)
	}
	return s.users.FindByID(id)
}

func (s *TenantService) DeletePlatformUser(id, actorID, ip, ua string) error {
	user, err := s.users.FindByID(id)
	if err != nil || user.TenantID != nil {
		return ErrNotFound
	}
	if err := s.users.Delete(id); err != nil {
		return err
	}
	s.audit.Record(actorID, models.ActionDelete, "platform_user", id, ip, ua, nil)
	return nil
}

// Dashboard returns cross-tenant platform metrics (no HR/employee detail).
func (s *TenantService) Dashboard() (map[models.TenantStatus]int64, int64, int64, int64, error) {
	byStatus, err := s.tenants.CountByStatus()
	if err != nil {
		return nil, 0, 0, 0, err
	}
	totalTenants, err := s.tenants.Count()
	if err != nil {
		return nil, 0, 0, 0, err
	}
	var totalUsers int64
	if err := s.db.Model(&models.User{}).Where("tenant_id IS NOT NULL").Count(&totalUsers).Error; err != nil {
		return nil, 0, 0, 0, err
	}
	var totalEmployees int64
	if err := s.db.Model(&models.Employee{}).Count(&totalEmployees).Error; err != nil {
		return nil, 0, 0, 0, err
	}
	return byStatus, totalTenants, totalUsers, totalEmployees, nil
}

func userUsernameFromEmail(email string) string {
	out := []rune(email[:maxIndex(email)])
	for i, r := range out {
		if !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '_' || r == '-') {
			out[i] = '_'
		}
	}
	return string(out)
}

func slugify(s string) string {
	var b strings.Builder
	prevDash := false
	for _, r := range strings.ToLower(strings.TrimSpace(s)) {
		ok := r >= 'a' && r <= 'z' || r >= '0' && r <= '9'
		if ok {
			b.WriteRune(r)
			prevDash = false
			continue
		}
		if !prevDash {
			b.WriteRune('-')
			prevDash = true
		}
	}
	out := strings.Trim(b.String(), "-")
	if len(out) > 120 {
		out = out[:120]
	}
	return out
}

func randomSlug() string {
	return "org-" + strconv.FormatInt(time.Now().UnixNano()%1000000, 10)
}

// isPlatformRole reports whether the role may be assigned to a platform user.
func isPlatformRole(r *models.Role) bool {
	if r.Scope == models.RoleScopePlatform {
		return true
	}
	switch models.RoleName(r.Name) {
	case models.RoleSuperAdmin, models.RolePlatformOwner, models.RolePlatformAdmin,
		models.RolePlatformSupport, models.RolePlatformAuditor:
		return true
	}
	return false
}

func maxIndex(s string) int {
	for i := 0; i < len(s); i++ {
		if s[i] == '@' {
			return i
		}
	}
	return len(s)
}
