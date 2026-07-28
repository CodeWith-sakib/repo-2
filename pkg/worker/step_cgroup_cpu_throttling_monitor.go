package worker

import (
	"sync"
	"time"
)

// CPUThrottlingMetrics records CFS cgroup throttling statistics.
type CPUThrottlingMetrics struct {
	PeriodsTotal   int64         `json:"periods_total"`
	ThrottledTotal int64         `json:"throttled_total"`
	ThrottleRatio  float64       `json:"throttle_ratio"`
	TimeThrottled  time.Duration `json:"time_throttled"`
	IsCritical     bool          `json:"is_critical"`
}

// CgroupCPUThrottlingMonitor checks Linux container CFS bandwidth quota throttling.
type CgroupCPUThrottlingMonitor struct {
	mu           sync.RWMutex
	warnRatio    float64
	lastPeriods  int64
	lastThrottled int64
}

// NewCgroupCPUThrottlingMonitor constructs a CFS throttling watcher.
func NewCgroupCPUThrottlingMonitor(warnRatio float64) *CgroupCPUThrottlingMonitor {
	if warnRatio <= 0.0 || warnRatio >= 1.0 {
		warnRatio = 0.15 // 15% periods throttled
	}
	return &CgroupCPUThrottlingMonitor{
		warnRatio: warnRatio,
	}
}

// RecordCFSStats records delta of cgroup cpu.stat metrics.
func (m *CgroupCPUThrottlingMonitor) RecordCFSStats(periods, throttled int64, throttledNanos int64) CPUThrottlingMetrics {
	m.mu.Lock()
	defer m.mu.Unlock()

	ratio := 0.0
	if periods > 0 {
		ratio = float64(throttled) / float64(periods)
	}

	metrics := CPUThrottlingMetrics{
		PeriodsTotal:   periods,
		ThrottledTotal: throttled,
		ThrottleRatio:  ratio,
		TimeThrottled:  time.Duration(throttledNanos) * time.Nanosecond,
		IsCritical:     ratio >= m.warnRatio,
	}

	m.lastPeriods = periods
	m.lastThrottled = throttled
	return metrics
}
