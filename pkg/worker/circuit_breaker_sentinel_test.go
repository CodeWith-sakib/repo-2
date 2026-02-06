package worker

import (
	"errors"
	"testing"
	"time"
)

func TestCircuitBreakerSentinel(t *testing.T) {
	cb := NewCircuitBreakerSentinel(3, 50*time.Millisecond)

	if !cb.Allow() {
		t.Fatal("expected initially closed breaker to allow")
	}

	// 2 failures -> stays closed
	cb.RecordResult(errors.New("fail 1"))
	cb.RecordResult(errors.New("fail 2"))
	if cb.State() != CircuitStateClosed {
		t.Errorf("expected closed state, got %s", cb.State())
	}

	// 3rd failure -> trips open
	cb.RecordResult(errors.New("fail 3"))
	if cb.State() != CircuitStateOpen {
		t.Errorf("expected open state, got %s", cb.State())
	}
	if cb.Allow() {
		t.Error("expected open breaker to reject execution")
	}

	// Wait for reset timeout -> transitions to HALF_OPEN
	time.Sleep(60 * time.Millisecond)
	if !cb.Allow() {
		t.Error("expected trial request allowed in HALF_OPEN")
	}

	// Successful probe resets to closed
	cb.RecordResult(nil)
	if cb.State() != CircuitStateClosed {
		t.Errorf("expected closed state after success, got %s", cb.State())
	}
}
