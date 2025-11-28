package events

import (
	"testing"
	"time"
)

func TestRollingBucketDeduplicator_DedupAndWindowRoll(t *testing.T) {
	// 2 buckets of 20ms each -> total 40ms window
	dedup, err := NewRollingBucketDeduplicator(20*time.Millisecond, 2)
	if err != nil {
		t.Fatalf("create deduplicator failed: %v", err)
	}

	// First record -> new
	if isDup := dedup.CheckAndRecord("evt-1"); isDup {
		t.Error("expected evt-1 to be new")
	}

	// Immediate duplicate -> duplicate
	if isDup := dedup.CheckAndRecord("evt-1"); !isDup {
		t.Error("expected evt-1 to be duplicate")
	}

	// Wait 50ms for window to advance past both buckets
	time.Sleep(50 * time.Millisecond)

	// Now evt-1 should have aged out
	if isDup := dedup.CheckAndRecord("evt-1"); isDup {
		t.Error("expected evt-1 to be aged out after window advance")
	}
}
