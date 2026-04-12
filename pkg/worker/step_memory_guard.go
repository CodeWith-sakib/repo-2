package worker

import (
	"fmt"
	"runtime"
	"sync"
)

// StepMemoryGuard inspects runtime heap allocations and throttles worker admissions during memory pressure.
type StepMemoryGuard struct {
	mu           sync.Mutex
	maxAllocBytes uint64
}

// NewStepMemoryGuard creates a heap memory guard with alloc limit.
func NewStepMemoryGuard(maxAllocMB uint64) *StepMemoryGuard {
	if maxAllocMB == 0 {
		maxAllocMB = 1024 // 1GB default
	}
	return &StepMemoryGuard{
		maxAllocBytes: maxAllocMB * 1024 * 1024,
	}
}

// CanAdmitTask checks if current heap alloc allows admitting a new step.
func (g *StepMemoryGuard) CanAdmitTask(estimatedStepMB uint64) (bool, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	estBytes := estimatedStepMB * 1024 * 1024
	if m.Alloc+estBytes > g.maxAllocBytes {
		return false, fmt.Errorf("insufficient memory: current alloc %d MB + est %d MB exceeds ceiling %d MB",
			m.Alloc/(1024*1024), estimatedStepMB, g.maxAllocBytes/(1024*1024))
	}

	return true, nil
}

// CurrentAllocMB returns current runtime heap allocation in megabytes.
func (g *StepMemoryGuard) CurrentAllocMB() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.Alloc / (1024 * 1024)
}
