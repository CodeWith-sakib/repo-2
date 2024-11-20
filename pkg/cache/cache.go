package cache

import (
	"container/list"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/events"
)

type cacheEntry struct {
	key       string
	workflow  *core.WorkflowDefinition
	expiresAt time.Time
	element   *list.Element
}

type DefinitionCache struct {
	mu         sync.RWMutex
	capacity   int
	ttl        time.Duration
	items      map[string]*cacheEntry
	evictList  *list.List
	hits       int64
	misses     int64
}

func NewDefinitionCache(capacity int, ttl time.Duration) *DefinitionCache {
	if capacity <= 0 {
		capacity = 256
	}
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	return &DefinitionCache{
		capacity:  capacity,
		ttl:       ttl,
		items:     make(map[string]*cacheEntry),
		evictList: list.New(),
	}
}

func cacheKey(id core.ID, version int) string {
	return fmt.Sprintf("%s:v%d", id, version)
}

func (c *DefinitionCache) Get(id core.ID, version int) (*core.WorkflowDefinition, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := cacheKey(id, version)
	entry, exists := c.items[key]
	if !exists {
		c.misses++
		return nil, false
	}

	if time.Now().After(entry.expiresAt) {
		c.removeEntry(entry)
		c.misses++
		return nil, false
	}

	c.evictList.MoveToFront(entry.element)
	c.hits++
	return entry.workflow, true
}

func (c *DefinitionCache) Put(wf *core.WorkflowDefinition) {
	if wf == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	key := cacheKey(wf.ID, wf.Version)
	now := time.Now()

	if entry, exists := c.items[key]; exists {
		entry.workflow = wf
		entry.expiresAt = now.Add(c.ttl)
		c.evictList.MoveToFront(entry.element)
		return
	}

	if c.evictList.Len() >= c.capacity {
		c.evictOldest()
	}

	elem := c.evictList.PushFront(key)
	c.items[key] = &cacheEntry{
		key:       key,
		workflow:  wf,
		expiresAt: now.Add(c.ttl),
		element:   elem,
	}
}

func (c *DefinitionCache) Invalidate(id core.ID) {
	c.mu.Lock()
	defer c.mu.Unlock()

	prefix := fmt.Sprintf("%s:", id)
	for key, entry := range c.items {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			c.removeEntry(entry)
		}
	}
}

func (c *DefinitionCache) InvalidateVersion(id core.ID, version int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := cacheKey(id, version)
	if entry, exists := c.items[key]; exists {
		c.removeEntry(entry)
	}
}

func (c *DefinitionCache) Purge() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*cacheEntry)
	c.evictList.Init()
}

func (c *DefinitionCache) removeEntry(entry *cacheEntry) {
	c.evictList.Remove(entry.element)
	delete(c.items, entry.key)
}

func (c *DefinitionCache) evictOldest() {
	elem := c.evictList.Back()
	if elem != nil {
		key := elem.Value.(string)
		if entry, exists := c.items[key]; exists {
			c.removeEntry(entry)
		}
	}
}

func (c *DefinitionCache) AttachEventBus(bus *events.Bus) {
	bus.Subscribe(core.EventWorkflowUpdated, func(ctx context.Context, e *core.Event) error {
		c.Invalidate(e.RunID)
		return nil
	})
	bus.Subscribe(core.EventWorkflowDeleted, func(ctx context.Context, e *core.Event) error {
		c.Invalidate(e.RunID)
		return nil
	})
}
