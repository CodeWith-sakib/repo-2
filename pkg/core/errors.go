package core

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound               = errors.New("entity not found")
	ErrAlreadyExists          = errors.New("entity already exists")
	ErrInvalidStateTransition = errors.New("invalid state transition")
	ErrCycleDetected          = errors.New("dependency cycle detected in workflow DAG")
	ErrValidationFailed       = errors.New("validation failed")
	ErrConflict               = errors.New("concurrent modification conflict")
	ErrLeaseExpired           = errors.New("worker lease has expired")
	ErrWorkerUnavailable      = errors.New("no matching worker available")
	ErrTaskTimeout            = errors.New("task execution timed out")
	ErrWorkflowCancelled      = errors.New("workflow run was cancelled")
	ErrInvalidConfiguration   = errors.New("invalid configuration")
	ErrUnauthorized           = errors.New("unauthorized access")
	ErrForbidden              = errors.New("forbidden operation")
	ErrRateLimited            = errors.New("rate limit exceeded")
)

type DomainError struct {
	Code    string
	Message string
	Err     error
}

func (e *DomainError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *DomainError) Unwrap() error {
	return e.Err
}

func NewDomainError(code, message string, underlying error) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
		Err:     underlying,
	}
}
