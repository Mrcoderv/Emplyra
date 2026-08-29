package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type stubResolver struct {
	perms map[string][]string
}

func (s stubResolver) PermissionsForRole(roleID string) []string {
	return s.perms[roleID]
}

func newTestEngine(resolver PermissionResolver) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(PrincipalKey, &Principal{UserID: "u1", Role: "HR_ADMIN", RoleID: "role-a"})
		c.Next()
	})
	r.GET("/secure", RBAC(resolver, "user:read"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	return r
}

func TestRBACAllows(t *testing.T) {
	resolver := stubResolver{perms: map[string][]string{"role-a": {"user:read", "user:create"}}}
	r := newTestEngine(resolver)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/secure", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestRBACDenies(t *testing.T) {
	resolver := stubResolver{perms: map[string][]string{"role-a": {"user:create"}}}
	r := newTestEngine(resolver)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/secure", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestRBACUnauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New() // no principal set
	r.GET("/secure", RBAC(stubResolver{}, "user:read"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/secure", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}
