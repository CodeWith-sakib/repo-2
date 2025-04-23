package events

import (
	"testing"
	"time"
)

func TestDeduplicator(t *testing.T) {
	dedup := NewDeduplicator(50*time.Millisecond, 20*time.Millisecond)
	defer dedup.Close()

	if dedup.IsDuplicate("evt-1") {
		t.Error("first event should not be duplicate")
	}

	if !dedup.IsDuplicate("evt-1") {
		t.Error("immediate second event should be duplicate")
	}

	// Wait for TTL expiry
	time.Sleep(60 * time.Millisecond)
	if dedup.IsDuplicate("evt-1") {
		t.Error("event after TTL should not be duplicate")
	}
}
