package core

import (
	"context"
	"testing"
	"time"
)

func TestWorkflowSignalChannel(t *testing.T) {
	hub := NewWorkflowSignalChannel()

	sig := WorkflowSignal{
		Name:      "user_approval",
		RunID:     "run-order-99",
		Payload:   "APPROVED",
		Timestamp: time.Now().UTC(),
	}

	go func() {
		time.Sleep(20 * time.Millisecond)
		_ = hub.Send(sig)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	received, err := hub.WaitSignal(ctx, "run-order-99", "user_approval")
	if err != nil {
		t.Fatalf("unexpected wait signal error: %v", err)
	}

	if received.Payload.(string) != "APPROVED" {
		t.Errorf("expected APPROVED payload, got %v", received.Payload)
	}
}
