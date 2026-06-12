package events

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"
)

// DeduplicationWindow enforces idempotency over a sliding time duration.
type DeduplicationWindow struct {
	mu       sync.RWMutex
	seenKeys map[string]time.Time
	window   time.Duration
}

// NewDeduplicationWindow instantiates a deduplication filter with a retention TTL.
func NewDeduplicationWindow(ttl time.Duration) *DeduplicationWindow {
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	return &DeduplicationWindow{
		seenKeys: make(map[string]time.Time),
		window:   ttl,
	}
}

// IsDuplicate checks and marks a key. Returns true if key was already seen within window.
func (d *DeduplicationWindow) IsDuplicate(payload []byte) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	hash := sha256.Sum256(payload)
	key := hex.EncodeToString(hash[:])
	now := time.Now()

	if expireAt, exists := d.seenKeys[key]; exists {
		if now.Before(expireAt) {
			return true
		}
	}

	d.seenKeys[key] = now.Add(d.window)
	return false
}

// PruneExpired evicts outdated keys from the tracking cache.
func (d *DeduplicationWindow) PruneExpired(now time.Time) int {
	d.mu.Lock()
	defer d.mu.Unlock()

	evicted := 0
	for k, exp := range d.seenKeys {
		if now.After(exp) {
			delete(d.seenKeys, k)
			evicted++
		}
	}
	return evicted
}

// Size returns total tracked keys currently held.
func (d *DeduplicationWindow) Size() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.seenKeys)
}
