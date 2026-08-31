package worker

import (
	"testing"
)

func TestCgroupMemoryPressureWatcher(t *testing.T) {
	watcher := NewCgroupMemoryPressureWatcher(10.0, 25.0)

	m1 := watcher.RecordPSIObservation(2.5, 1.8, 0.0)
	if m1.Level != PressureNone {
		t.Errorf("expected NONE, got %v", m1.Level)
	}

	m2 := watcher.RecordPSIObservation(15.0, 8.0, 2.0)
	if m2.Level != PressureModerate {
		t.Errorf("expected MODERATE, got %v", m2.Level)
	}

	m3 := watcher.RecordPSIObservation(35.0, 28.0, 15.0)
	if m3.Level != PressureCritical {
		t.Errorf("expected CRITICAL, got %v", m3.Level)
	}
}
