package worker

import (
	"sync"
	"time"
)

// IOThrottlingMetrics records Linux blkio / io.stat throttling statistics.
type IOThrottlingMetrics struct {
	ReadBytesSec  int64     `json:"read_bytes_sec"`
	WriteBytesSec int64     `json:"write_bytes_sec"`
	ReadIOPS      int64     `json:"read_iops"`
	WriteIOPS     int64     `json:"write_iops"`
	IsThrottled   bool      `json:"is_throttled"`
	ObservedAt    time.Time `json:"observed_at"`
}

// CgroupIOThrottlingMonitor monitors block I/O bandwidth and IOPS limits on worker storage mounts.
type CgroupIOThrottlingMonitor struct {
	mu         sync.RWMutex
	maxBPS     int64
	maxIOPS    int64
	warnBPS    int64
	warnIOPS   int64
}

// NewCgroupIOThrottlingMonitor initializes an I/O throttling observer.
func NewCgroupIOThrottlingMonitor(maxBytesPerSec, maxIOPS int64) *CgroupIOThrottlingMonitor {
	if maxBytesPerSec <= 0 {
		maxBytesPerSec = 100 * 1024 * 1024 // 100MB/s
	}
	if maxIOPS <= 0 {
		maxIOPS = 5000
	}
	return &CgroupIOThrottlingMonitor{
		maxBPS:   maxBytesPerSec,
		maxIOPS:  maxIOPS,
		warnBPS:  int64(float64(maxBytesPerSec) * 0.85),
		warnIOPS: int64(float64(maxIOPS) * 0.85),
	}
}

// RecordIOObservation records current transfer rates and detects impending disk bottlenecks.
func (m *CgroupIOThrottlingMonitor) RecordIOObservation(rBPS, wBPS, rIOPS, wIOPS int64) IOThrottlingMetrics {
	m.mu.Lock()
	defer m.mu.Unlock()

	totalBPS := rBPS + wBPS
	totalIOPS := rIOPS + wIOPS

	throttled := totalBPS >= m.warnBPS || totalIOPS >= m.warnIOPS

	return IOThrottlingMetrics{
		ReadBytesSec:  rBPS,
		WriteBytesSec: wBPS,
		ReadIOPS:      rIOPS,
		WriteIOPS:     wIOPS,
		IsThrottled:   throttled,
		ObservedAt:    time.Now(),
	}
}
