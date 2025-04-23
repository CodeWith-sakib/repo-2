package events

import (
	"sync"
	"time"
)

type Deduplicator struct {
	mu      sync.RWMutex
	seen    map[string]time.Time
	ttl     time.Duration
	stopCh  chan struct{}
}

func NewDeduplicator(ttl time.Duration, cleanupInterval time.Duration) *Deduplicator {
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	if cleanupInterval <= 0 {
		cleanupInterval = 1 * time.Minute
	}
	d := &Deduplicator{
		seen:   make(map[string]time.Time),
		ttl:    ttl,
		stopCh: make(chan struct{}),
	}
	go d.cleanupLoop(cleanupInterval)
	return d
}

// IsDuplicate checks if an event ID has been processed recently. If not, it records it.
func (d *Deduplicator) IsDuplicate(eventID string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now()
	exp, exists := d.seen[eventID]
	if exists && now.Before(exp) {
		return true
	}

	d.seen[eventID] = now.Add(d.ttl)
	return false
}

func (d *Deduplicator) Close() {
	close(d.stopCh)
}

func (d *Deduplicator) cleanupLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-d.stopCh:
			return
		case now := <-ticker.C:
			d.mu.Lock()
			for k, exp := range d.seen {
				if now.After(exp) {
					delete(d.seen, k)
				}
			}
			d.mu.Unlock()
		}
	}
}
