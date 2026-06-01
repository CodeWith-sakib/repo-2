package worker

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"
)

var (
	ErrQuotaExceeded = errors.New("worker memory quota exceeded")
)

// MemoryQuotaEnforcer tracks process heap allocation against allocated quotas.
type MemoryQuotaEnforcer struct {
	mu            sync.RWMutex
	maxHeapBytes  uint64
	stepAllocated map[string]uint64
}

// NewMemoryQuotaEnforcer constructs an enforcer with a given upper limit.
func NewMemoryQuotaEnforcer(maxHeapBytes uint64) *MemoryQuotaEnforcer {
	return &MemoryQuotaEnforcer{
		maxHeapBytes:  maxHeapBytes,
		stepAllocated: make(map[string]uint64),
	}
}

// AllocateQuota checks whether a step's requested memory can be accommodated.
func (e *MemoryQuotaEnforcer) AllocateQuota(ctx context.Context, stepID string, requestedBytes uint64) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	var totalAllocated uint64
	for _, b := range e.stepAllocated {
		totalAllocated += b
	}

	if e.maxHeapBytes > 0 && totalAllocated+requestedBytes > e.maxHeapBytes {
		return fmt.Errorf("%w: cannot allocate %d bytes (total %d / cap %d)",
			ErrQuotaExceeded, requestedBytes, totalAllocated+requestedBytes, e.maxHeapBytes)
	}

	e.stepAllocated[stepID] += requestedBytes
	return nil
}

// ReleaseQuota returns memory allocation when a step finishes.
func (e *MemoryQuotaEnforcer) ReleaseQuota(stepID string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.stepAllocated, stepID)
}

// CurrentAllocations returns total registered quota allocations.
func (e *MemoryQuotaEnforcer) CurrentAllocations() (uint64, uint64) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var total uint64
	for _, b := range e.stepAllocated {
		total += b
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return total, m.Alloc
}
