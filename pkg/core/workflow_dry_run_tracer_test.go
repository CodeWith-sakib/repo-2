package core

import (
	"context"
	"testing"
	"time"
)

func TestWorkflowDryRunTracer_Success(t *testing.T) {
	tracer := NewWorkflowDryRunTracer(50 * time.Millisecond)

	steps := map[string][]string{
		"stepA": {},
		"stepB": {"stepA"},
		"stepC": {"stepB"},
	}

	report, err := tracer.SimulateExecution(context.Background(), "wf-test-1", steps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.HasCycle {
		t.Errorf("expected no cycle")
	}
	if len(report.StepTraces) != 3 {
		t.Errorf("expected 3 step traces, got %d", len(report.StepTraces))
	}
	if report.TotalEstimatedTime != 150*time.Millisecond {
		t.Errorf("expected 150ms total time, got %v", report.TotalEstimatedTime)
	}
}

func TestWorkflowDryRunTracer_Cycle(t *testing.T) {
	tracer := NewWorkflowDryRunTracer(10 * time.Millisecond)

	steps := map[string][]string{
		"stepA": {"stepB"},
		"stepB": {"stepA"},
	}

	report, err := tracer.SimulateExecution(context.Background(), "wf-cycle-1", steps)
	if err == nil {
		t.Errorf("expected cycle error, got nil")
	}
	if !report.HasCycle {
		t.Errorf("expected HasCycle to be true")
	}
}
