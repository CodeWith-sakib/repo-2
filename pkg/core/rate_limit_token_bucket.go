package core

import (
	"sync"
	"time"
)

// TokenBucketLimiter implements a thread-safe token bucket rate limiter.
type TokenBucketLimiter struct {
	mu         sync.Mutex
	rate       float64 // tokens per second
	capacity   float64 // maximum burst size
	tokens     float64 // current available tokens
	lastRefill time.Time
}

// NewTokenBucketLimiter constructs a rate limiter with rate (tokens/sec) and burst capacity.
func NewTokenBucketLimiter(rate float64, capacity float64) *TokenBucketLimiter {
	return &TokenBucketLimiter{
		rate:       rate,
		capacity:   capacity,
		tokens:     capacity,
		lastRefill: time.Now().UTC(),
	}
}

// Allow attempts to consume 1 token. Returns true if allowed, false if rate limited.
func (tbl *TokenBucketLimiter) Allow() bool {
	return tbl.AllowN(time.Now().UTC(), 1.0)
}

// AllowN attempts to consume n tokens at specified timestamp.
func (tbl *TokenBucketLimiter) AllowN(now time.Time, n float64) bool {
	tbl.mu.Lock()
	defer tbl.mu.Unlock()

	elapsed := now.Sub(tbl.lastRefill).Seconds()
	if elapsed > 0 {
		tbl.tokens = tbl.tokens + elapsed*tbl.rate
		if tbl.tokens > tbl.capacity {
			tbl.tokens = tbl.capacity
		}
		tbl.lastRefill = now
	}

	if tbl.tokens >= n {
		tbl.tokens -= n
		return true
	}

	return false
}

// AvailableTokens returns currently available tokens without consuming.
func (tbl *TokenBucketLimiter) AvailableTokens() float64 {
	tbl.mu.Lock()
	defer tbl.mu.Unlock()

	elapsed := time.Now().UTC().Sub(tbl.lastRefill).Seconds()
	if elapsed > 0 {
		t := tbl.tokens + elapsed*tbl.rate
		if t > tbl.capacity {
			return tbl.capacity
		}
		return t
	}
	return tbl.tokens
}
