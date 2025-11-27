package worker

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type memoizedEntry struct {
	result    *StepResult
	cachedAt  time.Time
	expiresAt time.Time
}

// TaskMemoizationMetrics records memoization cache telemetry.
type TaskMemoizationMetrics struct {
	Hits   atomic.Int64
	Misses atomic.Int64
	Stores atomic.Int64
}

// TaskMemoizationCache caches deterministic task execution outputs to skip redundant work.
type TaskMemoizationCache struct {
	mu      sync.RWMutex
	cache   map[string]memoizedEntry
	metrics TaskMemoizationMetrics
}

// NewTaskMemoizationCache creates a task memoization cache.
func NewTaskMemoizationCache() *TaskMemoizationCache {
	return &TaskMemoizationCache{
		cache: make(map[string]memoizedEntry),
	}
}

// ComputeKey calculates a cryptographic hash over task type, config, and input payload.
func ComputeKey(taskType string, config, input json.RawMessage) string {
	h := sha256.New()
	h.Write([]byte(taskType))
	h.Write([]byte{0})
	h.Write(config)
	h.Write([]byte{0})
	h.Write(input)
	return hex.EncodeToString(h.Sum(nil))
}

// Get checks for an unexpired cached result.
func (c *TaskMemoizationCache) Get(key string) (*StepResult, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.cache[key]
	if !ok {
		c.metrics.Misses.Add(1)
		return nil, false
	}

	if time.Now().After(entry.expiresAt) {
		c.metrics.Misses.Add(1)
		return nil, false
	}

	c.metrics.Hits.Add(1)
	return entry.result, true
}

// Put stores a step result with a specified TTL.
func (c *TaskMemoizationCache) Put(key string, result *StepResult, ttl time.Duration) error {
	if key == "" || result == nil {
		return fmt.Errorf("invalid key or nil result")
	}
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	c.cache[key] = memoizedEntry{
		result:    result,
		cachedAt:  now,
		expiresAt: now.Add(ttl),
	}
	c.metrics.Stores.Add(1)
	return nil
}

// PruneExpired removes stale cache entries.
func (c *TaskMemoizationCache) PruneExpired() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	pruned := 0
	for k, v := range c.cache {
		if now.After(v.expiresAt) {
			delete(c.cache, k)
			pruned++
		}
	}
	return pruned
}

// Metrics returns cache hit/miss statistics.
func (c *TaskMemoizationCache) Metrics() (hits, misses, stores int64) {
	return c.metrics.Hits.Load(), c.metrics.Misses.Load(), c.metrics.Stores.Load()
}
