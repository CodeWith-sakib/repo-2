package worker

import (
	"testing"
	"time"
)

func TestWorkerAffinityRouter_ConsistentHit(t *testing.T) {
	workers := []string{"worker-1", "worker-2", "worker-3"}
	router := NewWorkerAffinityRouter(workers, time.Hour)

	runID := "run-order-555"

	// First assignment -> cold (assigned via consistent hash)
	w1, isWarm := router.GetOrAssignWorker(runID)
	if isWarm {
		t.Error("expected first assignment to be cold")
	}
	if w1 == "" {
		t.Fatal("expected non-empty assigned worker")
	}

	// Subsequent lookup -> warm hit on same worker!
	w2, isWarm2 := router.GetOrAssignWorker(runID)
	if !isWarm2 {
		t.Error("expected second assignment to be warm hit")
	}
	if w2 != w1 {
		t.Errorf("worker mismatch: %s vs %s", w2, w1)
	}
}

func TestWorkerAffinityRouter_WorkerRemoval(t *testing.T) {
	workers := []string{"worker-A", "worker-B"}
	router := NewWorkerAffinityRouter(workers, time.Hour)

	runID := "run-99"
	w1, _ := router.GetOrAssignWorker(runID)

	// Remove assigned worker from active list
	remaining := []string{"worker-other"}
	if w1 != "worker-A" {
		remaining = []string{"worker-A"}
	}
	router.UpdateWorkers(remaining)

	// Next lookup should reassign because previous worker is gone
	w2, isWarm := router.GetOrAssignWorker(runID)
	if isWarm {
		t.Error("should not be warm hit after worker pool removal")
	}
	if w2 == w1 {
		t.Errorf("should reassign to active worker, still got %s", w2)
	}
}
