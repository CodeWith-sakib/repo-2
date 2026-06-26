package worker

import (
	"testing"
	"time"
)

func TestStepFailureCircuitBreaker(t *testing.T) {
	cb := NewStepFailureCircuitBreaker(3, 100*time.Millisecond)

	if !cb.AllowExecution() || cb.Status() != StepCircuitClosed {
		t.Fatal("expected circuit closed initially")
	}

	cb.RecordFailure()
	cb.RecordFailure()
	if cb.Status() != StepCircuitClosed {
		t.Fatal("circuit should still be closed at 2 failures")
	}

	cb.RecordFailure() // 3rd failure trips
	if cb.Status() != StepCircuitOpen {
		t.Fatal("circuit should be open after 3 failures")
	}
	if cb.AllowExecution() {
		t.Fatal("AllowExecution should return false when open")
	}

	time.Sleep(120 * time.Millisecond)
	if !cb.AllowExecution() {
		t.Fatal("AllowExecution should allow probe after cooldown")
	}
	if cb.Status() != StepCircuitHalfOpen {
		t.Fatalf("expected HALF_OPEN, got %v", cb.Status())
	}

	cb.RecordSuccess()
	if cb.Status() != StepCircuitClosed {
		t.Fatalf("expected CLOSED after success, got %v", cb.Status())
	}
}
