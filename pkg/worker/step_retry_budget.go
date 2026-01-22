package worker

import (
	"sync"
	"time"
)

// RetryBudgetPool prevents cascading worker retry storms using token consumption limits.
type RetryBudgetPool struct {
	mu           sync.Mutex
	ratio        float64       // allowed retry percentage of initial attempts (e.g. 0.1 for 10%)
	minRetries   int           // minimum guaranteed retry allowance
	initialTries int           // total initial attempts in current window
	retriesUsed  int           // total retries granted in current window
	windowStart  time.Time
	windowTTL    time.Duration
}

// NewRetryBudgetPool constructs a retry budget manager with ratio and minimum floor.
func NewRetryBudgetPool(ratio float64, minRetries int, windowTTL time.Duration) *RetryBudgetPool {
	if ratio <= 0 {
		ratio = 0.1
	}
	if minRetries <= 0 {
		minRetries = 10
	}
	if windowTTL <= 0 {
		windowTTL = 1 * time.Minute
	}
	return &RetryBudgetPool{
		ratio:       ratio,
		minRetries:  minRetries,
		windowTTL:   windowTTL,
		windowStart: time.Now().UTC(),
	}
}

// RecordInitialAttempt registers a new first-time task invocation.
func (p *RetryBudgetPool) RecordInitialAttempt() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.rotateWindowIfNeeded(time.Now().UTC())
	p.initialTries++
}

// RequestRetry attempts to draw retry permission from the budget.
func (p *RetryBudgetPool) RequestRetry() bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now().UTC()
	p.rotateWindowIfNeeded(now)

	allowedRetries := int(float64(p.initialTries) * p.ratio)
	if allowedRetries < p.minRetries {
		allowedRetries = p.minRetries
	}

	if p.retriesUsed < allowedRetries {
		p.retriesUsed++
		return true
	}

	return false
}

func (p *RetryBudgetPool) rotateWindowIfNeeded(now time.Time) {
	if now.Sub(p.windowStart) >= p.windowTTL {
		p.windowStart = now
		p.initialTries = 0
		p.retriesUsed = 0
	}
}

// Stats returns current window metrics.
func (p *RetryBudgetPool) Stats() (initial, retries int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.initialTries, p.retriesUsed
}
