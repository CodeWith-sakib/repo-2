package worker

import (
	"context"
	"testing"
)

func TestPriorityWorkerStealing(t *testing.T) {
	w1 := NewPriorityWorker("w1", 10)
	w2 := NewPriorityWorker("w2", 10)

	w1.AddPeerQueue(w2.lowQueue)

	// Submit task to w2
	if !w2.Submit("task-w2", false) {
		t.Fatal("failed submitting to w2")
	}

	// w1 polls and steals from w2
	task, ok := w1.Poll(context.Background())
	if !ok || task != "task-w2" {
		t.Fatalf("expected stolen task-w2, got %s (ok=%v)", task, ok)
	}

	stats := w1.GetStats()
	if stats.TasksStolen != 1 || stats.TasksExecuted != 1 {
		t.Errorf("unexpected stats: %+v", stats)
	}
}
