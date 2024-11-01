package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/scheduler"
	"github.com/kestrelflow/kestrelflow/pkg/storage/memory"
)

func TestGRPCServiceBoundaries(t *testing.T) {
	ctx := context.Background()
	store := memory.NewStore()
	sched := scheduler.NewScheduler(store, 10*time.Millisecond)
	svc := NewService(store, sched)

	// Negative case: submit with empty workflow ID
	_, err := svc.SubmitWorkflowRun(ctx, SubmitRunReq{})
	if err == nil {
		t.Fatal("expected error for empty workflow ID, got nil")
	}

	// Negative case: get non-existent run
	_, err = svc.GetWorkflowRun(ctx, "run-does-not-exist")
	if err == nil {
		t.Fatal("expected error for non-existent run, got nil")
	}

	// Negative case: cancel empty run ID
	_, err = svc.CancelWorkflowRun(ctx, CancelRunReq{})
	if err == nil {
		t.Fatal("expected error for empty run ID on cancel, got nil")
	}

	// Positive workflow creation and submit
	wf := &core.WorkflowDefinition{
		ID:       core.NewID("wf"),
		TenantID: "default",
		Name:     "grpc-test-wf",
		Version:  1,
		Steps: []core.StepDefinition{
			{ID: "s1", TaskType: "http"},
		},
	}
	_ = store.CreateWorkflow(ctx, wf)

	status, err := svc.SubmitWorkflowRun(ctx, SubmitRunReq{WorkflowID: string(wf.ID)})
	if err != nil {
		t.Fatalf("failed submitting run: %v", err)
	}
	if status.State != string(core.RunStateRunning) {
		t.Errorf("expected running, got %s", status.State)
	}

	// Cancel via gRPC
	cancelledStatus, err := svc.CancelWorkflowRun(ctx, CancelRunReq{RunID: status.RunID})
	if err != nil {
		t.Fatalf("failed cancelling run: %v", err)
	}
	if cancelledStatus.State != string(core.RunStateCancelled) {
		t.Errorf("expected cancelled, got %s", cancelledStatus.State)
	}

	// Negative case: cancel already cancelled run
	_, err = svc.CancelWorkflowRun(ctx, CancelRunReq{RunID: status.RunID})
	if err == nil {
		t.Fatal("expected error when cancelling already cancelled run, got nil")
	}
}
