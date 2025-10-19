package worker

import (
	"context"
	"testing"
	"time"
)

func TestPriorityLeaseCoordinator_Preemption(t *testing.T) {
	cfg := PreemptionConfig{
		DefaultTTL:      time.Hour,
		MinPreemptDelta: 20,
	}
	coord := NewPriorityLeaseCoordinator(cfg)

	// Low priority worker (p=10) acquires lease
	lease1, ctx1, err := coord.AcquireLease(context.Background(), "task-1", "worker-low", 10)
	if err != nil {
		t.Fatalf("first acquire failed: %v", err)
	}
	if lease1.WorkerID != "worker-low" {
		t.Errorf("expected worker-low, got %s", lease1.WorkerID)
	}

	// Medium priority worker (p=25) tries to acquire -> delta is only 15 (< 20 threshold) -> fails
	_, _, err = coord.AcquireLease(context.Background(), "task-1", "worker-mid", 25)
	if err == nil {
		t.Fatal("expected error due to insufficient priority delta")
	}

	// High priority worker (p=50) acquires -> delta is 40 (>= 20) -> preempts worker-low!
	lease2, _, err := coord.AcquireLease(context.Background(), "task-1", "worker-high", 50)
	if err != nil {
		t.Fatalf("preemption acquire failed: %v", err)
	}
	if lease2.WorkerID != "worker-high" {
		t.Errorf("expected worker-high, got %s", lease2.WorkerID)
	}

	// ctx1 for worker-low should now be cancelled
	select {
	case <-ctx1.Done():
		// Expected cancellation!
	case <-time.After(100 * time.Millisecond):
		t.Fatal("preempted worker context was not cancelled")
	}

	if coord.PreemptionsCount() != 1 {
		t.Errorf("expected 1 preemption, got %d", coord.PreemptionsCount())
	}
}
