package events

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

func TestEventRouter(t *testing.T) {
	r := NewEventRouter()
	ctx := context.Background()

	var wildcardCalls int32
	var stepCalls int32
	var exactCalls int32

	r.Subscribe("*", func(ctx context.Context, e *core.Event) error {
		atomic.AddInt32(&wildcardCalls, 1)
		return nil
	})

	r.Subscribe("step.*", func(ctx context.Context, e *core.Event) error {
		atomic.AddInt32(&stepCalls, 1)
		return nil
	})

	r.Subscribe("step.completed", func(ctx context.Context, e *core.Event) error {
		atomic.AddInt32(&exactCalls, 1)
		return nil
	})

	evt := &core.Event{Type: core.EventStepCompleted}
	if err := r.Dispatch(ctx, evt); err != nil {
		t.Fatalf("dispatch failed: %v", err)
	}

	if atomic.LoadInt32(&wildcardCalls) != 1 {
		t.Errorf("expected 1 wildcard call, got %d", wildcardCalls)
	}
	if atomic.LoadInt32(&stepCalls) != 1 {
		t.Errorf("expected 1 step.* call, got %d", stepCalls)
	}
	if atomic.LoadInt32(&exactCalls) != 1 {
		t.Errorf("expected 1 exact call, got %d", exactCalls)
	}
}
