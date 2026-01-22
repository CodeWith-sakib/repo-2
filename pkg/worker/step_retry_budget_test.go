package worker

import (
	"testing"
	"time"
)

func TestRetryBudgetPool(t *testing.T) {
	pool := NewRetryBudgetPool(0.1, 2, 100*time.Millisecond)

	// Consume guaranteed floor of 2 retries
	if !pool.RequestRetry() {
		t.Fatal("expected first retry granted")
	}
	if !pool.RequestRetry() {
		t.Fatal("expected second retry granted")
	}
	// 3rd retry rejected without additional initial attempts
	if pool.RequestRetry() {
		t.Error("expected 3rd retry to be rejected due to exhausted budget")
	}

	// Add 30 initial attempts -> 10% ratio grants 3 retries (30 * 0.1 = 3)
	for i := 0; i < 30; i++ {
		pool.RecordInitialAttempt()
	}

	// Now 1 more retry should be granted (total 3 >= 2)
	if !pool.RequestRetry() {
		t.Error("expected retry granted after recording initial attempts")
	}
}
