package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/emplyra/backend/internal/repositories"
)

const EmployeeIDKey = "employee_id"

// EmployeeScope resolves the employee record linked to the authenticated user
// and stores its ID in the context. Used by employee-scoped modules.
func EmployeeScope(empRepo *repositories.EmployeeRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		p := GetPrincipal(c)
		if p != nil {
			if emp, err := empRepo.FindByUserID(p.UserID); err == nil {
				c.Set(EmployeeIDKey, emp.ID)
			}
		}
		c.Next()
	}
}

func EmployeeID(c *gin.Context) string {
	if v, ok := c.Get(EmployeeIDKey); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
