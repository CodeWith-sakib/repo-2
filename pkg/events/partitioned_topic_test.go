package events

import (
	"testing"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

func TestPartitionedTopic_PublishAndRead(t *testing.T) {
	topic, err := NewPartitionedTopic("orders", 3)
	if err != nil {
		t.Fatalf("create topic failed: %v", err)
	}

	// Publish with consistent routing key -> always lands in same partition
	key := "tenant-acme-user-99"
	evt1 := &core.Event{ID: "e1", Type: "order.created", Timestamp: time.Now()}
	evt2 := &core.Event{ID: "e2", Type: "order.paid", Timestamp: time.Now()}

	pe1, err := topic.Publish(evt1, key)
	if err != nil {
		t.Fatalf("publish 1 failed: %v", err)
	}
	pe2, err := topic.Publish(evt2, key)
	if err != nil {
		t.Fatalf("publish 2 failed: %v", err)
	}

	if pe1.Partition != pe2.Partition {
		t.Errorf("expected same partition for identical key: %d vs %d", pe1.Partition, pe2.Partition)
	}
	if pe2.Offset != pe1.Offset+1 {
		t.Errorf("expected consecutive offsets: %d and %d", pe1.Offset, pe2.Offset)
	}

	// Read from offset 0
	events, err := topic.ReadFromOffset(pe1.Partition, 0, 10)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}

	// Commit offset for group "order-worker"
	group := "order-worker"
	if err := topic.CommitOffset(group, pe1.Partition, pe2.Offset); err != nil {
		t.Fatalf("commit failed: %v", err)
	}

	committed := topic.GetCommittedOffset(group, pe1.Partition)
	if committed != pe2.Offset {
		t.Errorf("expected committed offset %d, got %d", pe2.Offset, committed)
	}
}
