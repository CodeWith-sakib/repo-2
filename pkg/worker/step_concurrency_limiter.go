package worker

import (
	"sync"
)

// TaskConcurrencyGovernor enforces global worker concurrency ceilings across heterogeneous task pools.
type TaskConcurrencyGovernor struct {
	mu           sync.Mutex
	globalLimit  int
	currentCount int
}

// NewTaskConcurrencyGovernor creates a concurrency governor.
func NewTaskConcurrencyGovernor(globalLimit int) *TaskConcurrencyGovernor {
	if globalLimit <= 0 {
		globalLimit = 100
	}
	return &TaskConcurrencyGovernor{
		globalLimit: globalLimit,
	}
}

// TryAcquire checks and increments active worker counter.
func (g *TaskConcurrencyGovernor) TryAcquire() bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.currentCount < g.globalLimit {
		g.currentCount++
		return true
	}
	return false
}

// Release yields worker execution slot.
func (g *TaskConcurrencyGovernor) Release() {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.currentCount > 0 {
		g.currentCount--
	}
}

// CurrentInFlight returns count of active executions.
func (g *TaskConcurrencyGovernor) CurrentInFlight() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.currentCount
}
