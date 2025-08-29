package worker

import (
	"fmt"
	"sync"
	"time"
)

// RetryBudgetPolicy defines the retry budget parameters for a task class.
type RetryBudgetPolicy struct {
	MaxAttempts      int
	BudgetWindowSize int           // sliding window of recent executions
	MaxRetryRatio    float64       // max fraction of executions that are retries (0.0–1.0)
	CooldownDuration time.Duration // mandatory pause after budget exhaustion
}

// windowEntry records whether an execution was a retry.
type windowEntry struct {
	isRetry bool
	at      time.Time
}

// RetryBudgetTracker implements token-bucket–style retry budget management:
// tracks a sliding window of recent executions and rejects excess retries.
type RetryBudgetTracker struct {
	mu                sync.Mutex
	policy            RetryBudgetPolicy
	window            []windowEntry
	budgetExhaustedAt time.Time
}

// NewRetryBudgetTracker creates a tracker from the given policy.
func NewRetryBudgetTracker(policy RetryBudgetPolicy) *RetryBudgetTracker {
	return &RetryBudgetTracker{
		policy: policy,
		window: make([]windowEntry, 0, policy.BudgetWindowSize),
	}
}

// CanRetry returns nil if a retry is permitted, or an error describing the budget refusal.
func (r *RetryBudgetTracker) CanRetry(attempt int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if attempt > r.policy.MaxAttempts {
		return fmt.Errorf("attempt %d exceeds max attempts %d", attempt, r.policy.MaxAttempts)
	}

	// Check if still in cooldown
	if !r.budgetExhaustedAt.IsZero() {
		if time.Since(r.budgetExhaustedAt) < r.policy.CooldownDuration {
			remaining := r.policy.CooldownDuration - time.Since(r.budgetExhaustedAt)
			return fmt.Errorf("retry budget exhausted, cooldown remaining: %s", remaining.Round(time.Millisecond))
		}
		r.budgetExhaustedAt = time.Time{}
	}

	// Trim window
	r.trim()

	// Count retries in window
	retries := 0
	for _, e := range r.window {
		if e.isRetry {
			retries++
		}
	}

	total := len(r.window) + 1 // +1 for this upcoming retry
	ratio := float64(retries+1) / float64(total)

	if ratio > r.policy.MaxRetryRatio {
		r.budgetExhaustedAt = time.Now()
		return fmt.Errorf("retry ratio %.2f exceeds budget limit %.2f", ratio, r.policy.MaxRetryRatio)
	}

	return nil
}

// RecordExecution adds an execution entry to the sliding window.
func (r *RetryBudgetTracker) RecordExecution(isRetry bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.trim()
	r.window = append(r.window, windowEntry{isRetry: isRetry, at: time.Now()})
}

func (r *RetryBudgetTracker) trim() {
	if len(r.window) >= r.policy.BudgetWindowSize {
		r.window = r.window[len(r.window)-r.policy.BudgetWindowSize+1:]
	}
}

// RetryStats returns current window stats for observability.
func (r *RetryBudgetTracker) RetryStats() (total, retries int, ratio float64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	total = len(r.window)
	for _, e := range r.window {
		if e.isRetry {
			retries++
		}
	}
	if total > 0 {
		ratio = float64(retries) / float64(total)
	}
	return
}
