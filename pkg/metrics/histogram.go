package metrics

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"time"
)

// BucketBoundary defines an upper bound for an explicit histogram bucket.
type BucketBoundary float64

// HistogramBucket stores observation count for a given upper boundary.
type HistogramBucket struct {
	UpperBound float64
	Count      uint64
}

// ExplicitHistogram is a thread-safe, fixed-bucket histogram with Prometheus-compatible semantics.
type ExplicitHistogram struct {
	mu         sync.RWMutex
	name       string
	labels     map[string]string
	boundaries []float64
	buckets    []uint64
	sum        float64
	count      uint64
	createdAt  time.Time
}

// NewExplicitHistogram creates a histogram with given boundaries (automatically sorted, +Inf added).
func NewExplicitHistogram(name string, boundaries []float64, labels map[string]string) *ExplicitHistogram {
	sorted := make([]float64, len(boundaries))
	copy(sorted, boundaries)
	sort.Float64s(sorted)
	// Append +Inf sentinel
	sorted = append(sorted, math.Inf(1))

	return &ExplicitHistogram{
		name:       name,
		labels:     labels,
		boundaries: sorted,
		buckets:    make([]uint64, len(sorted)),
		createdAt:  time.Now(),
	}
}

// Observe records a single observation value.
func (h *ExplicitHistogram) Observe(value float64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.sum += value
	h.count++

	for i, bound := range h.boundaries {
		if value <= bound {
			h.buckets[i]++
		}
	}
}

// Snapshot returns a copy of the current bucket counts and statistics.
func (h *ExplicitHistogram) Snapshot() ([]HistogramBucket, float64, uint64) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	out := make([]HistogramBucket, len(h.boundaries))
	for i, b := range h.boundaries {
		out[i] = HistogramBucket{UpperBound: b, Count: h.buckets[i]}
	}
	return out, h.sum, h.count
}

// Quantile estimates a quantile (0.0–1.0) via linear interpolation across buckets.
func (h *ExplicitHistogram) Quantile(q float64) float64 {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.count == 0 {
		return 0
	}

	target := q * float64(h.count)
	prev := 0.0
	prevBound := 0.0

	for i, bound := range h.boundaries {
		curr := float64(h.buckets[i])
		if curr >= target {
			// interpolate within this bucket
			fraction := (target - prev) / (curr - prev)
			if curr == prev {
				return bound
			}
			return prevBound + fraction*(bound-prevBound)
		}
		prev = curr
		prevBound = bound
	}
	return prevBound
}

// Mean returns the arithmetic mean of all observations.
func (h *ExplicitHistogram) Mean() float64 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.count == 0 {
		return 0
	}
	return h.sum / float64(h.count)
}

// String returns a text summary for logging.
func (h *ExplicitHistogram) String() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return fmt.Sprintf("histogram{name=%s count=%d sum=%.3f mean=%.3f}", h.name, h.count, h.sum, h.sum/float64(max64(h.count, 1)))
}

func max64(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}
