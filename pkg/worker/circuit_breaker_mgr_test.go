package worker

import (
	"testing"
	"time"
)

func TestCircuitBreakerManager_TripAndRecover(t *testing.T) {
	cfg := BreakerConfig{
		FailureThreshold:   2,
		CooldownPeriod:     20 * time.Millisecond,
		HalfOpenSuccessReq: 2,
	}
	mgr := NewCircuitBreakerManager(cfg)
	service := "payment-gateway"

	// Initially closed, allowed
	if err := mgr.Allow(service); err != nil {
		t.Fatalf("expected initial allow, got %v", err)
	}

	// 1 failure -> still closed
	mgr.RecordFailure(service)
	if mgr.State(service) != CircuitClosed {
		t.Errorf("expected closed after 1 failure, got %s", mgr.State(service))
	}

	// 2nd failure -> trips OPEN!
	mgr.RecordFailure(service)
	if mgr.State(service) != CircuitOpen {
		t.Errorf("expected open after 2 failures, got %s", mgr.State(service))
	}

	// Calls should now be rejected
	if err := mgr.Allow(service); err == nil {
		t.Error("expected error when breaker is OPEN")
	}

	// Wait for cooldown -> transitions to HALF_OPEN
	time.Sleep(30 * time.Millisecond)
	if err := mgr.Allow(service); err != nil {
		t.Errorf("expected probe allowed in HALF_OPEN, got: %v", err)
	}
	if mgr.State(service) != CircuitHalfOpen {
		t.Errorf("expected state HALF_OPEN, got %s", mgr.State(service))
	}

	// First success in half-open
	mgr.RecordSuccess(service)
	if mgr.State(service) != CircuitHalfOpen {
		t.Errorf("expected still HALF_OPEN after 1 success, got %s", mgr.State(service))
	}

	// Second success -> returns to CLOSED!
	mgr.RecordSuccess(service)
	if mgr.State(service) != CircuitClosed {
		t.Errorf("expected CLOSED after 2 successes, got %s", mgr.State(service))
	}
}
