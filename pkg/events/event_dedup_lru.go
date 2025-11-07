package events

import (
	"container/list"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type lruDedupEntry struct {
	key       string
	expiresAt time.Time
}

// LRUDedupMetrics tracks deduplication cache statistics.
type LRUDedupMetrics struct {
	TotalChecked       atomic.Int64
	DuplicatesDetected atomic.Int64
	EvictedDueToCap    atomic.Int64
	EvictedDueToTTL    atomic.Int64
}

// LRUEventDeduplicator provides a bounded, memory-safe sliding window for event deduplication.
type LRUEventDeduplicator struct {
	mu       sync.RWMutex
	capacity int
	ttl      time.Duration
	items    map[string]*list.Element
	evictSeq *list.List
	metrics  LRUDedupMetrics
}

// NewLRUEventDeduplicator initializes an LRU deduplicator.
func NewLRUEventDeduplicator(capacity int, ttl time.Duration) (*LRUEventDeduplicator, error) {
	if capacity <= 0 || ttl <= 0 {
		return nil, fmt.Errorf("capacity and ttl must be positive: cap=%d, ttl=%v", capacity, ttl)
	}
	return &LRUEventDeduplicator{
		capacity: capacity,
		ttl:      ttl,
		items:    make(map[string]*list.Element),
		evictSeq: list.New(),
	}, nil
}

// CheckAndRecord tests if an event key was already seen. Returns true if duplicate, false if new.
func (d *LRUEventDeduplicator) CheckAndRecord(key string) bool {
	d.metrics.TotalChecked.Add(1)

	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now()

	// Check if present
	if elem, exists := d.items[key]; exists {
		entry := elem.Value.(*lruDedupEntry)
		if now.Before(entry.expiresAt) {
			// Valid duplicate! Move to front
			d.evictSeq.MoveToFront(elem)
			d.metrics.DuplicatesDetected.Add(1)
			return true
		}
		// Expired entry, refresh it
		entry.expiresAt = now.Add(d.ttl)
		d.evictSeq.MoveToFront(elem)
		return false
	}

	// New entry
	if d.evictSeq.Len() >= d.capacity {
		// Evict oldest from back
		oldest := d.evictSeq.Back()
		if oldest != nil {
			d.evictSeq.Remove(oldest)
			delete(d.items, oldest.Value.(*lruDedupEntry).key)
			d.metrics.EvictedDueToCap.Add(1)
		}
	}

	entry := &lruDedupEntry{
		key:       key,
		expiresAt: now.Add(d.ttl),
	}
	elem := d.evictSeq.PushFront(entry)
	d.items[key] = elem

	return false
}

// Len returns the current number of tracked keys.
func (d *LRUEventDeduplicator) Len() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.items)
}

// Metrics returns telemetry counters.
func (d *LRUEventDeduplicator) Metrics() (total, dups, evictedCap, evictedTTL int64) {
	return d.metrics.TotalChecked.Load(),
		d.metrics.DuplicatesDetected.Load(),
		d.metrics.EvictedDueToCap.Load(),
		d.metrics.EvictedDueToTTL.Load()
}
