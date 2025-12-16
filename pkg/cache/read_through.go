package cache

import (
	"context"
	"errors"
	"sync"
	"time"
)

// CacheLoader defines loader function signature when cache misses occur.
type CacheLoader func(ctx context.Context, key string) (interface{}, error)

// ReadThroughCache provides thread-safe singleflight stampede suppression over standard SegmentedLRU cache.
type ReadThroughCache struct {
	cache       *SegmentedLRU
	loader      CacheLoader
	mu          sync.Mutex
	inFlight    map[string]*call
	defaultTTL  time.Duration
}

type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

// NewReadThroughCache constructs a read-through decorator wrapping a SegmentedLRU.
func NewReadThroughCache(backend *SegmentedLRU, ttl time.Duration, loader CacheLoader) *ReadThroughCache {
	return &ReadThroughCache{
		cache:      backend,
		loader:     loader,
		inFlight:   make(map[string]*call),
		defaultTTL: ttl,
	}
}

// GetOrLoad checks backend cache; if absent, triggers singleflight loader and caches result.
func (rtc *ReadThroughCache) GetOrLoad(ctx context.Context, key string) (interface{}, error) {
	if rtc.cache == nil {
		return nil, errors.New("backend cache is nil")
	}

	// Fast path: cache hit
	if val, ok := rtc.cache.Get(key); ok {
		return val, nil
	}

	if rtc.loader == nil {
		return nil, errors.New("no loader defined for cache miss")
	}

	rtc.mu.Lock()
	if c, exists := rtc.inFlight[key]; exists {
		rtc.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}

	c := new(call)
	c.wg.Add(1)
	rtc.inFlight[key] = c
	rtc.mu.Unlock()

	// Execute loader once
	c.val, c.err = rtc.loader(ctx, key)
	if c.err == nil && c.val != nil {
		rtc.cache.Set(key, c.val)
	}
	c.wg.Done()

	rtc.mu.Lock()
	delete(rtc.inFlight, key)
	rtc.mu.Unlock()

	return c.val, c.err
}
