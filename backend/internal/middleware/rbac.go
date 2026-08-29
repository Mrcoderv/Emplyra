package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/emplyra/backend/internal/repositories"
	"github.com/emplyra/backend/internal/utils"
)

// PermissionResolver loads the permission set granted to a role.
type PermissionResolver interface {
	PermissionsForRole(roleID string) []string
}

// cachedRolePermissions resolves role permissions with a short TTL cache so
// RBAC checks don't hammer the database on every request.
type cachedRolePermissions struct {
	mu    sync.Mutex
	ttl   time.Duration
	users *repositories.UserRepository
	cache map[string]cacheEntry
}

type cacheEntry struct {
	perms []string
	at    time.Time
}

func NewPermissionResolver(users *repositories.UserRepository, ttl time.Duration) *cachedRolePermissions {
	return &cachedRolePermissions{
		ttl:   ttl,
		users: users,
		cache: map[string]cacheEntry{},
	}
}

func (c *cachedRolePermissions) PermissionsForRole(roleID string) []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	if e, ok := c.cache[roleID]; ok && time.Since(e.at) < c.ttl {
		return e.perms
	}
	perms, err := c.users.RolePermissions(roleID)
	if err != nil {
		return nil
	}
	c.cache[roleID] = cacheEntry{perms: perms, at: time.Now()}
	return perms
}

// Invalidate drops cached permissions for a role (call after role/permission changes).
func (c *cachedRolePermissions) Invalidate(roleID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.cache, roleID)
}

func (c *cachedRolePermissions) InvalidateAll() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache = map[string]cacheEntry{}
}

// RBAC returns a middleware denying requests whose principal lacks `permission`.
func RBAC(resolver PermissionResolver, permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		p := GetPrincipal(c)
		if p == nil {
			utils.Abort(c, http.StatusUnauthorized, "unauthenticated", nil)
			return
		}
		for _, perm := range resolver.PermissionsForRole(p.RoleID) {
			if perm == permission {
				c.Next()
				return
			}
		}
		utils.Abort(c, http.StatusForbidden, "forbidden: missing permission "+permission, nil)
	}
}
