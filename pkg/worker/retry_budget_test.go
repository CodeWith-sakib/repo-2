package worker

import (
	"testing"
	"time"
)

func TestRetryBudgetTracker_Limits(t *testing.T) {
	policy := RetryBudgetPolicy{
		MaxAttempts:      5,
		BudgetWindowSize: 10,
		MaxRetryRatio:    0.3,
		CooldownDuration: 50 * time.Millisecond,
	}

	tracker := NewRetryBudgetTracker(policy)

	// Fill window with non-retries to establish headroom
	for i := 0; i < 7; i++ {
		tracker.RecordExecution(false)
	}

	// First retry attempt should be allowed
	if err := tracker.CanRetry(2); err != nil {
		t.Errorf("first retry should be allowed: %v", err)
	}
	tracker.RecordExecution(true)

	// Max attempts exceeded
	if err := tracker.CanRetry(6); err == nil {
		t.Error("expected error for attempt 6 > maxAttempts 5")
	}
}

func TestRetryBudgetTracker_BudgetExhausted(t *testing.T) {
	policy := RetryBudgetPolicy{
		MaxAttempts:      10,
		BudgetWindowSize: 4,
		MaxRetryRatio:    0.25, // at most 1-in-4 may be retries
		CooldownDuration: 50 * time.Millisecond,
	}

	tracker := NewRetryBudgetTracker(policy)

	// Fill window with all retries to exceed budget
	for i := 0; i < 4; i++ {
		tracker.RecordExecution(true)
	}

	// Budget should now be exhausted (all 4 entries are retries, ratio = 1.0 > 0.25)
	if err := tracker.CanRetry(1); err == nil {
		t.Error("expected budget exceeded error when all window entries are retries")
	}

	// After cooldown, a fresh retry should be gated normally
	time.Sleep(60 * time.Millisecond)
	// Now add some non-retries to bring ratio down
	tracker.RecordExecution(false)
	tracker.RecordExecution(false)
	tracker.RecordExecution(false)

	_, retries, ratio := tracker.RetryStats()
	if ratio > 0.5 {
		t.Errorf("expected ratio to drop after non-retry executions, got retries=%d ratio=%.2f", retries, ratio)
	}
}
