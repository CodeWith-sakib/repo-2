package worker

import (
	"testing"
	"time"
)

func TestAdaptiveConcurrencyLimiter(t *testing.T) {
	limiter := NewAdaptiveConcurrencyLimiter(5.0, 2.0, 20.0)

	if limiter.Limit() != 5 {
		t.Fatalf("expected initial limit 5, got %d", limiter.Limit())
	}

	// Acquire 5 slots
	for i := 0; i < 5; i++ {
		if !limiter.TryAcquire() {
			t.Fatalf("expected slot %d to be acquired", i)
		}
	}
	// 6th rejected
	if limiter.TryAcquire() {
		t.Error("expected 6th acquire to be rejected")
	}

	// Release with low RTT -> expands limit
	limiter.Release(10 * time.Millisecond)
	limiter.Release(10 * time.Millisecond)

	// After low RTT, limit should grow or stay at least initial
	if limiter.Limit() < 5 {
		t.Errorf("expected limit >= 5, got %d", limiter.Limit())
	}

	// Simulate high latency / overload -> throttles limit
	for i := 0; i < 10; i++ {
		_ = limiter.TryAcquire()
		limiter.Release(200 * time.Millisecond)
	}

	if limiter.Limit() > 10 {
		t.Errorf("expected limit to throttle under high RTT, got %d", limiter.Limit())
	}
}
