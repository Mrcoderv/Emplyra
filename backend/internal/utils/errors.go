package utils

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var ErrNotFound = errors.New("not found")

func ConvertValidationErrors(err error) map[string]string {
	out := map[string]string{}
	var verr validator.ValidationErrors
	if errors.As(err, &verr) {
		for _, fe := range verr {
			out[fe.Field()] = "failed on the '" + fe.Tag() + "' condition"
		}
	}
	if len(out) == 0 {
		out["_"] = err.Error()
	}
	return out
}

func Abort(c *gin.Context, status int, message string, errs interface{}) {
	c.AbortWithStatusJSON(status, map[string]interface{}{
		"success": false,
		"message": message,
		"errors":  errs,
	})
}

func NotFound(c *gin.Context, resource string) {
	Abort(c, http.StatusNotFound, "not found", map[string]string{"resource": resource})
}
