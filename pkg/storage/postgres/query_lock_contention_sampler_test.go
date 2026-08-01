package postgres

import (
	"context"
	"testing"
	"time"
)

func TestLockContentionSampler(t *testing.T) {
	sampler := NewLockContentionSampler(100 * time.Millisecond)

	sampler.RecordLockWait(context.Background(), "kf_workflows", "RowExclusiveLock", 50*time.Millisecond, 101)
	sampler.RecordLockWait(context.Background(), "kf_step_runs", "AccessExclusiveLock", 250*time.Millisecond, 102)

	if sampler.CriticalContentionCount() != 1 {
		t.Errorf("expected 1 critical contention event, got %d", sampler.CriticalContentionCount())
	}
	if sampler.TotalContentionTime() != 300*time.Millisecond {
		t.Errorf("expected 300ms total contention time, got %v", sampler.TotalContentionTime())
	}
}
