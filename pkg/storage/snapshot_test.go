package storage

import (
	"context"
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

func TestSnapshotManagerCheckpoint(t *testing.T) {
	sm := NewSnapshotManager()
	runID := core.NewID("run-snap")

	run := &core.WorkflowRun{
		ID:         runID,
		WorkflowID: core.NewID("wf"),
		Version:    1,
		State:      core.RunStateRunning,
	}

	steps := []*core.StepRun{
		{StepID: "step-1", State: core.StepStateCompleted, Output: []byte(`{"res":1}`)},
		{StepID: "step-2", State: core.StepStateRunning},
	}

	snap, err := sm.CreateCheckpoint(context.Background(), run, steps)
	if err != nil {
		t.Fatalf("failed creating checkpoint: %v", err)
	}

	if snap.StepStates["step-1"] != core.StepStateCompleted {
		t.Errorf("expected step-1 completed, got %s", snap.StepStates["step-1"])
	}

	latest, err := sm.GetLatest(runID)
	if err != nil || latest.RunID != runID {
		t.Fatalf("failed getting latest snapshot: %v", err)
	}
}
