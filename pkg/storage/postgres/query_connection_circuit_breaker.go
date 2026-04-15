package postgres

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"time"
)

// DBConnectionCircuitBreaker prevents database connection saturation by rejecting queries when connection acquisition times out.
type DBConnectionCircuitBreaker struct {
	mu           sync.Mutex
	failures     int
	threshold    int
	cooldown     time.Duration
	lastFailure  time.Time
	isOpen       bool
}

// NewDBConnectionCircuitBreaker creates a database connection circuit breaker.
func NewDBConnectionCircuitBreaker(threshold int, cooldown time.Duration) *DBConnectionCircuitBreaker {
	if threshold <= 0 {
		threshold = 3
	}
	if cooldown <= 0 {
		cooldown = 10 * time.Second
	}
	return &DBConnectionCircuitBreaker{
		threshold: threshold,
		cooldown:  cooldown,
	}
}

// PingOrCheck checks if database is reachable, tripping breaker if connection fails.
func (cb *DBConnectionCircuitBreaker) PingOrCheck(ctx context.Context, db *sql.DB) error {
	cb.mu.Lock()
	now := time.Now().UTC()
	if cb.isOpen {
		if now.Sub(cb.lastFailure) < cb.cooldown {
			cb.mu.Unlock()
			return errors.New("database connection circuit breaker is open; skipping connection attempt")
		}
	}
	cb.mu.Unlock()

	if db == nil {
		cb.RecordSuccess()
		return nil
	}

	err := db.PingContext(ctx)
	if err != nil {
		cb.RecordFailure(now)
		return err
	}

	cb.RecordSuccess()
	return nil
}

// RecordSuccess marks connection healthy and resets failure counter.
func (cb *DBConnectionCircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures = 0
	cb.isOpen = false
}

// RecordFailure marks a failed connection attempt.
func (cb *DBConnectionCircuitBreaker) RecordFailure(now time.Time) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures++
	cb.lastFailure = now
	if cb.failures >= cb.threshold {
		cb.isOpen = true
	}
}

// IsOpen checks if breaker is currently open.
func (cb *DBConnectionCircuitBreaker) IsOpen() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.isOpen
}
