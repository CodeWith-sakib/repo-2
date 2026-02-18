package postgres

import (
	"context"
	"sync/atomic"
	"testing"
)

func TestListenNotifyBus(t *testing.T) {
	bus := NewListenNotifyBus(nil)

	var calls int64
	bus.Subscribe("workflow_events", func(event NotificationEvent) {
		atomic.AddInt64(&calls, 1)
	})

	err := bus.Notify(context.Background(), "workflow_events", map[string]string{
		"event": "RUN_COMPLETED",
		"id":    "run-101",
	})
	if err != nil {
		t.Fatalf("unexpected notify error: %v", err)
	}

	if atomic.LoadInt64(&calls) != 1 {
		t.Errorf("expected 1 subscriber call, got %d", atomic.LoadInt64(&calls))
	}
}
