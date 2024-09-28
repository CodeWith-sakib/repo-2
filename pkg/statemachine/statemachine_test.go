package statemachine

import (
	"testing"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

func TestWorkflowLegalTransitions(t *testing.T) {
	engine := NewEngine()

	tests := []struct {
		name string
		from core.RunState
		to   core.RunState
	}{
		{"Pending to Running", core.RunStatePending, core.RunStateRunning},
		{"Pending to Cancelled", core.RunStatePending, core.RunStateCancelled},
		{"Running to Completed", core.RunStateRunning, core.RunStateCompleted},
		{"Running to Failed", core.RunStateRunning, core.RunStateFailed},
		{"Running to Cancelled", core.RunStateRunning, core.RunStateCancelled},
		{"Running to Suspended", core.RunStateRunning, core.RunStateSuspended},
		{"Suspended to Running", core.RunStateSuspended, core.RunStateRunning},
		{"Suspended to Cancelled", core.RunStateSuspended, core.RunStateCancelled},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			run := &core.WorkflowRun{
				ID:        core.NewID("run"),
				State:     tc.from,
				CreatedAt: time.Now(),
			}

			if !engine.CanTransitionWorkflow(tc.from, tc.to) {
				t.Fatalf("expected transition from %s to %s to be allowed", tc.from, tc.to)
			}

			if err := engine.TransitionWorkflow(run, tc.to); err != nil {
				t.Fatalf("unexpected transition error: %v", err)
			}

			if run.State != tc.to {
				t.Errorf("state mismatch: got %s, expected %s", run.State, tc.to)
			}
		})
	}
}

func TestWorkflowIllegalTransitions(t *testing.T) {
	engine := NewEngine()

	illegalCases := []struct {
		name string
		from core.RunState
		to   core.RunState
	}{
		{"Completed to Running", core.RunStateCompleted, core.RunStateRunning},
		{"Completed to Failed", core.RunStateCompleted, core.RunStateFailed},
		{"Failed to Pending", core.RunStateFailed, core.RunStatePending},
		{"Failed to Running", core.RunStateFailed, core.RunStateRunning},
		{"Cancelled to Completed", core.RunStateCancelled, core.RunStateCompleted},
		{"Cancelled to Running", core.RunStateCancelled, core.RunStateRunning},
		{"Pending to Completed directly", core.RunStatePending, core.RunStateCompleted},
		{"Suspended to Completed directly", core.RunStateSuspended, core.RunStateCompleted},
	}

	for _, tc := range illegalCases {
		t.Run(tc.name, func(t *testing.T) {
			run := &core.WorkflowRun{
				ID:        core.NewID("run"),
				State:     tc.from,
				CreatedAt: time.Now(),
			}

			if engine.CanTransitionWorkflow(tc.from, tc.to) {
				t.Fatalf("expected transition from %s to %s to be disallowed", tc.from, tc.to)
			}

			if err := engine.TransitionWorkflow(run, tc.to); err == nil {
				t.Fatalf("expected error transitioning from %s to %s, got nil", tc.from, tc.to)
			}
		})
	}
}

func TestStepTransitions(t *testing.T) {
	engine := NewEngine()

	stepLegal := []struct {
		from core.StepState
		to   core.StepState
	}{
		{core.StepStatePending, core.StepStateQueued},
		{core.StepStatePending, core.StepStateSkipped},
		{core.StepStateQueued, core.StepStateRunning},
		{core.StepStateRunning, core.StepStateCompleted},
		{core.StepStateRunning, core.StepStateFailed},
		{core.StepStateRunning, core.StepStateRetrying},
		{core.StepStateRetrying, core.StepStateQueued},
	}

	for _, tc := range stepLegal {
		sr := &core.StepRun{
			ID:     core.NewID("step"),
			StepID: "test-step",
			State:  tc.from,
		}
		if err := engine.TransitionStep(sr, tc.to); err != nil {
			t.Fatalf("failed legal step transition %s -> %s: %v", tc.from, tc.to, err)
		}
	}

	// Test illegal step transitions
	illegal := []struct {
		from core.StepState
		to   core.StepState
	}{
		{core.StepStateCompleted, core.StepStateRunning},
		{core.StepStateFailed, core.StepStatePending},
		{core.StepStateSkipped, core.StepStateCompleted},
	}

	for _, tc := range illegal {
		sr := &core.StepRun{
			ID:     core.NewID("step"),
			StepID: "test-step",
			State:  tc.from,
		}
		if err := engine.TransitionStep(sr, tc.to); err == nil {
			t.Fatalf("expected error for step transition %s -> %s, got nil", tc.from, tc.to)
		}
	}
}
