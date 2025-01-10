package webhook

import (
	"testing"
	"time"
)

func TestDeliveryQueueLifecycle(t *testing.T) {
	q := NewDeliveryQueue()

	item := q.Enqueue("del-1", "https://example.com/hook", "secret", []byte("{}"), 3)
	if item.Status != StatusPending {
		t.Fatalf("expected PENDING, got %s", item.Status)
	}

	due := q.GetDueDeliveries(time.Now().Add(1*time.Second), 10)
	if len(due) != 1 || due[0].ID != "del-1" {
		t.Fatalf("expected del-1 due, got %v", due)
	}
	if due[0].Status != StatusInFlight {
		t.Fatalf("expected IN_FLIGHT, got %s", due[0].Status)
	}

	// Fail attempt 1
	_ = q.MarkFailed("del-1", "network timeout", 10*time.Millisecond)
	if item.Attempts != 1 || item.Status != StatusPending {
		t.Fatalf("expected 1 attempt and PENDING, got %d / %s", item.Attempts, item.Status)
	}

	// Fail attempt 2
	_ = q.MarkFailed("del-1", "500 Internal Server Error", 10*time.Millisecond)
	// Fail attempt 3 -> should become FAILED
	_ = q.MarkFailed("del-1", "500 Internal Server Error", 10*time.Millisecond)
	if item.Status != StatusFailed {
		t.Fatalf("expected StatusFailed, got %s", item.Status)
	}
}
