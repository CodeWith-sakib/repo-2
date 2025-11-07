package events

import (
	"testing"
	"time"
)

func TestLRUEventDeduplicator_DedupAndCapacity(t *testing.T) {
	// Capacity = 2
	dedup, err := NewLRUEventDeduplicator(2, time.Hour)
	if err != nil {
		t.Fatalf("failed to create deduplicator: %v", err)
	}

	// 1st time "e1" -> not duplicate
	if isDup := dedup.CheckAndRecord("e1"); isDup {
		t.Error("e1 should be new")
	}

	// 2nd time "e1" -> duplicate!
	if isDup := dedup.CheckAndRecord("e1"); !isDup {
		t.Error("e1 should be duplicate")
	}

	// Insert "e2" -> not duplicate
	if isDup := dedup.CheckAndRecord("e2"); isDup {
		t.Error("e2 should be new")
	}

	// Insert "e3" -> exceeds capacity of 2, oldest ("e1") evicted!
	if isDup := dedup.CheckAndRecord("e3"); isDup {
		t.Error("e3 should be new")
	}

	// "e1" was evicted, so inserting "e1" now should NOT be duplicate
	if isDup := dedup.CheckAndRecord("e1"); isDup {
		t.Error("e1 should not be duplicate after being evicted from cache")
	}

	total, dups, evicted, _ := dedup.Metrics()
	if total != 5 || dups != 1 || evicted != 2 {
		t.Errorf("metrics mismatch: total=%d dups=%d evicted=%d", total, dups, evicted)
	}
}
