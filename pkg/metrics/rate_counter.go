package metrics

import (
	"sync"
	"sync/atomic"
	"time"
)

// RateWindow is a fixed-duration sliding window for rate computation.
type RateWindow struct {
	mu         sync.Mutex
	buckets    []uint64
	numBuckets int
	bucketDur  time.Duration
	total      atomic.Int64
	startTime  time.Time
}

// NewRateWindow creates a sliding window rate counter divided into numBuckets time slots.
func NewRateWindow(window time.Duration, numBuckets int) *RateWindow {
	if numBuckets < 1 {
		numBuckets = 1
	}
	return &RateWindow{
		buckets:    make([]uint64, numBuckets),
		numBuckets: numBuckets,
		bucketDur:  window / time.Duration(numBuckets),
		startTime:  time.Now(),
	}
}

// Increment records n events at the current instant.
func (r *RateWindow) Increment(n uint64) {
	r.mu.Lock()
	idx := r.currentBucketIndex()
	r.buckets[idx] += n
	r.mu.Unlock()
	r.total.Add(int64(n))
}

// RatePerSecond computes the average events-per-second over the configured window.
func (r *RateWindow) RatePerSecond() float64 {
	r.mu.Lock()
	defer r.mu.Unlock()

	total := uint64(0)
	for _, b := range r.buckets {
		total += b
	}

	windowSeconds := float64(r.numBuckets) * r.bucketDur.Seconds()
	if windowSeconds == 0 {
		return 0
	}
	return float64(total) / windowSeconds
}

// Total returns the lifetime event count.
func (r *RateWindow) Total() int64 {
	return r.total.Load()
}

// Reset zeroes out all window buckets (but not the lifetime total).
func (r *RateWindow) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.buckets {
		r.buckets[i] = 0
	}
}

func (r *RateWindow) currentBucketIndex() int {
	elapsed := time.Since(r.startTime)
	bucketIndex := int(elapsed/r.bucketDur) % r.numBuckets
	return bucketIndex
}

// MultiDimensionalCounter tracks multiple labeled event counters atomically.
type MultiDimensionalCounter struct {
	mu       sync.RWMutex
	counters map[string]*atomic.Int64
}

// NewMultiDimensionalCounter creates an empty multi-label event counter.
func NewMultiDimensionalCounter() *MultiDimensionalCounter {
	return &MultiDimensionalCounter{
		counters: make(map[string]*atomic.Int64),
	}
}

// Inc increments a labeled counter by 1.
func (m *MultiDimensionalCounter) Inc(label string) {
	m.mu.RLock()
	c, ok := m.counters[label]
	m.mu.RUnlock()

	if !ok {
		m.mu.Lock()
		if c2, ok2 := m.counters[label]; ok2 {
			c = c2
		} else {
			c = &atomic.Int64{}
			m.counters[label] = c
		}
		m.mu.Unlock()
	}

	c.Add(1)
}

// Add adds delta to a labeled counter.
func (m *MultiDimensionalCounter) Add(label string, delta int64) {
	m.mu.RLock()
	c, ok := m.counters[label]
	m.mu.RUnlock()

	if !ok {
		m.mu.Lock()
		if c2, ok2 := m.counters[label]; ok2 {
			c = c2
		} else {
			c = &atomic.Int64{}
			m.counters[label] = c
		}
		m.mu.Unlock()
	}

	c.Add(delta)
}

// Get returns the count for a label.
func (m *MultiDimensionalCounter) Get(label string) int64 {
	m.mu.RLock()
	c, ok := m.counters[label]
	m.mu.RUnlock()
	if !ok {
		return 0
	}
	return c.Load()
}

// Snapshot returns a copy of all label counts.
func (m *MultiDimensionalCounter) Snapshot() map[string]int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make(map[string]int64, len(m.counters))
	for k, v := range m.counters {
		out[k] = v.Load()
	}
	return out
}
