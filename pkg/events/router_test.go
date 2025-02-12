package events

import (
	"sync/atomic"
	"testing"
)

func TestEventRouter(t *testing.T) {
	r := NewEventRouter()

	var wildcardCalls int32
	var stepCalls int32
	var exactCalls int32

	r.Subscribe("*", func(e *EventEnvelope) error {
		atomic.AddInt32(&wildcardCalls, 1)
		return nil
	})

	r.Subscribe("step.*", func(e *EventEnvelope) error {
		atomic.AddInt32(&stepCalls, 1)
		return nil
	})

	r.Subscribe("step.completed", func(e *EventEnvelope) error {
		atomic.AddInt32(&exactCalls, 1)
		return nil
	})

	evt := &EventEnvelope{Type: "step.completed"}
	if err := r.Dispatch(evt); err != nil {
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
