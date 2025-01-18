package metrics

import (
	"math"
	"sort"
	"sync"
	"time"
)

type ExecutionSample struct {
	Duration  time.Duration
	Timestamp time.Time
}

type LatencyHistogram struct {
	mu        sync.RWMutex
	samples   []time.Duration
	maxSample int
}

func NewLatencyHistogram(maxSamples int) *LatencyHistogram {
	if maxSamples <= 0 {
		maxSamples = 10000
	}
	return &LatencyHistogram{
		samples:   make([]time.Duration, 0, maxSamples),
		maxSample: maxSamples,
	}
}

func (h *LatencyHistogram) Record(d time.Duration) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.samples) >= h.maxSample {
		h.samples = h.samples[1:]
	}
	h.samples = append(h.samples, d)
}

func (h *LatencyHistogram) Percentile(p float64) time.Duration {
	h.mu.RLock()
	defer h.mu.RUnlock()

	n := len(h.samples)
	if n == 0 {
		return 0
	}

	sorted := make([]time.Duration, n)
	copy(sorted, h.samples)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

	rank := int(math.Ceil(p/100.0*float64(n))) - 1
	if rank < 0 {
		rank = 0
	}
	if rank >= n {
		rank = n - 1
	}

	return sorted[rank]
}

func (h *LatencyHistogram) Mean() time.Duration {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if len(h.samples) == 0 {
		return 0
	}

	var total int64
	for _, s := range h.samples {
		total += s.Nanoseconds()
	}
	return time.Duration(total / int64(len(h.samples)))
}
