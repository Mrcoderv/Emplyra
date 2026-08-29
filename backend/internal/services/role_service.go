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

func (s *RoleService) List() ([]models.Role, error) {
	return s.roles.List()
}

func (s *RoleService) Get(id string) (*models.Role, error) {
	r, err := s.roles.FindByID(id)
	if err != nil {
		return nil, ErrNotFound
	}
	return r, nil
}

func (s *RoleService) Create(name, description string, permissionIDs []string, actorID, ip, ua string) (*models.Role, error) {
	if existing, _ := s.roles.FindByName(name); existing != nil {
		return nil, ErrDuplicate
	}
	role := &models.Role{Name: name, Description: description}
	if err := s.roles.Create(role, permissionIDs); err != nil {
		return nil, err
	}
	s.audit.Record(actorID, models.ActionCreate, "role", role.ID, ip, ua, map[string]string{"name": name})
	return role, nil
}

func (s *RoleService) Update(id, name, description string, permissionIDs []string, actorID, ip, ua string) (*models.Role, error) {
	role, err := s.roles.FindByID(id)
	if err != nil {
		return nil, ErrNotFound
	}
	role.Name = name
	role.Description = description
	if err := s.roles.Update(role, permissionIDs); err != nil {
		return nil, err
	}
	s.audit.Record(actorID, models.ActionPermissionChange, "role", id, ip, ua, map[string]string{"name": name})
	return s.roles.FindByID(id)
}

func (s *RoleService) Delete(id, actorID, ip, ua string) error {
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
