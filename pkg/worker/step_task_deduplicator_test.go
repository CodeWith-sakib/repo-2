package worker

import (
	"testing"
	"time"
)

func TestStepTaskDeduplicator(t *testing.T) {
	dedup := NewStepTaskDeduplicator(50 * time.Millisecond)

	key := "run-100:step-payment"
	if !dedup.TryAcquire(key) {
		t.Fatal("expected first acquire to succeed")
	}

	// Concurrent acquire should be denied
	if dedup.TryAcquire(key) {
		t.Error("expected duplicate acquire to fail")
	}

	if !dedup.IsExecuting(key) {
		t.Error("expected key to report executing")
	}

	// Release lock
	dedup.Release(key)
	if !dedup.TryAcquire(key) {
		t.Error("expected acquire to succeed after release")
	}
}
