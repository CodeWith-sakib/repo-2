package postgres

import (
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"
)

// PoolTelemetryProbe instruments a connection pool and exposes derived metrics:
// wait latency percentiles, throughput (conns/sec), and saturation ratio.
type PoolTelemetryProbe struct {
	mu             sync.Mutex
	waitSamples    []time.Duration
	acquiredTotal  atomic.Int64
	failedTotal    atomic.Int64
	activeConns    atomic.Int64
	maxConns       int
	windowStart    time.Time
	windowDuration time.Duration
}

// PoolTelemetrySnapshot is a point-in-time snapshot of pool health metrics.
type PoolTelemetrySnapshot struct {
	ActiveConns      int64
	MaxConns         int
	SaturationRatio  float64
	AcquiredTotal    int64
	FailedTotal      int64
	ThroughputPerSec float64
	P50WaitMs        float64
	P95WaitMs        float64
	P99WaitMs        float64
}

// NewPoolTelemetryProbe creates a probe for a pool of capacity maxConns.
func NewPoolTelemetryProbe(maxConns int, windowDuration time.Duration) *PoolTelemetryProbe {
	return &PoolTelemetryProbe{
		maxConns:       maxConns,
		windowStart:    time.Now(),
		windowDuration: windowDuration,
	}
}

// RecordAcquire records a successful connection acquisition with its wait latency.
func (p *PoolTelemetryProbe) RecordAcquire(wait time.Duration) {
	p.acquiredTotal.Add(1)
	p.activeConns.Add(1)
	p.mu.Lock()
	p.waitSamples = append(p.waitSamples, wait)
	p.mu.Unlock()
}

// RecordRelease decrements the active connection counter.
func (p *PoolTelemetryProbe) RecordRelease() {
	p.activeConns.Add(-1)
}

// RecordFailure records a connection acquisition failure (timeout/exhaustion).
func (p *PoolTelemetryProbe) RecordFailure() {
	p.failedTotal.Add(1)
}

// Snapshot computes derived metrics over the current observation window.
func (p *PoolTelemetryProbe) Snapshot() PoolTelemetrySnapshot {
	p.mu.Lock()
	samples := make([]time.Duration, len(p.waitSamples))
	copy(samples, p.waitSamples)
	elapsed := time.Since(p.windowStart)
	p.mu.Unlock()

	active := p.activeConns.Load()
	acquired := p.acquiredTotal.Load()
	failed := p.failedTotal.Load()

	saturation := 0.0
	if p.maxConns > 0 {
		saturation = float64(active) / float64(p.maxConns)
	}

	throughput := 0.0
	if elapsed > 0 {
		throughput = float64(acquired) / elapsed.Seconds()
	}

	return PoolTelemetrySnapshot{
		ActiveConns:      active,
		MaxConns:         p.maxConns,
		SaturationRatio:  saturation,
		AcquiredTotal:    acquired,
		FailedTotal:      failed,
		ThroughputPerSec: throughput,
		P50WaitMs:        percentileMs(samples, 50),
		P95WaitMs:        percentileMs(samples, 95),
		P99WaitMs:        percentileMs(samples, 99),
	}
}

// ResetWindow clears the observation window for rolling statistics.
func (p *PoolTelemetryProbe) ResetWindow() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.waitSamples = p.waitSamples[:0]
	p.windowStart = time.Now()
}

// String returns a human-readable summary of the current snapshot.
func (s PoolTelemetrySnapshot) String() string {
	return fmt.Sprintf(
		"active=%d/%d sat=%.1f%% throughput=%.1f/s p50=%.2fms p95=%.2fms p99=%.2fms failures=%d",
		s.ActiveConns, s.MaxConns,
		s.SaturationRatio*100,
		s.ThroughputPerSec,
		s.P50WaitMs, s.P95WaitMs, s.P99WaitMs,
		s.FailedTotal,
	)
}

// percentileMs computes a given percentile (0-100) from a slice of durations, in milliseconds.
func percentileMs(samples []time.Duration, pct float64) float64 {
	n := len(samples)
	if n == 0 {
		return 0
	}
	sorted := make([]time.Duration, n)
	copy(sorted, samples)
	sortDurations(sorted)
	idx := int(math.Ceil(pct/100.0*float64(n))) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= n {
		idx = n - 1
	}
	return float64(sorted[idx].Milliseconds())
}

func sortDurations(ds []time.Duration) {
	// simple insertion sort for small slices
	for i := 1; i < len(ds); i++ {
		key := ds[i]
		j := i - 1
		for j >= 0 && ds[j] > key {
			ds[j+1] = ds[j]
			j--
		}
		ds[j+1] = key
	}
}
