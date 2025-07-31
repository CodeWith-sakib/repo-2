package worker

import (
	"testing"
)

func TestTaskFlightTracker(t *testing.T) {
	tr := NewTaskFlightTracker()

	tr.Track("t-1", "w-1")
	tr.Track("t-2", "w-1")
	tr.Track("t-3", "w-2")

	if tr.InFlightCount() != 3 {
		t.Errorf("expected 3 inflight, got %d", tr.InFlightCount())
	}

	w1Tasks := tr.ActiveTasksForWorker("w-1")
	if len(w1Tasks) != 2 {
		t.Errorf("expected 2 tasks for w-1, got %d", len(w1Tasks))
	}

	tr.Complete("t-1")
	if tr.InFlightCount() != 2 {
		t.Errorf("expected 2 inflight after completion, got %d", tr.InFlightCount())
	}
}
