package core

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrStepRateLimitExceeded = errors.New("step execution rate limit exceeded")
)

// StepRateLimiter applies token-leaky bucket throttling per step type across the cluster.
type StepRateLimiter struct {
	mu         sync.Mutex
	limits     map[string]float64 // stepType -> maxOpsPerSec
	tokens     map[string]float64
	lastRefill map[string]time.Time
}

// NewStepRateLimiter initializes a rate limiter.
func NewStepRateLimiter() *StepRateLimiter {
	return &StepRateLimiter{
		limits:     make(map[string]float64),
		tokens:     make(map[string]float64),
		lastRefill: make(map[string]time.Time),
	}
}

// SetLimit sets max throughput in executions per second for a specific stepType.
func (l *StepRateLimiter) SetLimit(stepType string, opsPerSec float64) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if opsPerSec <= 0 {
		opsPerSec = 100.0
	}
	l.limits[stepType] = opsPerSec
	l.tokens[stepType] = opsPerSec
	l.lastRefill[stepType] = time.Now()
}

// Allow checks if one operation can proceed right now.
func (l *StepRateLimiter) Allow(stepType string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	maxOps, ok := l.limits[stepType]
	if !ok {
		return true // unconstrained
	}

	now := time.Now()
	elapsed := now.Sub(l.lastRefill[stepType]).Seconds()
	l.tokens[stepType] += elapsed * maxOps
	if l.tokens[stepType] > maxOps {
		l.tokens[stepType] = maxOps
	}
	l.lastRefill[stepType] = now

	if l.tokens[stepType] >= 1.0 {
		l.tokens[stepType] -= 1.0
		return true
	}
	return false
}

// Wait blocks until a token becomes available or context cancels.
func (l *StepRateLimiter) Wait(ctx context.Context, stepType string) error {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		if l.Allow(stepType) {
			return nil
		}
		select {
		case <-ctx.Done():
			return ErrStepRateLimitExceeded
		case <-ticker.C:
		}
	}
}
