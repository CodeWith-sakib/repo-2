package worker

import (
	"testing"
	"time"
)

func TestCooperativeHandoffCoordinator_Lifecycle(t *testing.T) {
	coord := NewCooperativeHandoffCoordinator("worker-node-1", 5*time.Second)
	payload := map[string]interface{}{"step": "compute-pi", "iteration": 1000}

	msg, err := coord.InitiateHandoff("task-xyz", "worker-node-2", payload, 42)
	if err != nil {
		t.Fatalf("failed to initiate handoff: %v", err)
	}

	if msg.State != HandoffStateInitiated {
		t.Errorf("expected initiated state, got %s", msg.State)
	}

	// Ack from wrong worker should fail
	if err := coord.AcknowledgeHandoff("task-xyz", "worker-node-99"); err == nil {
		t.Errorf("expected error acknowledging from wrong target worker")
	}

	// Ack from correct worker
	if err := coord.AcknowledgeHandoff("task-xyz", "worker-node-2"); err != nil {
		t.Fatalf("acknowledge failed: %v", err)
	}

	// Commit handoff
	if err := coord.CommitHandoff("task-xyz"); err != nil {
		t.Fatalf("commit failed: %v", err)
	}

	final, ok := coord.GetHandoff("task-xyz")
	if !ok || final.State != HandoffStateCommitted {
		t.Errorf("expected committed state, got %v", final)
	}
}
