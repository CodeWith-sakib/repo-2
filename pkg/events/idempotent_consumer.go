package events

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

// DedupWindowEntry tracks timestamp of a processed event.
type DedupWindowEntry struct {
	ProcessedAt time.Time
}

// ConsumerMetrics tracks idempotency filter metrics.
type ConsumerMetrics struct {
	TotalReceived     atomic.Int64
	Processed         atomic.Int64
	DuplicatesSkipped atomic.Int64
	ProcessingErrors  atomic.Int64
}

// IdempotentEventConsumer wraps an event handler and guarantees exactly-once processing per event ID.
type IdempotentEventConsumer struct {
	mu      sync.RWMutex
	seen    map[string]DedupWindowEntry
	ttl     time.Duration
	handler EventHandler
	metrics ConsumerMetrics
}

// NewIdempotentEventConsumer creates a consumer with sliding TTL deduplication window.
func NewIdempotentEventConsumer(handler EventHandler, ttl time.Duration) *IdempotentEventConsumer {
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	return &IdempotentEventConsumer{
		seen:    make(map[string]DedupWindowEntry),
		ttl:     ttl,
		handler: handler,
	}
}

// HandleEvent intercepts an event, drops it if duplicate, or invokes inner handler.
func (c *IdempotentEventConsumer) HandleEvent(ctx context.Context, event *core.Event) error {
	c.metrics.TotalReceived.Add(1)

	if event == nil || string(event.ID) == "" {
		return fmt.Errorf("event or event ID cannot be empty")
	}

	key := string(event.ID)
	now := time.Now()

	c.mu.Lock()
	entry, exists := c.seen[key]
	if exists {
		if now.Sub(entry.ProcessedAt) < c.ttl {
			// Duplicate! Skip processing
			c.mu.Unlock()
			c.metrics.DuplicatesSkipped.Add(1)
			return nil
		}
		// TTL expired, allow reprocessing
	}

	// Mark seen tentatively
	c.seen[key] = DedupWindowEntry{ProcessedAt: now}
	c.mu.Unlock()

	// Execute user handler
	if err := c.handler(ctx, event); err != nil {
		c.metrics.ProcessingErrors.Add(1)
		// On error, remove key from seen so it can be retried
		c.mu.Lock()
		delete(c.seen, key)
		c.mu.Unlock()
		return err
	}

	c.metrics.Processed.Add(1)
	return nil
}

// PruneExpired purges keys older than the TTL.
func (c *IdempotentEventConsumer) PruneExpired() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	pruned := 0
	for k, v := range c.seen {
		if now.Sub(v.ProcessedAt) >= c.ttl {
			delete(c.seen, k)
			pruned++
		}
	}
	return pruned
}

// Metrics returns telemetry counters.
func (c *IdempotentEventConsumer) Metrics() (total, processed, dups, errs int64) {
	return c.metrics.TotalReceived.Load(),
		c.metrics.Processed.Load(),
		c.metrics.DuplicatesSkipped.Load(),
		c.metrics.ProcessingErrors.Load()
}
