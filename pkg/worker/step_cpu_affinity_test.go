package worker

import (
	"testing"
)

func TestWorkerCPUAffinityManager(t *testing.T) {
	mgr := NewWorkerCPUAffinityManager()

	core0, ok := mgr.PinTask("task-1")
	if !ok || core0 < 0 {
		t.Fatalf("expected task-1 pinned to valid core, got %d", core0)
	}

	// Duplicate pin returns same core
	core0Again, ok := mgr.PinTask("task-1")
	if !ok || core0Again != core0 {
		t.Errorf("expected duplicate pin to return identical core %d, got %d", core0, core0Again)
	}

	if mgr.ActivePinnedCount() != 1 {
		t.Errorf("expected 1 pinned task, got %d", mgr.ActivePinnedCount())
	}

	mgr.UnpinTask("task-1")
	if mgr.ActivePinnedCount() != 0 {
		t.Errorf("expected 0 pinned tasks after unpin, got %d", mgr.ActivePinnedCount())
	}
}
