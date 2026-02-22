package events

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

func TestBatchEventFlusher(t *testing.T) {
	var flushedCount int64
	handler := func(ctx context.Context, batch []*core.Event) error {
		atomic.AddInt64(&flushedCount, int64(len(batch)))
		return nil
	}

	flusher := NewBatchEventFlusher(3, 50*time.Millisecond, handler)
	ctx := context.Background()

	// Push 2 events (no auto-flush yet)
	_ = flusher.Push(ctx, &core.Event{ID: "e1"})
	_ = flusher.Push(ctx, &core.Event{ID: "e2"})

	if atomic.LoadInt64(&flushedCount) != 0 {
		t.Errorf("expected 0 flushed, got %d", atomic.LoadInt64(&flushedCount))
	}

	// 3rd event reaches maxBatchSize -> auto flushes
	_ = flusher.Push(ctx, &core.Event{ID: "e3"})
	if atomic.LoadInt64(&flushedCount) != 3 {
		t.Errorf("expected 3 flushed, got %d", atomic.LoadInt64(&flushedCount))
	}

	// Test timer flush
	flusher.Start(ctx)
	_ = flusher.Push(ctx, &core.Event{ID: "e4"})
	time.Sleep(100 * time.Millisecond)
	flusher.Stop()

	if atomic.LoadInt64(&flushedCount) != 4 {
		t.Errorf("expected 4 total flushed after timer, got %d", atomic.LoadInt64(&flushedCount))
	}
}
