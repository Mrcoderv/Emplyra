package seed

import (
	"log/slog"

	"gorm.io/gorm"

	"github.com/emplyra/backend/internal/auth"
	"github.com/emplyra/backend/internal/config"
	"github.com/emplyra/backend/internal/models"
)

// Run seeds permissions, system roles and the super admin account.
// It is idempotent: safe to run on every startup.
func Run(db *gorm.DB, cfg *config.Config) error {
	if err := db.Transaction(func(tx *gorm.DB) error {
		return seedPermissionsAndRoles(tx)
	}); err != nil {
		return err
	}
	if cfg.SeedSuperAdmin {
		if err := seedSuperAdmin(db, cfg); err != nil {
			return err
		}
	}
	return nil
}

func seedPermissionsAndRoles(tx *gorm.DB) error {
	for _, p := range permissionsCatalog() {
		var existing models.Permission
		err := tx.Where("name = ?", p.Name).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			if err := tx.Create(&p).Error; err != nil {
				return err
			}
			continue
		}
		if err != nil {
			return err
		}
		existing.Description = p.Description
		existing.Module = p.Module
		if err := tx.Save(&existing).Error; err != nil {
			return err
		}
	}

	for roleName, permNames := range RolePermissions() {
		var role models.Role
		err := tx.Where("name = ?", string(roleName)).First(&role).Error
		if err == gorm.ErrRecordNotFound {
			role = models.Role{Name: string(roleName), IsSystem: true, Scope: roleScope(roleName)}
			if err := tx.Create(&role).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else if role.IsSystem && role.Scope != roleScope(roleName) {
			if err := tx.Model(&role).Update("scope", roleScope(roleName)).Error; err != nil {
				return err
			}
		}

		var perms []models.Permission
		if err := tx.Where("name IN ?", permNames).Find(&perms).Error; err != nil {
			return err
		}
		if err := tx.Model(&role).Association("Permissions").Replace(perms); err != nil {
			return err
		}
	}
	return nil
}

// roleScope returns the scope system roles are anchored to.
func roleScope(role models.RoleName) string {
	switch role {
	case models.RolePlatformOwner, models.RolePlatformAdmin, models.RolePlatformSupport, models.RolePlatformAuditor, models.RoleSuperAdmin:
		return models.RoleScopePlatform
	default:
		return models.RoleScopeTenant
	}
}

func seedSuperAdmin(db *gorm.DB, cfg *config.Config) error {
	roleName := cfg.SuperAdminRole
	if roleName == "" {
		roleName = string(models.RoleSuperAdmin)
	}
	return db.Transaction(func(tx *gorm.DB) error {
		var user models.User
		email := auth.NormalizeEmail(cfg.SuperAdminEmail)
		err := tx.Where("LOWER(email) = LOWER(?)", email).First(&user).Error
		if err == gorm.ErrRecordNotFound {
			hash, err := auth.HashPassword(cfg.SuperAdminPassword)
			if err != nil {
				return err
			}
			var role models.Role
			if err := tx.Where("name = ?", roleName).First(&role).Error; err != nil {
				return err
			}
			username := userUsernameFromEmail(email)
			u := models.User{
				Username:     username,
				Email:        email,
				PasswordHash: hash,
				FirstName:    cfg.SuperAdminName,
				Status:       models.UserStatusActive,
				RoleID:       role.ID,
			}
			if err := tx.Create(&u).Error; err != nil {
				return err
			}
			slog.Info("super admin seeded", "email", email, "username", username)
			return nil
		}
		if err != nil {
			return err
		}
		return nil
	})
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

func maxIndex(s string) int {
	for i := 0; i < len(s); i++ {
		if s[i] == '@' {
			return i
		}
	}
	return len(s)
}
