package retry

import (
	"errors"
	"sync"
	"time"
)

type CircuitState string

const (
	StateClosed   CircuitState = "CLOSED"
	StateHalfOpen CircuitState = "HALF_OPEN"
	StateOpen     CircuitState = "OPEN"
)

var ErrCircuitOpen = errors.New("circuit breaker is open")

type CircuitBreakerConfig struct {
	MaxFailures int
	Timeout     time.Duration
}

type CircuitBreaker struct {
	mu           sync.RWMutex
	cfg          CircuitBreakerConfig
	state        CircuitState
	failures     int
	lastFailure  time.Time
	successCount int
}

func NewCircuitBreaker(cfg CircuitBreakerConfig) *CircuitBreaker {
	if cfg.MaxFailures <= 0 {
		cfg.MaxFailures = 5
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 10 * time.Second
	}
	return &CircuitBreaker{
		cfg:   cfg,
		state: StateClosed,
	}
}

func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

func (cb *CircuitBreaker) Execute(op func() error) error {
	cb.mu.Lock()
	now := time.Now()

	if cb.state == StateOpen {
		if now.Sub(cb.lastFailure) > cb.cfg.Timeout {
			cb.state = StateHalfOpen
			cb.successCount = 0
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
		cb.failures++
		cb.lastFailure = time.Now()
		cb.state = StateOpen
		return err
	}

	if cb.state == StateHalfOpen {
		cb.successCount++
		if cb.successCount >= 2 {
			cb.state = StateClosed
			cb.failures = 0
		}
	} else {
		cb.failures = 0
	}

	return nil
}
