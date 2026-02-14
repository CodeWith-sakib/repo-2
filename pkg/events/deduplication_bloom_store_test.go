package events

import (
	"testing"
	"time"
)

func TestDeduplicationBloomStore(t *testing.T) {
	store := NewDeduplicationBloomStore(1000, 10*time.Minute)

	// First occurrence
	if store.AddOrCheck("evt_001") {
		t.Error("expected evt_001 to be fresh, not duplicate")
	}

	// Immediate repeat -> duplicate
	if !store.AddOrCheck("evt_001") {
		t.Error("expected evt_001 to be recognized as duplicate")
	}

	// Distinct event
	if store.AddOrCheck("evt_002") {
		t.Error("expected evt_002 to be fresh")
	}

	// Expiration check
	now := time.Now().UTC()
	if store.IsExpired(now) {
		t.Error("store should not be expired immediately")
	}
	if !store.IsExpired(now.Add(15 * time.Minute)) {
		t.Error("store should report expired after TTL")
	}
}
