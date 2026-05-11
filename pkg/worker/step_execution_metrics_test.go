package worker

import (
	"testing"
	"time"
)

func TestStepExecutionTelemetry(t *testing.T) {
	m := NewStepExecutionTelemetry()

	m.RecordExecution(20*time.Millisecond, true)
	m.RecordExecution(40*time.Millisecond, true)
	m.RecordExecution(60*time.Millisecond, false)

	if m.TotalRuns.Load() != 3 {
		t.Errorf("expected 3 total runs, got %d", m.TotalRuns.Load())
	}
	if m.SuccessRuns.Load() != 2 {
		t.Errorf("expected 2 success runs, got %d", m.SuccessRuns.Load())
	}

	mean := m.MeanDuration()
	if mean != 40*time.Millisecond {
		t.Errorf("expected 40ms mean duration, got %v", mean)
	}

	maxDur := time.Duration(m.MaxDuration.Load())
	if maxDur != 60*time.Millisecond {
		t.Errorf("expected 60ms max duration, got %v", maxDur)
	}
}
