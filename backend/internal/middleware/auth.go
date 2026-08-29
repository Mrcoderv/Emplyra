package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/emplyra/backend/internal/auth"
	"github.com/emplyra/backend/internal/utils"
)

const PrincipalKey = "principal"

type Principal struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	RoleID   string `json:"role_id"`
}

func Auth(jwt *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		hdr := c.GetHeader("Authorization")
		if hdr == "" || !strings.HasPrefix(hdr, "Bearer ") {
			utils.Abort(c, http.StatusUnauthorized, "missing or invalid authorization header", nil)
			return
		}
		token := strings.TrimPrefix(hdr, "Bearer ")
		claims, err := jwt.Parse(token)
		if err != nil {
			utils.Abort(c, http.StatusUnauthorized, "invalid or expired token", nil)
			return
		}
		c.Set(PrincipalKey, &Principal{
			UserID:   claims.UserID,
			Username: claims.Username,
			Role:     claims.Role,
			RoleID:   claims.RoleID,
		})
		c.Next()
	}
}

func GetPrincipal(c *gin.Context) *Principal {
	v, ok := c.Get(PrincipalKey)
	if !ok {
		return nil
	}
	p, ok := v.(*Principal)
	if !ok {
		return nil
	}
	return p
}

func MustPrincipal(c *gin.Context) *Principal {
	p := GetPrincipal(c)
	if p == nil {
		utils.Abort(c, http.StatusUnauthorized, "unauthenticated", nil)
		c.Abort()
	}
	return p
}
