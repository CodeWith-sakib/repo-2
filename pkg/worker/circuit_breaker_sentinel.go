package worker

import (
	"errors"
	"sync"
	"time"
)

// CircuitBreakerState defines sentinel operational states.
type CircuitBreakerState string

const (
	CircuitStateClosed   CircuitBreakerState = "CLOSED"
	CircuitStateOpen     CircuitBreakerState = "OPEN"
	CircuitStateHalfOpen CircuitBreakerState = "HALF_OPEN"
)

// CircuitBreakerSentinel prevents repeated calls to failing downstream worker dependencies.
type CircuitBreakerSentinel struct {
	mu            sync.Mutex
	state         CircuitBreakerState
	failureCount  int
	failureThreshold int
	resetTimeout  time.Duration
	openedAt      time.Time
}

// NewCircuitBreakerSentinel creates a circuit breaker sentinel.
func NewCircuitBreakerSentinel(failureThreshold int, resetTimeout time.Duration) *CircuitBreakerSentinel {
	if failureThreshold <= 0 {
		failureThreshold = 5
	}
	if resetTimeout <= 0 {
		resetTimeout = 30 * time.Second
	}
	return &CircuitBreakerSentinel{
		state:            CircuitStateClosed,
		failureThreshold: failureThreshold,
		resetTimeout:     resetTimeout,
	}
}

// Allow checks whether execution may proceed under current circuit state.
func (cb *CircuitBreakerSentinel) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now().UTC()
	if cb.state == CircuitStateOpen {
		if now.Sub(cb.openedAt) >= cb.resetTimeout {
			cb.state = CircuitStateHalfOpen
			return true
		}
		return false
	}

	return true
}

// RecordResult records success or failure to transition circuit states.
func (cb *CircuitBreakerSentinel) RecordResult(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err == nil {
		cb.failureCount = 0
		cb.state = CircuitStateClosed
		return
	}

	cb.failureCount++
	if cb.failureCount >= cb.failureThreshold || cb.state == CircuitStateHalfOpen {
		cb.state = CircuitStateOpen
		cb.openedAt = time.Now().UTC()
	}
}

// State returns current circuit state string.
func (cb *CircuitBreakerSentinel) State() CircuitBreakerState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

// ErrCircuitOpen is returned when breaker rejects execution.
var ErrCircuitOpen = errors.New("circuit breaker is open; downstream service unavailable")
