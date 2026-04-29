package events

import (
	"sync"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

type cachedEventItem struct {
	event     *core.Event
	expiresAt time.Time
}

// EventTTLCache provides short-term caching of recently dispatched events with periodic expiration sweeps.
type EventTTLCache struct {
	mu    sync.RWMutex
	items map[string]cachedEventItem
	ttl   time.Duration
}

// NewEventTTLCache creates an event TTL cache.
func NewEventTTLCache(ttl time.Duration) *EventTTLCache {
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	return &EventTTLCache{
		items: make(map[string]cachedEventItem),
		ttl:   ttl,
	}
}

// Put stores an event with current timestamp + TTL expiration.
func (c *EventTTLCache) Put(event *core.Event) {
	if event == nil || event.ID == "" {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[string(event.ID)] = cachedEventItem{
		event:     event,
		expiresAt: time.Now().UTC().Add(c.ttl),
	}
}

// Get retrieves cached event if present and unexpired.
func (c *EventTTLCache) Get(eventID string) (*core.Event, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[eventID]
	if !exists {
		return nil, false
	}

	if time.Now().UTC().After(item.expiresAt) {
		return nil, false
	}

	return item.event, true
}

// SweepExpired purges all items past their expiration timestamp.
func (c *EventTTLCache) SweepExpired(now time.Time) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	count := 0
	for id, item := range c.items {
		if now.After(item.expiresAt) {
			delete(c.items, id)
			count++
		}
	}
	return count
}
