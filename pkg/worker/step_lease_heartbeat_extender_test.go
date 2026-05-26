package worker

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestStepLeaseHeartbeatExtender(t *testing.T) {
	var renewCalls int64
	hook := func(ctx context.Context, stepRunID string, duration time.Duration) error {
		atomic.AddInt64(&renewCalls, 1)
		return nil
	}

	extender := NewStepLeaseHeartbeatExtender("step-run-101", 30*time.Millisecond, 100*time.Millisecond, hook)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	extender.Start(ctx)
	time.Sleep(100 * time.Millisecond)
	extender.Stop()

	calls := atomic.LoadInt64(&renewCalls)
	if calls < 2 {
		t.Errorf("expected at least 2 renewals, got %d", calls)
	}
}
