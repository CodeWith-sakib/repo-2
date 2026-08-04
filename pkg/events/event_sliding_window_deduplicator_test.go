package events

import (
	"testing"
	"time"
)

func TestSlidingWindowDeduplicator(t *testing.T) {
	dedup := NewSlidingWindowDeduplicator(100 * time.Millisecond)

	msg := []byte("event-telemetry-packet")

	if dedup.CheckAndRecord(msg) {
		t.Fatal("first check should not be duplicate")
	}

	if !dedup.CheckAndRecord(msg) {
		t.Fatal("second check must be duplicate")
	}

	time.Sleep(120 * time.Millisecond)
	evicted := dedup.Prune(time.Now())
	if evicted != 1 {
		t.Errorf("expected 1 evicted, got %d", evicted)
	}

	if dedup.CheckAndRecord(msg) {
		t.Fatal("check after expiration should not be duplicate")
	}
}
