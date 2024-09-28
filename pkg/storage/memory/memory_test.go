package memory

import (
	"context"
	"testing"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/storage"
)

func TestMemoryStoreWorkflows(t *testing.T) {
	ctx := context.Background()
	store := NewStore()

	wf := &core.WorkflowDefinition{
		ID:          core.NewID("wf"),
		TenantID:    "tenant-1",
		Name:        "test-pipeline",
		Version:     1,
		Description: "integration test pipeline",
		Steps: []core.StepDefinition{
			{ID: "step-1", TaskType: "http"},
		},
	}

	if err := store.CreateWorkflow(ctx, wf); err != nil {
		t.Fatalf("failed to create workflow: %v", err)
	}

	fetched, err := store.GetWorkflow(ctx, wf.ID)
	if err != nil {
		t.Fatalf("failed to get workflow: %v", err)
	}
	if fetched.Name != wf.Name {
		t.Errorf("name mismatch: got %s, expected %s", fetched.Name, wf.Name)
	}

	// Update
	wf.Name = "updated-pipeline"
	if err := store.UpdateWorkflow(ctx, wf); err != nil {
		t.Fatalf("failed to update workflow: %v", err)
	}

	list, count, err := store.ListWorkflows(ctx, storage.WorkflowFilter{TenantID: "tenant-1"})
	if err != nil {
		t.Fatalf("failed to list workflows: %v", err)
	}
	if count != 1 || len(list) != 1 {
		t.Fatalf("expected 1 workflow, got count=%d, len=%d", count, len(list))
	}
	if list[0].Name != "updated-pipeline" {
		t.Errorf("expected updated name, got %s", list[0].Name)
	}
}

func TestMemoryStoreTaskQueue(t *testing.T) {
	ctx := context.Background()
	store := NewStore()

	task := &storage.QueuedTask{
		ID:       core.NewID("task"),
		RunID:    core.NewID("run"),
		StepID:   "extract",
		Priority: core.PriorityHigh,
	}

	if err := store.EnqueueTask(ctx, task); err != nil {
		t.Fatalf("enqueue failed: %v", err)
	}

	dequeued, err := store.DequeueTasks(ctx, "worker-1", 10, 500*time.Millisecond)
	if err != nil {
		t.Fatalf("dequeue failed: %v", err)
	}
	if len(dequeued) != 1 {
		t.Fatalf("expected 1 task dequeued, got %d", len(dequeued))
	}
	if dequeued[0].LeaseWorker != "worker-1" {
		t.Errorf("lease worker mismatch: got %s", dequeued[0].LeaseWorker)
	}

	// Ack task
	if err := store.AckTask(ctx, task.ID, "worker-1"); err != nil {
		t.Fatalf("ack failed: %v", err)
	}

	// Dequeue again should be empty
	dequeuedAgain, _ := store.DequeueTasks(ctx, "worker-1", 10, time.Second)
	if len(dequeuedAgain) != 0 {
		t.Errorf("expected 0 tasks after ack, got %d", len(dequeuedAgain))
	}
}
