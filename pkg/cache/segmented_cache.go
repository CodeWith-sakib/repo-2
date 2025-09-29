package cache

import (
	"container/list"
	"fmt"
	"sync"
	"sync/atomic"
)

// slruNode stores the entry in the linked list.
type slruNode struct {
	key   string
	value interface{}
}

// SLRUMetrics tracks operational statistics of the Segmented LRU cache.
type SLRUMetrics struct {
	Hits       atomic.Int64
	Misses     atomic.Int64
	Promotions atomic.Int64
	Demotions  atomic.Int64
	Evictions  atomic.Int64
}

// SegmentedLRU implements a 2-stage Segmented LRU cache with Probationary and Protected tiers.
type SegmentedLRU struct {
	mu           sync.RWMutex
	probationCap int
	protectedCap int

	probationList *list.List
	probationMap  map[string]*list.Element

	protectedList *list.List
	protectedMap  map[string]*list.Element

	metrics SLRUMetrics
}

// NewSegmentedLRU initializes an SLRU cache with given capacities for probation and protected tiers.
func NewSegmentedLRU(probationCap, protectedCap int) (*SegmentedLRU, error) {
	if probationCap <= 0 || protectedCap <= 0 {
		return nil, fmt.Errorf("capacities must be positive: probation=%d protected=%d", probationCap, protectedCap)
	}
	return &SegmentedLRU{
		probationCap:  probationCap,
		protectedCap:  protectedCap,
		probationList: list.New(),
		probationMap:  make(map[string]*list.Element),
		protectedList: list.New(),
		protectedMap:  make(map[string]*list.Element),
	}, nil
}

// Get looks up a key. If found in probation, it is promoted to protected.
func (c *SegmentedLRU) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 1. Check protected tier
	if elem, ok := c.protectedMap[key]; ok {
		c.protectedList.MoveToFront(elem)
		c.metrics.Hits.Add(1)
		return elem.Value.(*slruNode).value, true
	}

	// 2. Check probationary tier
	if elem, ok := c.probationMap[key]; ok {
		node := elem.Value.(*slruNode)
		// Hit in probation -> promote to protected!
		c.probationList.Remove(elem)
		delete(c.probationMap, key)

		// Make room in protected if full -> demote oldest to probation
		if c.protectedList.Len() >= c.protectedCap {
			c.demoteOldestProtected()
		}

		newElem := c.protectedList.PushFront(node)
		c.protectedMap[key] = newElem

		c.metrics.Hits.Add(1)
		c.metrics.Promotions.Add(1)
		return node.value, true
	}

	c.metrics.Misses.Add(1)
	return nil, false
}

// Set puts a key/value. If it exists, it updates and moves to front. If new, placed in probationary tier.
func (c *SegmentedLRU) Set(key string, val interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if already in protected
	if elem, ok := c.protectedMap[key]; ok {
		elem.Value.(*slruNode).value = val
		c.protectedList.MoveToFront(elem)
		return
	}

	// Check if already in probation
	if elem, ok := c.probationMap[key]; ok {
		elem.Value.(*slruNode).value = val
		// Update also promotes to protected
		c.probationList.Remove(elem)
		delete(c.probationMap, key)

		if c.protectedList.Len() >= c.protectedCap {
			c.demoteOldestProtected()
		}
		newElem := c.protectedList.PushFront(&slruNode{key: key, value: val})
		c.protectedMap[key] = newElem
		c.metrics.Promotions.Add(1)
		return
	}

	// New entry: add to probation
	if c.probationList.Len() >= c.probationCap {
		// Evict oldest from probation
		oldest := c.probationList.Back()
		if oldest != nil {
			c.probationList.Remove(oldest)
			delete(c.probationMap, oldest.Value.(*slruNode).key)
			c.metrics.Evictions.Add(1)
		}
	}

	elem := c.probationList.PushFront(&slruNode{key: key, value: val})
	c.probationMap[key] = elem
}

func (c *SegmentedLRU) demoteOldestProtected() {
	oldest := c.protectedList.Back()
	if oldest == nil {
		return
	}
	c.protectedList.Remove(oldest)
	delete(c.protectedMap, oldest.Value.(*slruNode).key)
	c.metrics.Demotions.Add(1)

	// Add to probation
	if c.probationList.Len() >= c.probationCap {
		probOldest := c.probationList.Back()
		if probOldest != nil {
			c.probationList.Remove(probOldest)
			delete(c.probationMap, probOldest.Value.(*slruNode).key)
			c.metrics.Evictions.Add(1)
		}
	}
	elem := c.probationList.PushFront(oldest.Value)
	c.probationMap[oldest.Value.(*slruNode).key] = elem
}

// Len returns the total number of items stored across both tiers.
func (c *SegmentedLRU) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.probationList.Len() + c.protectedList.Len()
}

// Metrics returns current counters.
func (c *SegmentedLRU) Metrics() (hits, misses, promotions, demotions, evictions int64) {
	return c.metrics.Hits.Load(),
		c.metrics.Misses.Load(),
		c.metrics.Promotions.Load(),
		c.metrics.Demotions.Load(),
		c.metrics.Evictions.Load()
}
