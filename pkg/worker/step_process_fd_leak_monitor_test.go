package worker

import (
	"context"
	"testing"
)

func TestProcessFDLeakMonitor(t *testing.T) {
	monitor := NewProcessFDLeakMonitor(1000)

	m1 := monitor.RecordFDObservation(context.Background(), 200)
	if m1.IsExhausted {
		t.Error("200/1000 FDs should not trigger exhaustion")
	}

	m2 := monitor.RecordFDObservation(context.Background(), 850)
	if !m2.IsExhausted {
		t.Error("850/1000 FDs should trigger exhaustion warning")
	}
	if monitor.MaxCap() != 1000 {
		t.Errorf("expected max cap 1000, got %d", monitor.MaxCap())
	}
}
