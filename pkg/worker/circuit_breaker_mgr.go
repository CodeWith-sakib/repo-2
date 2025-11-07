package worker

import (
	"fmt"
	"sync"
	"time"
)

// CircuitState defines the operational state of a circuit breaker.
type CircuitState string

const (
	CircuitClosed   CircuitState = "CLOSED"
	CircuitOpen     CircuitState = "OPEN"
	CircuitHalfOpen CircuitState = "HALF_OPEN"
)

// BreakerConfig sets thresholds for opening and recovering the circuit breaker.
type BreakerConfig struct {
	FailureThreshold   int           // consecutive failures to trip open
	CooldownPeriod     time.Duration // time in OPEN before testing HALF_OPEN
	HalfOpenSuccessReq int           // consecutive successes in HALF_OPEN to close
}

// DefaultBreakerConfig returns standard default configuration.
func DefaultBreakerConfig() BreakerConfig {
	return BreakerConfig{
		FailureThreshold:   5,
		CooldownPeriod:     10 * time.Second,
		HalfOpenSuccessReq: 3,
	}
}

// ServiceBreaker tracks failure and recovery stats for a single service.
type ServiceBreaker struct {
	Name            string
	State           CircuitState
	Failures        int
	SuccessesInHalf int
	LastTripAt      time.Time
	TotalTrips      int64
}

// CircuitBreakerManager coordinates circuit breakers across all external services.
type CircuitBreakerManager struct {
	mu       sync.RWMutex
	breakers map[string]*ServiceBreaker
	cfg      BreakerConfig
}

// NewCircuitBreakerManager creates a circuit breaker manager.
func NewCircuitBreakerManager(cfg BreakerConfig) *CircuitBreakerManager {
	if cfg.FailureThreshold <= 0 {
		cfg.FailureThreshold = 5
	}
	if cfg.CooldownPeriod <= 0 {
		cfg.CooldownPeriod = 10 * time.Second
	}
	if cfg.HalfOpenSuccessReq <= 0 {
		cfg.HalfOpenSuccessReq = 2
	}
	return &CircuitBreakerManager{
		breakers: make(map[string]*ServiceBreaker),
		cfg:      cfg,
	}
}

// Allow checks whether a call to service is permitted. Returns error if circuit is open.
func (m *CircuitBreakerManager) Allow(service string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	b := m.getOrCreate(service)
	now := time.Now()

	switch b.State {
	case CircuitClosed:
		return nil

	case CircuitOpen:
		if now.Sub(b.LastTripAt) >= m.cfg.CooldownPeriod {
			// Transition to HALF_OPEN to probe recovery
			b.State = CircuitHalfOpen
			b.SuccessesInHalf = 0
			return nil
		}
		return fmt.Errorf("circuit breaker for service %q is OPEN (tripped %v ago)", service, now.Sub(b.LastTripAt).Round(time.Millisecond))

	case CircuitHalfOpen:
		// Permit probe requests
		return nil

	default:
		return nil
	}
}

// RecordSuccess records a successful call to the service.
func (m *CircuitBreakerManager) RecordSuccess(service string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	b := m.getOrCreate(service)
	b.Failures = 0

	if b.State == CircuitHalfOpen {
		b.SuccessesInHalf++
		if b.SuccessesInHalf >= m.cfg.HalfOpenSuccessReq {
			// Fully recovered! Return to CLOSED
			b.State = CircuitClosed
			b.SuccessesInHalf = 0
		}
	}
}

// RecordFailure records a failure for the service, potentially tripping to OPEN.
func (m *CircuitBreakerManager) RecordFailure(service string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	b := m.getOrCreate(service)
	b.Failures++

	if b.State == CircuitHalfOpen {
		// Probe failed -> trip immediately back to OPEN
		b.State = CircuitOpen
		b.LastTripAt = time.Now()
		b.TotalTrips++
		return
	}

	if b.State == CircuitClosed && b.Failures >= m.cfg.FailureThreshold {
		b.State = CircuitOpen
		b.LastTripAt = time.Now()
		b.TotalTrips++
	}
}

// State returns current state of a service's breaker.
func (m *CircuitBreakerManager) State(service string) CircuitState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	b, ok := m.breakers[service]
	if !ok {
		return CircuitClosed
	}
	return b.State
}

func (m *CircuitBreakerManager) getOrCreate(service string) *ServiceBreaker {
	b, ok := m.breakers[service]
	if !ok {
		b = &ServiceBreaker{
			Name:  service,
			State: CircuitClosed,
		}
		m.breakers[service] = b
	}
	return b
}
