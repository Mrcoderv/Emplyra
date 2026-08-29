package services

import "errors"

var (
	ErrNotFound                 = errors.New("resource not found")
	ErrDuplicate                = errors.New("record already exists")
	ErrInsufficientLeaveBalance = errors.New("insufficient leave balance")
	ErrLeaveOverlap             = errors.New("leave dates overlap with an existing application")
	ErrAlreadyCheckedIn         = errors.New("already checked in for this date")
	ErrNoCheckIn                = errors.New("check-out requires a check-in first")
	ErrDuplicateApplication     = errors.New("candidate has already applied to this job")
	ErrCandidateAlreadyEmployee = errors.New("candidate already exists as an employee")
	ErrEnrollmentDuplicate      = errors.New("employee is already enrolled in this program")

	// Google Forms integration.
	ErrGoogleNotConfigured      = errors.New("google oauth is not configured")
	ErrGoogleNotAuthorized      = errors.New("google account is not authorized")
	ErrGoogleInvalidSpreadsheet = errors.New("invalid google spreadsheet")
	ErrGooglePermissionDenied   = errors.New("google permission denied")
	ErrGoogleRateLimit          = errors.New("google api rate limit exceeded")
	ErrGoogleNetwork            = errors.New("google api network failure")
	ErrGoogleAPIStatus          = errors.New("google api returned an error")
	ErrGoogleInvalidForm        = errors.New("invalid google form url")
	ErrGoogleMissingHeader      = errors.New("google form sheet is missing a required column")
	ErrGoogleMissingEmail       = errors.New("google form sheet has no email column mapped")
	ErrGoogleNotConnected       = errors.New("job is not connected to a google form")
	ErrGoogleNoData             = errors.New("google form sheet has no responses")
	ErrGoogleTargetInvalid      = errors.New("unsupported google form hrms field target")
	ErrGoogleOAuthStateInvalid  = errors.New("invalid or expired google oauth state")
)

func MapError(err error) error {
	if err == ErrDuplicate {
		return errors.New("record already exists")
	}
	if err == ErrNotFound {
		return errors.New("resource not found")
	}
	return err
}
