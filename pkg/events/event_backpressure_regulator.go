package events

import (
	"sync"
	"time"
)

// EventBackpressureRegulator regulates producer publishing velocity based on consumer lag.
type EventBackpressureRegulator struct {
	mu           sync.RWMutex
	maxQueueDepth int
	currentDepth int
	throttleSleep time.Duration
}

// NewEventBackpressureRegulator creates a backpressure regulator.
func NewEventBackpressureRegulator(maxQueueDepth int, maxSleep time.Duration) *EventBackpressureRegulator {
	if maxQueueDepth <= 0 {
		maxQueueDepth = 10000
	}
	if maxSleep <= 0 {
		maxSleep = 100 * time.Millisecond
	}
	return &EventBackpressureRegulator{
		maxQueueDepth: maxQueueDepth,
		throttleSleep: maxSleep,
	}
}

// UpdateQueueDepth updates current backlog metric.
func (r *EventBackpressureRegulator) UpdateQueueDepth(depth int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.currentDepth = depth
}

// ComputeThrottleDelay derives producer sleep duration to apply backpressure.
func (r *EventBackpressureRegulator) ComputeThrottleDelay() time.Duration {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.currentDepth <= r.maxQueueDepth/2 {
		return 0 // Healthy buffer, no throttling
	}

	ratio := float64(r.currentDepth) / float64(r.maxQueueDepth)
	if ratio > 1.0 {
		ratio = 1.0
	}

	return time.Duration(float64(r.throttleSleep) * ratio)
}

// IsOverloaded checks if consumer backlog has exceeded maximum queue depth.
func (r *EventBackpressureRegulator) IsOverloaded() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.currentDepth >= r.maxQueueDepth
}
