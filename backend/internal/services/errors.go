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
