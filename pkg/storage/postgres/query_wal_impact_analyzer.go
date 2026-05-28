package postgres

import (
	"fmt"
	"sync"
	"time"
)

// WALImpactMetrics records write-ahead-log bytes generated and frequency.
type WALImpactMetrics struct {
	TotalWALBytes   int64         `json:"total_wal_bytes"`
	TotalOperations int64         `json:"total_operations"`
	AverageBytesOp  float64       `json:"avg_bytes_op"`
	EstimatedFsyncs int64         `json:"estimated_fsyncs"`
	SamplingWindow  time.Duration `json:"sampling_window"`
}

// QueryWALImpactAnalyzer estimates write amplification and WAL consumption.
type QueryWALImpactAnalyzer struct {
	mu             sync.Mutex
	walBytes       int64
	operations     int64
	fsyncCount     int64
	windowStart    time.Time
	segmentSize    int64
}

// NewQueryWALImpactAnalyzer initializes an analyzer.
func NewQueryWALImpactAnalyzer(segmentSizeBytes int64) *QueryWALImpactAnalyzer {
	if segmentSizeBytes <= 0 {
		segmentSizeBytes = 16 * 1024 * 1024 // 16MB default Postgres WAL segment
	}
	return &QueryWALImpactAnalyzer{
		windowStart: time.Now(),
		segmentSize: segmentSizeBytes,
	}
}

// RecordWrite records an insertion/update operation and estimated raw tuple payload.
func (a *QueryWALImpactAnalyzer) RecordWrite(tupleBytes int64) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// WAL record overhead is roughly 64 bytes per tuple plus tuple size
	walCost := tupleBytes + 64
	a.walBytes += walCost
	a.operations++
	if a.walBytes/a.segmentSize > (a.walBytes-walCost)/a.segmentSize {
		a.fsyncCount++
	}
}

// GetMetrics computes WAL generation metrics over the current observation window.
func (a *QueryWALImpactAnalyzer) GetMetrics() WALImpactMetrics {
	a.mu.Lock()
	defer a.mu.Unlock()

	avg := float64(0)
	if a.operations > 0 {
		avg = float64(a.walBytes) / float64(a.operations)
	}

	return WALImpactMetrics{
		TotalWALBytes:   a.walBytes,
		TotalOperations: a.operations,
		AverageBytesOp:  avg,
		EstimatedFsyncs: a.fsyncCount,
		SamplingWindow:  time.Since(a.windowStart),
	}
}

// SummaryString returns formatted WAL impact info.
func (m WALImpactMetrics) SummaryString() string {
	return fmt.Sprintf("WAL: %d bytes across %d ops (avg %.1f B/op), %d fsyncs",
		m.TotalWALBytes, m.TotalOperations, m.AverageBytesOp, m.EstimatedFsyncs)
}
