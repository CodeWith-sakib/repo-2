package worker

import (
	"errors"
	"fmt"
	"sync"
)

var (
	ErrNoAvailableCPUCores = errors.New("insufficient CPU cores available for task affinity allocation")
)

// CPUAffinityGroup allocates specific CPU socket/core IDs to computational tasks.
type CPUAffinityGroup struct {
	mu           sync.RWMutex
	totalCores   int
	coreAllocMap map[int]string // coreID -> stepID
}

// NewCPUAffinityGroup initializes core tracking for a multi-core worker machine.
func NewCPUAffinityGroup(totalCores int) *CPUAffinityGroup {
	if totalCores <= 0 {
		totalCores = 8
	}
	return &CPUAffinityGroup{
		totalCores:   totalCores,
		coreAllocMap: make(map[int]string),
	}
}

// AllocateCores pins a step to N dedicated physical cores.
func (g *CPUAffinityGroup) AllocateCores(stepID string, requestedCores int) ([]int, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	var available []int
	for c := 0; c < g.totalCores; c++ {
		if _, occupied := g.coreAllocMap[c]; !occupied {
			available = append(available, c)
		}
	}

	if len(available) < requestedCores {
		return nil, fmt.Errorf("%w: requested %d, available %d", ErrNoAvailableCPUCores, requestedCores, len(available))
	}

	allocated := available[:requestedCores]
	for _, c := range allocated {
		g.coreAllocMap[c] = stepID
	}
	return allocated, nil
}

// ReleaseCores unbinds all cores assigned to stepID.
func (g *CPUAffinityGroup) ReleaseCores(stepID string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	for c, owner := range g.coreAllocMap {
		if owner == stepID {
			delete(g.coreAllocMap, c)
		}
	}
}

// Utilization returns fraction of cores currently locked.
func (g *CPUAffinityGroup) Utilization() float64 {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return float64(len(g.coreAllocMap)) / float64(g.totalCores)
}
