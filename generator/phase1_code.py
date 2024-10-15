# Generator for Phase 1 core packages

CORE_ERRORS_GO = """package core

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound                = errors.New("entity not found")
	ErrAlreadyExists           = errors.New("entity already exists")
	ErrInvalidStateTransition  = errors.New("invalid state transition")
	ErrCycleDetected           = errors.New("dependency cycle detected in workflow DAG")
	ErrValidationFailed        = errors.New("validation failed")
	ErrConflict                = errors.New("concurrent modification conflict")
	ErrLeaseExpired            = errors.New("worker lease has expired")
	ErrWorkerUnavailable       = errors.New("no matching worker available")
	ErrTaskTimeout             = errors.New("task execution timed out")
	ErrWorkflowCancelled       = errors.New("workflow run was cancelled")
	ErrInvalidConfiguration    = errors.New("invalid configuration")
	ErrUnauthorized            = errors.New("unauthorized access")
	ErrForbidden               = errors.New("forbidden operation")
	ErrRateLimited             = errors.New("rate limit exceeded")
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
"""

CORE_TYPES_GO = """package core

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

type ID string

func NewID(prefix string) ID {
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	if prefix == "" {
		return ID(hex.EncodeToString(b))
	}
	return ID(fmt.Sprintf("%s-%s", prefix, hex.EncodeToString(b)))
}

func (id ID) String() string {
	return string(id)
}

func (id ID) IsEmpty() bool {
	return len(id) == 0
}

type Metadata map[string]string

func (m Metadata) Clone() Metadata {
	if m == nil {
		return nil
	}
	c := make(Metadata, len(m))
	for k, v := range m {
		c[k] = v
	}
	return c
}

type JSONPayload []byte

func (p JSONPayload) Clone() JSONPayload {
	if p == nil {
		return nil
	}
	c := make(JSONPayload, len(p))
	copy(c, p)
	return c
}

type Priority int

const (
	PriorityLow      Priority = 10
	PriorityNormal   Priority = 50
	PriorityHigh     Priority = 80
	PriorityCritical Priority = 100
)

type TimeRange struct {
	Start time.Time
	End   time.Time
}

func (tr TimeRange) Contains(t time.Time) bool {
	if !tr.Start.IsZero() && t.Before(tr.Start) {
		return false
	}
	if !tr.End.IsZero() && t.After(tr.End) {
		return false
	}
	return true
}
"""

print("Phase 1 code templates initialized.")
