package worker

import (
	"context"
	"sync"
	"time"
)

// FileDescriptorMetrics captures file handle counts per worker process.
type FileDescriptorMetrics struct {
	OpenFDCount int       `json:"open_fd_count"`
	MaxFDCap    int       `json:"max_fd_cap"`
	UsageRatio  float64   `json:"usage_ratio"`
	IsExhausted bool      `json:"is_exhausted"`
	SampledAt   time.Time `json:"sampled_at"`
}

// ProcessFDLeakMonitor checks file descriptor consumption to prevent socket/file resource exhaustion.
type ProcessFDLeakMonitor struct {
	mu       sync.RWMutex
	maxFD    int
	warnFD   int
}

// NewProcessFDLeakMonitor creates a file descriptor consumption monitor.
func NewProcessFDLeakMonitor(maxFD int) *ProcessFDLeakMonitor {
	if maxFD <= 0 {
		maxFD = 1024
	}
	return &ProcessFDLeakMonitor{
		maxFD:  maxFD,
		warnFD: int(float64(maxFD) * 0.80),
	}
}

// RecordFDObservation records observed open descriptor count.
func (m *ProcessFDLeakMonitor) RecordFDObservation(ctx context.Context, openCount int) FileDescriptorMetrics {
	m.mu.Lock()
	defer m.mu.Unlock()

	ratio := float64(openCount) / float64(m.maxFD)
	return FileDescriptorMetrics{
		OpenFDCount: openCount,
		MaxFDCap:    m.maxFD,
		UsageRatio:  ratio,
		IsExhausted: openCount >= m.warnFD,
		SampledAt:   time.Now(),
	}
}

// MaxCap returns configured ceiling.
func (m *ProcessFDLeakMonitor) MaxCap() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.maxFD
}
