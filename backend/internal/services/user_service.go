package services

import (
	"errors"
	"strings"

	"github.com/emplyra/backend/internal/auditmanager"
	"github.com/emplyra/backend/internal/auth"
	"github.com/emplyra/backend/internal/models"
	"github.com/emplyra/backend/internal/repositories"
	"github.com/emplyra/backend/internal/utils"
	"gorm.io/gorm"
)

type UserService struct {
	users *repositories.UserRepository
	roles *repositories.RoleRepository
	audit *auditmanager.Service
}

func NewUserService(users *repositories.UserRepository, roles *repositories.RoleRepository, audit *auditmanager.Service) *UserService {
	return &UserService{users: users, roles: roles, audit: audit}
}

func (s *UserService) List(p utils.Pagination, search string) ([]models.User, int64, error) {
	paginate := func(q *gorm.DB) *gorm.DB { return q.Offset(p.Offset).Limit(p.Limit) }
	filter := func(q *gorm.DB) *gorm.DB {
		if search != "" {
			like := "%" + strings.ToLower(search) + "%"
			return q.Where("LOWER(username) LIKE ? OR LOWER(email) LIKE ? OR LOWER(first_name) LIKE ? OR LOWER(last_name) LIKE ?", like, like, like, like)
		}
		return q
	}
	return s.users.List(paginate, filter)
}

func (s *UserService) Create(in struct {
	Username  string
	Email     string
	Password  string
	FirstName string
	LastName  string
	RoleID    string
	Status    string
}, callerScope, actorID, ip, ua string) (*models.User, error) {
	email := auth.NormalizeEmail(in.Email)
	if u, err := s.users.FindByEmail(email); err == nil && u != nil {
		return nil, ErrDuplicate
	}
	if u, err := s.users.FindByUsername(in.Username); err == nil && u != nil {
		return nil, ErrDuplicate
	}
	if err := s.ensureAssignableRole(in.RoleID, callerScope); err != nil {
		return nil, err
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
		RoleID:       in.RoleID,
		Status:       status,
	}
	if err := s.users.Create(user); err != nil {
		return nil, err
	}
	s.audit.Record(actorID, models.ActionCreate, "user", user.ID, ip, ua, map[string]string{"username": user.Username})
	return user, nil
}

func (s *UserService) Update(id string, in struct {
	FirstName string
	LastName  string
	RoleID    string
	Status    string
}, callerScope, actorID, ip, ua string) (*models.User, error) {
	user, err := s.users.FindByID(id)
	if err != nil {
		return nil, ErrNotFound
	}
	fields := map[string]interface{}{}
	if in.FirstName != "" {
		fields["first_name"] = in.FirstName
	}
	if in.LastName != "" {
		fields["last_name"] = in.LastName
	}
	if in.RoleID != "" {
		if err := s.ensureAssignableRole(in.RoleID, callerScope); err != nil {
			return nil, err
		}
		if user.RoleID != in.RoleID {
			s.audit.Record(actorID, models.ActionRoleChange, "user", user.ID, ip, ua, map[string]string{"from_role": user.RoleID, "to_role": in.RoleID})
		}
		fields["role_id"] = in.RoleID
	}
	if in.Status != "" {
		fields["status"] = in.Status
	}
	if len(fields) > 0 {
		if err := s.users.Update(id, fields); err != nil {
			return nil, err
		}
		s.audit.Record(actorID, models.ActionUpdate, "user", id, ip, ua, nil)
	}
	return s.users.FindByID(id)
}

func (s *UserService) Delete(id, actorID, ip, ua string) error {
	if err := s.users.Delete(id); err != nil {
		return ErrNotFound
	}
	s.audit.Record(actorID, models.ActionDelete, "user", id, ip, ua, nil)
	return nil
}

func (s *UserService) ChangePassword(userID, current, newPassword, ip, ua string) error {
	user, err := s.users.FindByID(userID)
	if err != nil {
		return ErrNotFound
	}
	if !auth.VerifyPassword(user.PasswordHash, current) {
		return errors.New("current password is incorrect")
	}
	hash, err := auth.HashPassword(newPassword)
	if err != nil {
		return err
	}
	if err := s.users.Update(userID, map[string]interface{}{"password_hash": hash}); err != nil {
		return err
	}
	s.audit.Record(userID, models.ActionUpdate, "user", userID, ip, ua, map[string]string{"field": "password"})
	return nil
}

func (s *UserService) GetByID(id string) (*models.User, error) {
	u, err := s.users.FindByID(id)
	if err != nil {
		return nil, ErrNotFound
	}
	return u, nil
}

// ensureAssignableRole verifies a role may be assigned by a caller of the given
// scope. Platform callers may assign any role; tenant callers may only assign
// tenant-scope roles (platform roles are off-limits to tenant administrators).
func (s *UserService) ensureAssignableRole(roleID, callerScope string) error {
	role, err := s.roles.FindByID(roleID)
	if err != nil {
		return ErrNotFound
	}
	if callerScope == models.RoleScopePlatform {
		return nil
	}
	if role.Scope != models.RoleScopeTenant {
		return ErrRoleScopeMismatch
	}
	return nil
}
