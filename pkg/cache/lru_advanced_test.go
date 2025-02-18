package cache

import (
	"testing"
	"time"
)

func TestTwoQueueCache(t *testing.T) {
	c := NewTwoQueueCache(4)

	c.Set("k1", "v1", 0)
	c.Set("k2", "v2", 0)
	c.Set("k3", "v3", 0)
	c.Set("k4", "v4", 0)

	val, ok := c.Get("k1")
	if !ok || val != "v1" {
		t.Fatalf("expected v1, got %v", val)
	}

	// Adding k5 should evict from A1in
	c.Set("k5", "v5", 0)

	hits, misses, count := c.Stats()
	if count > 4 {
		t.Errorf("cache size exceeded capacity: %d", count)
	}
	if hits != 1 || misses != 0 {
		t.Errorf("unexpected stats: hits=%d, misses=%d", hits, misses)
	}

	// Test TTL
	c.Set("short", "lived", 10*time.Millisecond)
	time.Sleep(20 * time.Millisecond)
	if _, ok := c.Get("short"); ok {
		t.Error("expected expired key to miss")
	}
}
