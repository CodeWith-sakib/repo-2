package events

import (
	"context"
	"fmt"
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

func TestDLQEventRouter(t *testing.T) {
	router := NewDLQEventRouter(5)

	evt := &core.Event{ID: "evt-fail-1"}
	err := router.RouteToDLQ(context.Background(), evt, fmt.Errorf("connection refused"), 3)
	if err != nil {
		t.Fatalf("unexpected routing error: %v", err)
	}

	if router.PendingCount() != 1 {
		t.Errorf("expected 1 record in DLQ, got %d", router.PendingCount())
	}

	batch := router.PopBatch(1)
	if len(batch) != 1 || batch[0].Event.ID != "evt-fail-1" {
		t.Errorf("unexpected popped DLQ event")
	}

	if router.PendingCount() != 0 {
		t.Errorf("expected 0 records after popping batch, got %d", router.PendingCount())
	}
}
