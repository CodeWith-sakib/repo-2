package events

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"
)

// RollingDedupItem holds arrival timestamp for an event message digest.
type RollingDedupItem struct {
	Digest    string
	ArrivedAt time.Time
}

// SlidingWindowDeduplicator guarantees at-most-once processing semantics within a sliding time window.
type SlidingWindowDeduplicator struct {
	mu     sync.Mutex
	window time.Duration
	items  map[string]time.Time
}

// NewSlidingWindowDeduplicator constructs a deduplicator.
func NewSlidingWindowDeduplicator(window time.Duration) *SlidingWindowDeduplicator {
	if window <= 0 {
		window = 10 * time.Minute
	}
	return &SlidingWindowDeduplicator{
		window: window,
		items:  make(map[string]time.Time),
	}
}

// CheckAndRecord returns true if the item was already processed within window. If not, records it.
func (d *SlidingWindowDeduplicator) CheckAndRecord(payload []byte) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	hash := sha256.Sum256(payload)
	digest := hex.EncodeToString(hash[:16])
	now := time.Now()

	if ts, exists := d.items[digest]; exists {
		if now.Sub(ts) <= d.window {
			return true // duplicate
		}
	}

	d.items[digest] = now
	return false
}

// Prune cleans out records that have aged beyond window duration.
func (d *SlidingWindowDeduplicator) Prune(now time.Time) int {
	d.mu.Lock()
	defer d.mu.Unlock()

	evicted := 0
	for dig, ts := range d.items {
		if now.Sub(ts) > d.window {
			delete(d.items, dig)
			evicted++
		}
	}
	return evicted
}

// Size returns count of tracked entries.
func (d *SlidingWindowDeduplicator) Size() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.items)
}
