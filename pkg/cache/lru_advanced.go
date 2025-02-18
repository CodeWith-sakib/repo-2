package cache

import (
	"sync"
	"time"
)

type CacheItem2Q struct {
	Key       string
	Value     interface{}
	ExpiresAt time.Time
	Hits      int
	inA1in    bool
	inA1out   bool
	inAm      bool
}

type TwoQueueCache struct {
	mu        sync.RWMutex
	capacity  int
	items     map[string]*CacheItem2Q
	a1in      []string
	a1out     []string
	am        []string
	hits      int64
	misses    int64
}

func NewTwoQueueCache(capacity int) *TwoQueueCache {
	if capacity < 4 {
		capacity = 4
	}
	return &TwoQueueCache{
		capacity: capacity,
		items:    make(map[string]*CacheItem2Q),
		a1in:     make([]string, 0),
		a1out:    make([]string, 0),
		am:       make([]string, 0),
	}
}

func (c *TwoQueueCache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, exists := c.items[key]
	if !exists {
		c.misses++
		return nil, false
	}

	if !item.ExpiresAt.IsZero() && time.Now().After(item.ExpiresAt) {
		c.removeItem(key)
		c.misses++
		return nil, false
	}

	item.Hits++
	c.hits++

	// Promote from A1in to Am if re-referenced
	if item.inA1in {
		item.inA1in = false
		item.inAm = true
		c.am = append(c.am, key)
	}

	return item.Value, true
}

func (c *TwoQueueCache) Set(key string, val interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var exp time.Time
	if ttl > 0 {
		exp = time.Now().Add(ttl)
	}

	if item, exists := c.items[key]; exists {
		item.Value = val
		item.ExpiresAt = exp
		return
	}

	if len(c.items) >= c.capacity {
		c.evict()
	}

	item := &CacheItem2Q{
		Key:       key,
		Value:     val,
		ExpiresAt: exp,
		inA1in:    true,
	}
	c.items[key] = item
	c.a1in = append(c.a1in, key)
}

func (c *TwoQueueCache) evict() {
	if len(c.a1in) > 0 {
		evictKey := c.a1in[0]
		c.a1in = c.a1in[1:]
		c.removeItem(evictKey)
		return
	}
	if len(c.am) > 0 {
		evictKey := c.am[0]
		c.am = c.am[1:]
		c.removeItem(evictKey)
		return
	}
}

func (c *TwoQueueCache) removeItem(key string) {
	delete(c.items, key)
}

func (c *TwoQueueCache) Stats() (int64, int64, int) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.hits, c.misses, len(c.items)
}
