package worker

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestStepHeartbeatRecorder(t *testing.T) {
	var flushes int64
	sink := func(ctx context.Context, report StepProgressReport) error {
		atomic.AddInt64(&flushes, 1)
		return nil
	}

	rec := NewStepHeartbeatRecorder("run-100", "step-etl", 50*time.Millisecond, sink)
	rec.UpdateProgress(50.0, 500, 1000, "Extracting records")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rec.Start(ctx)
	time.Sleep(150 * time.Millisecond)
	rec.Stop()

	count := atomic.LoadInt64(&flushes)
	if count < 2 {
		t.Errorf("expected at least 2 heartbeat flushes, got %d", count)
	}
}
