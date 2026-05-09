package events

import (
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

func TestEventStreamPartitioner(t *testing.T) {
	part := NewEventStreamPartitioner(16)

	p1 := part.AssignPartition("tenant-alpha")
	p2 := part.AssignPartition("tenant-alpha")
	if p1 != p2 {
		t.Errorf("expected deterministic partition index: %d vs %d", p1, p2)
	}

	if p1 < 0 || p1 >= 16 {
		t.Errorf("partition index out of range: %d", p1)
	}

	evt := &core.Event{
		RunID: "run-100",
		ID:         "evt-01",
	}
	evtPart := part.PartitionEvent(evt)
	if evtPart < 0 || evtPart >= 16 {
		t.Errorf("event partition out of range: %d", evtPart)
	}
}
