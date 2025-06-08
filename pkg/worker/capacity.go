package worker

import (
	"sync"
)

type DynamicCapacityManager struct {
	mu       sync.RWMutex
	capacity int
}

func NewDynamicCapacityManager(initial int) *DynamicCapacityManager {
	if initial <= 0 {
		initial = 4
	}
	return &DynamicCapacityManager{capacity: initial}
}

func (m *DynamicCapacityManager) GetCapacity() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.capacity
}

func (m *DynamicCapacityManager) SetCapacity(c int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if c > 0 {
		m.capacity = c
	}
}
