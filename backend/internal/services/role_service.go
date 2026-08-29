package services

import (
	"github.com/emplyra/backend/internal/auditmanager"
	"github.com/emplyra/backend/internal/models"
	"github.com/emplyra/backend/internal/repositories"
)

type RoleService struct {
	roles *repositories.RoleRepository
	perms *repositories.PermissionRepository
	audit *auditmanager.Service
}

func NewRoleService(roles *repositories.RoleRepository, perms *repositories.PermissionRepository, audit *auditmanager.Service) *RoleService {
	return &RoleService{roles: roles, perms: perms, audit: audit}
}

func (s *RoleService) List(callerScope string) ([]models.Role, error) {
	if callerScope == models.RoleScopePlatform {
		return s.roles.List()
	}
	return s.roles.ListByScope(models.RoleScopeTenant)
}

func (s *RoleService) Get(id, callerScope string) (*models.Role, error) {
	r, err := s.roles.FindByID(id)
	if err != nil {
		return nil, ErrNotFound
	}
	if callerScope != models.RoleScopePlatform && r.Scope != models.RoleScopeTenant {
		return nil, ErrNotFound
	}
	return r, nil
}

func (s *RoleService) Create(name, description string, permissionIDs []string, callerScope, actorID, ip, ua string) (*models.Role, error) {
	if existing, _ := s.roles.FindByName(name); existing != nil {
		return nil, ErrDuplicate
	}
	scope := models.RoleScopeTenant
	if callerScope == models.RoleScopePlatform {
		scope = models.RoleScopePlatform
	}
	role := &models.Role{Name: name, Description: description, Scope: scope}
	if err := s.roles.Create(role, permissionIDs); err != nil {
		return nil, err
	}
	s.audit.Record(actorID, models.ActionCreate, "role", role.ID, ip, ua, map[string]string{"name": name, "scope": scope})
	return role, nil
}

func (s *RoleService) Update(id, name, description string, permissionIDs []string, callerScope, actorID, ip, ua string) (*models.Role, error) {
	role, err := s.roles.FindByID(id)
	if err != nil {
		return nil, ErrNotFound
	}
	if callerScope != models.RoleScopePlatform && role.Scope != models.RoleScopeTenant {
		return nil, ErrRoleScopeMismatch
	}
	role.Name = name
	role.Description = description
	if err := s.roles.Update(role, permissionIDs); err != nil {
		return nil, err
	}
	s.audit.Record(actorID, models.ActionPermissionChange, "role", id, ip, ua, map[string]string{"name": name})
	return s.roles.FindByID(id)
}

func (s *RoleService) Delete(id, actorID, ip, ua string, callerScope string) error {
	role, err := s.roles.FindByID(id)
	if err != nil {
		return ErrNotFound
	}
	if callerScope != models.RoleScopePlatform && role.Scope != models.RoleScopeTenant {
		return ErrRoleScopeMismatch
	}
	if err := s.roles.Delete(id); err != nil {
		if err == repositories.ErrRoleInUse {
			return err
		}
		return ErrNotFound
	}
	s.audit.Record(actorID, models.ActionDelete, "role", id, ip, ua, nil)
	return nil
}

func (s *RoleService) Permissions() ([]models.Permission, error) {
	return s.perms.List()
}
