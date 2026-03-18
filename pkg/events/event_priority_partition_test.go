package events

import (
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

func TestPriorityPartitionRouter(t *testing.T) {
	router := NewPriorityPartitionRouter(8)

	part1 := router.RoutePartition("tenant-a", "entity-123")
	part2 := router.RoutePartition("tenant-a", "entity-123")
	if part1 != part2 {
		t.Errorf("expected deterministic partition routing: %d vs %d", part1, part2)
	}

	if part1 < 0 || part1 >= 8 {
		t.Errorf("partition out of bounds: %d", part1)
	}

	topic := router.TargetTopic("orders", "critical")
	if topic != "orders.critical" {
		t.Errorf("expected orders.critical, got %s", topic)
	}

	if !router.ValidateEvent(&core.Event{ID: "e1"}) {
		t.Error("expected valid event to pass")
	}
}
