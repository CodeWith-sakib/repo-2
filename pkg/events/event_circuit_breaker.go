package events

import (
	"sync"
	"time"
)

// EventPublishCircuitBreaker wraps event publisher dispatch with trip protection on downstream outages.
type EventPublishCircuitBreaker struct {
	mu           sync.Mutex
	failures     int
	threshold    int
	lastFailure  time.Time
	cooldown     time.Duration
	isOpen       bool
}

// NewEventPublishCircuitBreaker creates an event circuit breaker.
func NewEventPublishCircuitBreaker(threshold int, cooldown time.Duration) *EventPublishCircuitBreaker {
	if threshold <= 0 {
		threshold = 5
	}
	if cooldown <= 0 {
		cooldown = 15 * time.Second
	}
	return &EventPublishCircuitBreaker{
		threshold: threshold,
		cooldown:  cooldown,
	}
}

// CanPublish checks whether message publishing is allowed.
func (cb *EventPublishCircuitBreaker) CanPublish(now time.Time) bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.isOpen {
		if now.Sub(cb.lastFailure) >= cb.cooldown {
			// Half-open trial
			return true
		}
		return false
	}

	return true
}

// RecordSuccess resets failure counts on successful message delivery.
func (cb *EventPublishCircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures = 0
	cb.isOpen = false
}

// RecordFailure increments failures and trips breaker if threshold exceeded.
func (cb *EventPublishCircuitBreaker) RecordFailure(now time.Time) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.lastFailure = now
	if cb.failures >= cb.threshold {
		cb.isOpen = true
	}
}

// IsOpen checks if breaker is currently open.
func (cb *EventPublishCircuitBreaker) IsOpen() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.isOpen
}
