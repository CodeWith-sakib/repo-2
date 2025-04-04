package postgres

import (
	"context"
	"errors"
	"sync"
	"time"
)

type CircuitState int

const (
	StateClosed CircuitState = iota
	StateHalfOpen
	StateOpen
)

var (
	ErrCircuitOpen = errors.New("database circuit breaker is open")
)

type DBCircuitBreaker struct {
	mu           sync.RWMutex
	state        CircuitState
	failCount    int
	maxFailures  int
	resetTimeout time.Duration
	lastFailTime time.Time
}

func NewDBCircuitBreaker(maxFailures int, resetTimeout time.Duration) *DBCircuitBreaker {
	if maxFailures <= 0 {
		maxFailures = 5
	}
	if resetTimeout <= 0 {
		resetTimeout = 10 * time.Second
	}
	return &DBCircuitBreaker{
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        StateClosed,
	}
}

func (cb *DBCircuitBreaker) Execute(ctx context.Context, op func() error) error {
	cb.mu.Lock()
	now := time.Now()

	if cb.state == StateOpen {
		if now.Sub(cb.lastFailTime) > cb.resetTimeout {
			cb.state = StateHalfOpen
		} else {
			cb.mu.Unlock()
			return ErrCircuitOpen
		}
	}
	cb.mu.Unlock()

	err := op()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failCount++
		cb.lastFailTime = time.Now()
		if cb.failCount >= cb.maxFailures || cb.state == StateHalfOpen {
			cb.state = StateOpen
		}
		return err
	}

	// Success
	if cb.state == StateHalfOpen || cb.failCount > 0 {
		cb.state = StateClosed
		cb.failCount = 0
	}
	return nil
}

func (cb *DBCircuitBreaker) State() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}
