package events

import (
	"testing"
)

func TestDeadLetterQueue(t *testing.T) {
	dlq := NewDeadLetterQueue()
	dlq.Enqueue("evt-1", []byte("payload"), "timeout", 3)
	dlq.Enqueue("evt-2", []byte("payload"), "500 Internal Server Error", 5)

	if dlq.Count() != 2 {
		t.Errorf("expected 2 items in DLQ, got %d", dlq.Count())
	}
	items := dlq.Items()
	if items[0].EventID != "evt-1" || items[1].FailureReason != "500 Internal Server Error" {
		t.Errorf("unexpected DLQ items: %+v", items)
	}
}
