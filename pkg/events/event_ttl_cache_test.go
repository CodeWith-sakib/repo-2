package events

import (
	"testing"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

func TestEventTTLCache(t *testing.T) {
	cache := NewEventTTLCache(100 * time.Millisecond)

	evt := &core.Event{ID: "e1", Type: "run.started"}
	cache.Put(evt)

	retrieved, ok := cache.Get("e1")
	if !ok || retrieved.ID != "e1" {
		t.Fatalf("expected to retrieve e1, got ok=%v", ok)
	}

	// Wait for TTL expiration
	time.Sleep(120 * time.Millisecond)
	_, ok = cache.Get("e1")
	if ok {
		t.Error("expected e1 to be expired")
	}

	// Purge expired
	purged := cache.SweepExpired(time.Now().UTC())
	if purged != 1 {
		t.Errorf("expected 1 item purged, got %d", purged)
	}
}
