package worker

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrCircuitBreakerTripped = errors.New("circuit breaker is open; rejecting execution")
)

type StepCircuitStatus string

const (
	StepCircuitClosed   StepCircuitStatus = "CLOSED"
	StepCircuitHalfOpen StepCircuitStatus = "HALF_OPEN"
	StepCircuitOpen     StepCircuitStatus = "OPEN"
)

// StepFailureCircuitBreaker shields external systems from repeated cascading task failures.
type StepFailureCircuitBreaker struct {
	mu               sync.RWMutex
	consecutiveFails int
	failThreshold    int
	cooldownWindow   time.Duration
	openedAt         time.Time
	status           StepCircuitStatus
}

// NewStepFailureCircuitBreaker initializes a circuit breaker.
func NewStepFailureCircuitBreaker(threshold int, cooldown time.Duration) *StepFailureCircuitBreaker {
	if threshold <= 0 {
		threshold = 5
	}
	if cooldown <= 0 {
		cooldown = 30 * time.Second
	}
	return &StepFailureCircuitBreaker{
		failThreshold:  threshold,
		cooldownWindow: cooldown,
		status:         StepCircuitClosed,
	}
}

// AllowExecution verifies whether a request is permitted to proceed.
func (cb *StepFailureCircuitBreaker) AllowExecution() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()
	if cb.status == StepCircuitOpen {
		if now.Sub(cb.openedAt) > cb.cooldownWindow {
			cb.status = StepCircuitHalfOpen
			return true
		}
		return false
	}
	return true
}

// RecordSuccess records successful run, resetting breaker to CLOSED.
func (cb *StepFailureCircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.consecutiveFails = 0
	cb.status = StepCircuitClosed
}

// RecordFailure increments fail counts and trips circuit to OPEN if threshold exceeded.
func (cb *StepFailureCircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.consecutiveFails++
	if cb.consecutiveFails >= cb.failThreshold {
		cb.status = StepCircuitOpen
		cb.openedAt = time.Now()
	}
}

// Status returns current circuit status.
func (cb *StepFailureCircuitBreaker) Status() StepCircuitStatus {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.status
}
