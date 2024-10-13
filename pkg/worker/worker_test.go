package worker

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"testing"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/scheduler"
	"github.com/kestrelflow/kestrelflow/pkg/storage/memory"
)

type MockExecutor struct {
	executedCount int32
}

func (m *MockExecutor) Execute(ctx context.Context, sctx StepContext) (*StepResult, error) {
	atomic.AddInt32(&m.executedCount, 1)
	return &StepResult{Output: json.RawMessage(`{"result":"ok"}`)}, nil
}

func TestWorkerPoolExecutionCleanRace(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	store := memory.NewStore()
	sched := scheduler.NewScheduler(store, 10*time.Millisecond)
	reg := NewExecutorRegistry()

	mockExec := &MockExecutor{}
	reg.Register("mock", mockExec)

	cfg := Config{
		WorkerID:          "worker-node-1",
		Concurrency:       4,
		PollInterval:      5 * time.Millisecond,
		LeaseDuration:     time.Second,
		HeartbeatInterval: 100 * time.Millisecond,
	}
	pool := NewPool(cfg, store, sched, reg)

	go pool.Start(ctx)
	defer pool.Stop()

	// Create workflow with 3 parallel steps
	wf := &core.WorkflowDefinition{
		ID:       core.NewID("wf-parallel"),
		TenantID: "tenant-a",
		Name:     "parallel-workflow",
		Version:  1,
		Steps: []core.StepDefinition{
			{ID: "task-1", TaskType: "mock"},
			{ID: "task-2", TaskType: "mock"},
			{ID: "task-3", TaskType: "mock"},
		},
	}

	if err := store.CreateWorkflow(ctx, wf); err != nil {
		t.Fatalf("failed to create workflow: %v", err)
	}

	run, err := sched.SubmitRun(ctx, wf, nil, core.PriorityHigh)
	if err != nil {
		t.Fatalf("failed to submit run: %v", err)
	}

	// Wait for completion
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		r, _ := store.GetRun(ctx, run.ID)
		if r != nil && r.State == core.RunStateCompleted {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	r, err := store.GetRun(ctx, run.ID)
	if err != nil || r.State != core.RunStateCompleted {
		t.Fatalf("expected run completed, got %v (state: %v)", err, r.State)
	}

	if count := atomic.LoadInt32(&mockExec.executedCount); count != 3 {
		t.Fatalf("expected 3 tasks executed, got %d", count)
	}
}
