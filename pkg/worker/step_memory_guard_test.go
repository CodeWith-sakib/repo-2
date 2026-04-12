package worker

import (
	"testing"
)

func TestStepMemoryGuard(t *testing.T) {
	guard := NewStepMemoryGuard(50000) // 50GB high limit for testing

	admit, err := guard.CanAdmitTask(10)
	if err != nil || !admit {
		t.Fatalf("expected task admission to pass: %v", err)
	}

	// Tiny limit should reject large allocation
	strictGuard := NewStepMemoryGuard(1) // 1MB limit
	admitStrict, err := strictGuard.CanAdmitTask(100)
	if admitStrict || err == nil {
		t.Error("expected admission rejection on memory exhaustion")
	}
}
