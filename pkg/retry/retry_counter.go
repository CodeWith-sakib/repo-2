package retry

import (
	"sync"
)

type RetryCounter struct {
	mu      sync.Mutex
	counts  map[string]int
}

func NewRetryCounter() *RetryCounter {
	return &RetryCounter{
		counts: make(map[string]int),
	}
}

func (c *RetryCounter) Inc(key string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.counts[key]++
	return c.counts[key]
}

func (c *RetryCounter) Get(key string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.counts[key]
}

func (c *RetryCounter) Reset(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.counts, key)
}
