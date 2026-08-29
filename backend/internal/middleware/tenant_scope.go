package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/emplyra/backend/internal/models"
	"github.com/emplyra/backend/internal/utils"
)

const (
	TenantIDKey    = "tenant_id"
	SupportModeKey = "support_mode"
	TenantHeader   = "X-Tenant-ID"
)

// TenantResolver loads tenants for scope validation.
type TenantResolver interface {
	FindByID(id string) (*models.Tenant, error)
}

// TenantScope resolves the effective tenant of a request and stores it in the
// context. Rules:
//
//   - Tenant users always operate inside the tenant bound to their account
//     (from the JWT); suspended/inactive tenants are blocked (403).
//   - Platform users operate platform-wide with no tenant, UNLESS they present
//     the X-Tenant-ID header (controlled support access) and hold the
//     platform:tenant-access permission; that enables "Open Tenant" support mode.
//
// A tenant id is never accepted from a regular tenant user.
func TenantScope(tenants TenantResolver, resolver PermissionResolver, required string) gin.HandlerFunc {
	return func(c *gin.Context) {
		p := GetPrincipal(c)
		if p == nil {
			utils.Abort(c, http.StatusUnauthorized, "unauthenticated", nil)
			return
		}

		if p.Scope == models.RoleScopePlatform {
			requested := c.GetHeader(TenantHeader)
			if requested != "" {
				if !hasPermission(resolver, p.RoleID, required) {
					utils.Abort(c, http.StatusForbidden, "forbidden: opening a tenant requires "+required, nil)
					return
				}
				tenant, err := tenants.FindByID(requested)
				if err != nil {
					utils.Abort(c, http.StatusBadRequest, "unknown tenant", nil)
					return
				}
				setTenant(c, tenant.ID)
				c.Set(SupportModeKey, true)
				c.Next()
				return
			}
			c.Next()
			return
		}

		tenant, err := tenants.FindByID(p.TenantID)
		if err != nil {
			utils.Abort(c, http.StatusForbidden, "tenant not found", nil)
			return
		}
		if !tenant.IsOperational() {
			utils.Abort(c, http.StatusForbidden, "tenant "+string(tenant.Status), nil)
			return
		}
		setTenant(c, tenant.ID)
		c.Next()
	}
}

func hasPermission(resolver PermissionResolver, roleID, perm string) bool {
	for _, p := range resolver.PermissionsForRole(roleID) {
		if p == perm {
			return true
		}
	}
	return false
}

func setTenant(c *gin.Context, tenantID string) {
	c.Set(TenantIDKey, tenantID)
}

func TenantID(c *gin.Context) string {
	if v, ok := c.Get(TenantIDKey); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func IsSupportMode(c *gin.Context) bool {
	v, _ := c.Get(SupportModeKey)
	b, _ := v.(bool)
	return b
}
