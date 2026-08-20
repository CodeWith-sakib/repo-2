package worker

import (
	"context"
	"runtime"
	"sync"
	"time"
)

// ProcessRSSMetrics captures resident memory usage and garbage collector stats.
type ProcessRSSMetrics struct {
	AllocBytes      uint64    `json:"alloc_bytes"`
	TotalAllocBytes uint64    `json:"total_alloc_bytes"`
	SysBytes        uint64    `json:"sys_bytes"`
	NumGC           uint32    `json:"num_gc"`
	SampledAt       time.Time `json:"sampled_at"`
}

// ProcessRSSMemoryWatcher monitors Go runtime allocator consumption and forces emergency GC if memory balloons.
type ProcessRSSMemoryWatcher struct {
	mu           sync.RWMutex
	maxAlloc     uint64
	lastMetrics  ProcessRSSMetrics
	gcTriggerCnt int
}

// NewProcessRSSMemoryWatcher initializes a memory consumption watcher.
func NewProcessRSSMemoryWatcher(maxAllocBytes uint64) *ProcessRSSMemoryWatcher {
	return &ProcessRSSMemoryWatcher{
		maxAlloc: maxAllocBytes,
	}
}

// SampleMemory reads current runtime memory stats and evaluates allocation ceilings.
func (w *ProcessRSSMemoryWatcher) SampleMemory(ctx context.Context) ProcessRSSMetrics {
	w.mu.Lock()
	defer w.mu.Unlock()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	metrics := ProcessRSSMetrics{
		AllocBytes:      m.Alloc,
		TotalAllocBytes: m.TotalAlloc,
		SysBytes:        m.Sys,
		NumGC:           m.NumGC,
		SampledAt:       time.Now(),
	}

	if w.maxAlloc > 0 && m.Alloc > w.maxAlloc {
		runtime.GC()
		w.gcTriggerCnt++
	}

	w.lastMetrics = metrics
	return metrics
}

// EmergencyGCTriggerCount returns number of forced GC sweeps triggered by allocator spikes.
func (w *ProcessRSSMemoryWatcher) EmergencyGCTriggerCount() int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.gcTriggerCnt
}
