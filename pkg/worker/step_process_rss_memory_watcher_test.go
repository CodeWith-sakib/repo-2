package worker

import (
	"context"
	"testing"
)

func TestProcessRSSMemoryWatcher(t *testing.T) {
	watcher := NewProcessRSSMemoryWatcher(100 * 1024 * 1024) // 100MB limit

	m := watcher.SampleMemory(context.Background())
	if m.AllocBytes == 0 {
		t.Error("expected non-zero memory allocation")
	}
	if watcher.EmergencyGCTriggerCount() != 0 {
		t.Error("expected 0 emergency GC triggers")
	}
}
