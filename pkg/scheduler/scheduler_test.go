package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/storage"
	"github.com/kestrelflow/kestrelflow/pkg/storage/memory"
)

func TestSchedulerDiamondDAG(t *testing.T) {
	ctx := context.Background()
	store := memory.NewStore()
	sched := NewScheduler(store, 10*time.Millisecond)

	// Diamond DAG: A -> (B, C) -> D
	wf := &core.WorkflowDefinition{
		ID:       core.NewID("wf-diamond"),
		TenantID: "default",
		Name:     "diamond-pipeline",
		Version:  1,
		Steps: []core.StepDefinition{
			{ID: "step-a", TaskType: "shell"},
			{ID: "step-b", TaskType: "shell", DependsOn: []string{"step-a"}},
			{ID: "step-c", TaskType: "http", DependsOn: []string{"step-a"}},
			{ID: "step-d", TaskType: "shell", DependsOn: []string{"step-b", "step-c"}},
		},
	}

	if err := store.CreateWorkflow(ctx, wf); err != nil {
		t.Fatalf("failed creating workflow: %v", err)
	}

	run, err := sched.SubmitRun(ctx, wf, nil, core.PriorityNormal)
	if err != nil {
		t.Fatalf("failed submitting run: %v", err)
	}

	if run.State != core.RunStateRunning {
		t.Fatalf("expected running state, got %s", run.State)
	}

	// Dequeue step-a
	tasks, err := store.DequeueTasks(ctx, "w1", 10, time.Second)
	if err != nil || len(tasks) != 1 || tasks[0].StepID != "step-a" {
		t.Fatalf("expected step-a in queue, got %v", tasks)
	}
	_ = store.AckTask(ctx, tasks[0].ID, "w1")

	// Complete step-a
	if err := sched.HandleStepCompleted(ctx, run.ID, "step-a", []byte(`{"status":"ok"}`)); err != nil {
		t.Fatalf("failed completing step-a: %v", err)
	}

	// Both step-b and step-c should now be enqueued
	tasks, err = store.DequeueTasks(ctx, "w1", 10, time.Second)
	if err != nil || len(tasks) != 2 {
		t.Fatalf("expected 2 tasks (b & c), got %d", len(tasks))
	}

	// Ack and complete step-b
	for _, task := range tasks {
		_ = store.AckTask(ctx, task.ID, "w1")
		if task.StepID == "step-b" {
			_ = sched.HandleStepCompleted(ctx, run.ID, "step-b", nil)
		}
	}

	// step-d should NOT be enqueued yet because step-c is still running
	tasks, _ = store.DequeueTasks(ctx, "w1", 10, time.Second)
	if len(tasks) != 0 {
		t.Fatalf("step-d enqueued prematurely before step-c finished")
	}

	// Complete step-c
	_ = sched.HandleStepCompleted(ctx, run.ID, "step-c", nil)

	// Now step-d must be enqueued
	tasks, _ = store.DequeueTasks(ctx, "w1", 10, time.Second)
	if len(tasks) != 1 || tasks[0].StepID != "step-d" {
		t.Fatalf("expected step-d enqueued, got %v", tasks)
	}
	_ = store.AckTask(ctx, tasks[0].ID, "w1")

	// Complete step-d
	_ = sched.HandleStepCompleted(ctx, run.ID, "step-d", nil)

	// Workflow should now be COMPLETED
	finalRun, err := store.GetRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("failed to get final run: %v", err)
	}
	if finalRun.State != core.RunStateCompleted {
		t.Fatalf("expected completed workflow, got %s", finalRun.State)
	}
}
