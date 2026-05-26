package postgres

import (
	"sort"
	"sync"
	"time"
)

// QueryLatencyHistogram provides quantile percentile calculations (P50, P90, P99) for SQL latency monitoring.
type QueryLatencyHistogram struct {
	mu      sync.RWMutex
	samples []time.Duration
	maxSize int
}

// NewQueryLatencyHistogram creates a latency histogram sampler.
func NewQueryLatencyHistogram(maxSize int) *QueryLatencyHistogram {
	if maxSize <= 0 {
		maxSize = 1000
	}
	return &QueryLatencyHistogram{
		samples: make([]time.Duration, 0, maxSize),
		maxSize: maxSize,
	}
}

// RecordSample registers an execution latency duration.
func (h *QueryLatencyHistogram) RecordSample(d time.Duration) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.samples) >= h.maxSize {
		h.samples = h.samples[1:]
	}
	h.samples = append(h.samples, d)
}

// Percentile calculates the estimated duration at target percentile (0.0 to 1.0).
func (h *QueryLatencyHistogram) Percentile(p float64) time.Duration {
	h.mu.RLock()
	defer h.mu.RUnlock()

	n := len(h.samples)
	if n == 0 {
		return 0
	}

	sorted := make([]time.Duration, n)
	copy(sorted, h.samples)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	idx := int(float64(n-1) * p)
	if idx < 0 {
		idx = 0
	} else if idx >= n {
		idx = n - 1
	}

	return sorted[idx]
}

// SampleCount returns total number of currently buffered latency samples.
func (h *QueryLatencyHistogram) SampleCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.samples)
}
