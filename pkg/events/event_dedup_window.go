package events

import (
	"fmt"
	"sync"
	"time"
)

type timeBucket struct {
	start time.Time
	keys  map[string]struct{}
}

// RollingBucketDeduplicator maintains sliding time windows using discrete ring-buffered buckets.
type RollingBucketDeduplicator struct {
	mu             sync.RWMutex
	bucketDuration time.Duration
	numBuckets     int
	buckets        []*timeBucket
	currentBucket  int
}

// NewRollingBucketDeduplicator creates a deduplicator with bucketDuration and total buckets to retain.
func NewRollingBucketDeduplicator(bucketDuration time.Duration, numBuckets int) (*RollingBucketDeduplicator, error) {
	if bucketDuration <= 0 || numBuckets <= 0 {
		return nil, fmt.Errorf("bucketDuration and numBuckets must be positive: %v, %d", bucketDuration, numBuckets)
	}

	now := time.Now()
	buckets := make([]*timeBucket, numBuckets)
	for i := 0; i < numBuckets; i++ {
		buckets[i] = &timeBucket{
			start: now.Add(time.Duration(i) * bucketDuration),
			keys:  make(map[string]struct{}),
		}
	}

	return &RollingBucketDeduplicator{
		bucketDuration: bucketDuration,
		numBuckets:     numBuckets,
		buckets:        buckets,
		currentBucket:  0,
	}, nil
}

// CheckAndRecord returns true if key was already seen within the active rolling window, false if new.
func (d *RollingBucketDeduplicator) CheckAndRecord(key string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.advanceBuckets(time.Now())

	// 1. Check if key exists in any active bucket
	for _, b := range d.buckets {
		if _, ok := b.keys[key]; ok {
			return true // duplicate!
		}
	}

	// 2. Add to current bucket
	d.buckets[d.currentBucket].keys[key] = struct{}{}
	return false
}

func (d *RollingBucketDeduplicator) advanceBuckets(now time.Time) {
	curr := d.buckets[d.currentBucket]
	if now.Sub(curr.start) >= d.bucketDuration {
		steps := int(now.Sub(curr.start) / d.bucketDuration)
		if steps > d.numBuckets {
			steps = d.numBuckets
		}

		for i := 0; i < steps; i++ {
			d.currentBucket = (d.currentBucket + 1) % d.numBuckets
			// Reset recycled bucket
			d.buckets[d.currentBucket].start = now
			d.buckets[d.currentBucket].keys = make(map[string]struct{})
		}
	}
}

// TotalTrackedKeys returns total keys currently held across all buckets.
func (d *RollingBucketDeduplicator) TotalTrackedKeys() int {
	d.mu.RLock()
	defer d.mu.RUnlock()

	total := 0
	for _, b := range d.buckets {
		total += len(b.keys)
	}
	return total
}
