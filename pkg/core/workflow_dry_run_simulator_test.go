package core

import (
	"context"
	"testing"
	"time"
)

func TestWorkflowDryRunSimulator(t *testing.T) {
	sim := NewWorkflowDryRunSimulator(50 * time.Millisecond)

	wf := &WorkflowDefinition{
		ID:       NewID("wf_sim_01"),
		TenantID: "tenant-sim",
		Name:     "Simulation Test DAG",
		Version:  1,
		Steps: []StepDefinition{
			{ID: "s1", TaskType: "init"},
			{ID: "s2", TaskType: "process", DependsOn: []string{"s1"}},
			{ID: "s3", TaskType: "export", DependsOn: []string{"s2"}},
		},
	}

	report, err := sim.Simulate(context.Background(), wf)
	if err != nil {
		t.Fatalf("unexpected simulation error: %v", err)
	}

	if !report.Passed {
		t.Fatalf("expected simulation report to pass: %v", report.ValidationErrors)
	}
	if report.TotalSteps != 3 {
		t.Errorf("expected 3 steps, got %d", report.TotalSteps)
	}
	if report.ExecutionWaves != 3 {
		t.Errorf("expected 3 waves, got %d", report.ExecutionWaves)
	}
	if report.EstimatedDuration != 150*time.Millisecond {
		t.Errorf("expected 150ms estimated duration, got %v", report.EstimatedDuration)
	}
}
