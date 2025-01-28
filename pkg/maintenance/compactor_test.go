package maintenance

import (
	"context"
	"testing"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/storage/memory"
)

func TestHistoryCompactor(t *testing.T) {
	store := memory.NewMemoryStore()
	ctx := context.Background()

	wf := &core.Workflow{ID: core.NewID("wf"), Name: "test", Version: 1}
	_ = store.CreateWorkflow(ctx, wf)

	completedTime := time.Now().Add(-40 * 24 * time.Hour)
	run := &core.WorkflowRun{
		ID:          core.NewID("run-old"),
		WorkflowID:  wf.ID,
		State:       core.RunStateCompleted,
		CompletedAt: &completedTime,
	}
	_ = store.CreateRun(ctx, run)

	compactor := NewHistoryCompactor(store, CompactionPolicy{
		RetentionPeriod: 30 * 24 * time.Hour,
		BatchSize:       10,
	})

	stats, err := compactor.Compact(ctx)
	if err != nil {
		t.Fatalf("compaction failed: %v", err)
	}
	if stats.RunsEvaluated != 1 {
		t.Fatalf("expected 1 run evaluated, got %d", stats.RunsEvaluated)
	}
	if stats.RunsPruned != 1 {
		t.Fatalf("expected 1 run pruned, got %d", stats.RunsPruned)
	}
}
