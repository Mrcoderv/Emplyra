package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"

	"github.com/emplyra/backend/internal/responses"
	"github.com/emplyra/backend/internal/services"
)

var errNoEmployeeLink = errors.New("no employee profile linked to this user")

func parseFormDate(s string) (datatypes.Date, bool) {
	if s == "" {
		return datatypes.Date{}, true
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return datatypes.Date{}, false
	}
	return datatypes.Date(t), true
}

func mapServiceError(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}
	switch {
	case errors.Is(err, services.ErrNotFound):
		responses.Error(c, http.StatusNotFound, err.Error(), nil)
	case errors.Is(err, services.ErrDuplicate):
		responses.Error(c, http.StatusConflict, err.Error(), nil)
	case errors.Is(err, services.ErrInvalidCredentials):
		responses.Error(c, http.StatusUnauthorized, "invalid credentials", nil)
	case errors.Is(err, services.ErrAccountDisabled):
		responses.Error(c, http.StatusForbidden, "account is disabled", nil)
	case errors.Is(err, services.ErrRoleScopeMismatch):
		responses.Error(c, http.StatusForbidden, err.Error(), nil)
	case errors.Is(err, services.ErrTokenInvalid):
		responses.Error(c, http.StatusUnauthorized, "invalid or expired token", nil)
	case errors.Is(err, services.ErrInsufficientLeaveBalance):
		responses.Error(c, http.StatusUnprocessableEntity, err.Error(), nil)
	case errors.Is(err, services.ErrLeaveOverlap):
		responses.Error(c, http.StatusConflict, err.Error(), nil)
	case errors.Is(err, services.ErrAlreadyCheckedIn):
		responses.Error(c, http.StatusConflict, err.Error(), nil)
	case errors.Is(err, services.ErrNoCheckIn):
		responses.Error(c, http.StatusBadRequest, err.Error(), nil)
	case errors.Is(err, services.ErrDuplicateApplication):
		responses.Error(c, http.StatusConflict, err.Error(), nil)
	case errors.Is(err, services.ErrCandidateAlreadyEmployee):
		responses.Error(c, http.StatusConflict, err.Error(), nil)
	case errors.Is(err, services.ErrEnrollmentDuplicate):
		responses.Error(c, http.StatusConflict, err.Error(), nil)
	case errors.Is(err, services.ErrGoogleNotConfigured):
		responses.Error(c, http.StatusServiceUnavailable, err.Error(), nil)
	case errors.Is(err, services.ErrGoogleNotAuthorized):
		responses.Error(c, http.StatusUnauthorized, err.Error(), nil)
	case errors.Is(err, services.ErrGooglePermissionDenied):
		responses.Error(c, http.StatusForbidden, err.Error(), nil)
	case errors.Is(err, services.ErrGoogleRateLimit):
		responses.Error(c, http.StatusTooManyRequests, err.Error(), nil)
	case errors.Is(err, services.ErrGoogleNetwork), errors.Is(err, services.ErrGoogleAPIStatus):
		responses.Error(c, http.StatusBadGateway, err.Error(), nil)
	case errors.Is(err, services.ErrGoogleInvalidForm), errors.Is(err, services.ErrGoogleNoData):
		responses.Error(c, http.StatusUnprocessableEntity, err.Error(), nil)
	case errors.Is(err, services.ErrGoogleInvalidSpreadsheet), errors.Is(err, services.ErrGoogleNotConnected):
		responses.Error(c, http.StatusBadRequest, err.Error(), nil)
	case errors.Is(err, services.ErrGoogleMissingHeader), errors.Is(err, services.ErrGoogleMissingEmail),
		errors.Is(err, services.ErrGoogleTargetInvalid), errors.Is(err, services.ErrGoogleOAuthStateInvalid):
		responses.Error(c, http.StatusUnprocessableEntity, err.Error(), nil)
	default:
		slog.Error("unhandled service error", "err", err, "path", c.Request.URL.Path)
		responses.Error(c, http.StatusInternalServerError, "internal server error", nil)
	}
	return true
}

func clientIP(c *gin.Context) string {
	ip := c.ClientIP()
	if ip == "" {
		ip = "unknown"
	}
	return ip
}

func userAgent(c *gin.Context) string {
	ua := c.GetHeader("User-Agent")
	if len(ua) > 255 {
		ua = ua[:255]
	}
	return ua
}

func sanitizeField(s string) string {
	return strings.TrimSpace(s)
}
