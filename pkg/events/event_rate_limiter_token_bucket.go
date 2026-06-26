package events

import (
	"sync"
	"time"
)

// EventTokenBucketLimiter throttles event ingestion rate per partition or tenant.
type EventTokenBucketLimiter struct {
	mu           sync.Mutex
	capacity     float64
	tokens       float64
	refillRate   float64 // tokens per second
	lastRefillAt time.Time
}

// NewEventTokenBucketLimiter creates a rate limiter with given capacity and refill rate per second.
func NewEventTokenBucketLimiter(capacity, refillPerSec float64) *EventTokenBucketLimiter {
	if capacity <= 0 {
		capacity = 100
	}
	if refillPerSec <= 0 {
		refillPerSec = 50
	}
	return &EventTokenBucketLimiter{
		capacity:     capacity,
		tokens:       capacity,
		refillRate:   refillPerSec,
		lastRefillAt: time.Now(),
	}
}

// Allow attempts to consume N tokens from the bucket.
func (l *EventTokenBucketLimiter) Allow(tokens float64) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(l.lastRefillAt).Seconds()
	l.tokens += elapsed * l.refillRate
	if l.tokens > l.capacity {
		l.tokens = l.capacity
	}
	l.lastRefillAt = now

	if l.tokens >= tokens {
		l.tokens -= tokens
		return true
	}
	return false
}

// AvailableTokens returns current number of tokens in the bucket.
func (l *EventTokenBucketLimiter) AvailableTokens() float64 {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(l.lastRefillAt).Seconds()
	tokens := l.tokens + elapsed*l.refillRate
	if tokens > l.capacity {
		tokens = l.capacity
	}
	return tokens
}
