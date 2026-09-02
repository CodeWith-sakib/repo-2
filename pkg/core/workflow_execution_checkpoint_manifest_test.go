package core

import (
	"testing"
)

func TestWorkflowExecutionCheckpointManifest(t *testing.T) {
	manifest := NewWorkflowExecutionCheckpointManifest("wf-cp-1", "run-500", 1)

	chk := manifest.RecordStepCheckpoint("step-transform", "COMPLETED", `{"rows_processed": 10000}`, 1, map[string]string{"partition": "p0"})
	if chk.StepID != "step-transform" {
		t.Errorf("expected step-transform, got %s", chk.StepID)
	}
	if chk.StateChecksum == "" {
		t.Error("expected non-empty state checksum")
	}

	retrieved, ok := manifest.GetCheckpoint("step-transform")
	if !ok || retrieved.Status != "COMPLETED" {
		t.Fatalf("checkpoint lookup failed: %v", retrieved)
	}

	completed := manifest.CompletedStepIDs()
	if len(completed) != 1 || completed[0] != "step-transform" {
		t.Errorf("expected 1 completed step, got %v", completed)
	}
}
