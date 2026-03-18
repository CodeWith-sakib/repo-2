package worker

import (
	"runtime"
	"sync"
)

// WorkerCPUAffinityManager allocates dedicated CPU cores to high-performance worker tasks.
type WorkerCPUAffinityManager struct {
	mu            sync.Mutex
	availableCPUs int
	pinnedTasks   map[string]int // taskID -> coreID
}

// NewWorkerCPUAffinityManager creates an affinity manager querying system logical core count.
func NewWorkerCPUAffinityManager() *WorkerCPUAffinityManager {
	cpus := runtime.NumCPU()
	if cpus <= 0 {
		cpus = 4
	}
	return &WorkerCPUAffinityManager{
		availableCPUs: cpus,
		pinnedTasks:   make(map[string]int),
	}
}

// PinTask binds a task ID to an assigned CPU core index.
func (m *WorkerCPUAffinityManager) PinTask(taskID string) (int, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if core, exists := m.pinnedTasks[taskID]; exists {
		return core, true
	}

	if len(m.pinnedTasks) >= m.availableCPUs {
		return -1, false // No unpinned cores available
	}

	// Assign lowest available core index
	assignedCore := len(m.pinnedTasks) % m.availableCPUs
	m.pinnedTasks[taskID] = assignedCore
	return assignedCore, true
}

// UnpinTask releases assigned CPU core for a completed task.
func (m *WorkerCPUAffinityManager) UnpinTask(taskID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.pinnedTasks, taskID)
}

// ActivePinnedCount returns count of currently pinned worker routines.
func (m *WorkerCPUAffinityManager) ActivePinnedCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.pinnedTasks)
}
