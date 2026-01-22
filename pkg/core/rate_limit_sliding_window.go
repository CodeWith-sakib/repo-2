package core

import (
	"sync"
	"time"
)

// SlidingWindowLimiter implements a high-precision sliding window log rate limiter.
type SlidingWindowLimiter struct {
	mu         sync.Mutex
	limit      int
	windowSize time.Duration
	timestamps []time.Time
}

// NewSlidingWindowLimiter creates a sliding window limiter.
func NewSlidingWindowLimiter(limit int, windowSize time.Duration) *SlidingWindowLimiter {
	if limit <= 0 {
		limit = 100
	}
	if windowSize <= 0 {
		windowSize = time.Minute
	}
	return &SlidingWindowLimiter{
		limit:      limit,
		windowSize: windowSize,
		timestamps: make([]time.Time, 0, limit),
	}
}

// Allow checks if another request is permitted within the sliding window at current time.
func (l *SlidingWindowLimiter) Allow() bool {
	return l.AllowAt(time.Now().UTC())
}

// AllowAt checks if request is permitted at given timestamp, sliding the window boundary.
func (l *SlidingWindowLimiter) AllowAt(now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	cutoff := now.Add(-l.windowSize)

	// Prune timestamps older than cutoff
	validIdx := 0
	for i, t := range l.timestamps {
		if t.After(cutoff) {
			validIdx = i
			break
		}
		if i == len(l.timestamps)-1 {
			validIdx = len(l.timestamps)
		}
	}
	if validIdx > 0 {
		l.timestamps = l.timestamps[validIdx:]
	}

	if len(l.timestamps) < l.limit {
		l.timestamps = append(l.timestamps, now)
		return true
	}

	return false
}

// CurrentCount returns number of active requests in the sliding window.
func (l *SlidingWindowLimiter) CurrentCount(now time.Time) int {
	l.mu.Lock()
	defer l.mu.Unlock()

	cutoff := now.Add(-l.windowSize)
	count := 0
	for _, t := range l.timestamps {
		if t.After(cutoff) {
			count++
		}
	}
	return count
}
