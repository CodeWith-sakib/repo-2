package cache

import (
	"container/list"
	"fmt"
	"hash/fnv"
	"sync"
)

// lruEntry is the value stored in the doubly-linked list.
type lruEntry struct {
	key   string
	value interface{}
}

// lruShard is a single LRU shard protected by its own mutex.
type lruShard struct {
	mu       sync.Mutex
	capacity int
	items    map[string]*list.Element
	order    *list.List
}

func newLRUShard(capacity int) *lruShard {
	return &lruShard{
		capacity: capacity,
		items:    make(map[string]*list.Element),
		order:    list.New(),
	}
}

func (s *lruShard) Get(key string) (interface{}, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	elem, ok := s.items[key]
	if !ok {
		return nil, false
	}
	s.order.MoveToFront(elem)
	return elem.Value.(*lruEntry).value, true
}

func (s *lruShard) Set(key string, val interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if elem, ok := s.items[key]; ok {
		s.order.MoveToFront(elem)
		elem.Value.(*lruEntry).value = val
		return
	}
	if s.order.Len() >= s.capacity {
		oldest := s.order.Back()
		if oldest != nil {
			s.order.Remove(oldest)
			delete(s.items, oldest.Value.(*lruEntry).key)
		}
	}
	elem := s.order.PushFront(&lruEntry{key: key, value: val})
	s.items[key] = elem
}

func (s *lruShard) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if elem, ok := s.items[key]; ok {
		s.order.Remove(elem)
		delete(s.items, key)
	}
}

func (s *lruShard) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.order.Len()
}

// ShardedLRUCache is a high-concurrency LRU cache with N independent shards,
// each with its own lock and eviction list, reducing mutex contention at scale.
type ShardedLRUCache struct {
	shards    []*lruShard
	shardMask uint32
}

// NewShardedLRUCache creates a sharded LRU cache.
// numShards must be a power of two. totalCapacity is divided across shards.
func NewShardedLRUCache(numShards, totalCapacity int) (*ShardedLRUCache, error) {
	if numShards <= 0 || numShards&(numShards-1) != 0 {
		return nil, fmt.Errorf("numShards must be a positive power of two, got %d", numShards)
	}
	perShard := totalCapacity / numShards
	if perShard < 1 {
		perShard = 1
	}
	shards := make([]*lruShard, numShards)
	for i := range shards {
		shards[i] = newLRUShard(perShard)
	}
	return &ShardedLRUCache{
		shards:    shards,
		shardMask: uint32(numShards - 1),
	}, nil
}

func (c *ShardedLRUCache) shardFor(key string) *lruShard {
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	return c.shards[h.Sum32()&c.shardMask]
}

// Get retrieves a value by key.
func (c *ShardedLRUCache) Get(key string) (interface{}, bool) {
	return c.shardFor(key).Get(key)
}

// Set inserts or updates a value. Evicts LRU entries within the shard when full.
func (c *ShardedLRUCache) Set(key string, val interface{}) {
	c.shardFor(key).Set(key, val)
}

// Delete removes a key from the cache.
func (c *ShardedLRUCache) Delete(key string) {
	c.shardFor(key).Delete(key)
}

// Len returns the total number of cached entries across all shards.
func (c *ShardedLRUCache) Len() int {
	total := 0
	for _, s := range c.shards {
		total += s.Len()
	}
	return total
}
