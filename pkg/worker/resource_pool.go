package worker

import (
	"fmt"
	"sync"
	"time"
)

// ResourceVector represents multi-dimensional compute quotas.
type ResourceVector struct {
	CPUMillicores int64 // e.g. 1000 = 1 CPU core
	MemoryMB      int64 // e.g. 2048 MB
	GPUCount      int   // e.g. 1 GPU
}

// ResourceLease represents an active reservation of compute resources.
type ResourceLease struct {
	LeaseID    string
	TaskID     string
	Allocated  ResourceVector
	AcquiredAt time.Time
}

// MultiDimensionalResourcePool manages allocation and reclamation of multi-vector compute slots.
type MultiDimensionalResourcePool struct {
	mu        sync.RWMutex
	capacity  ResourceVector
	allocated ResourceVector
	leases    map[string]*ResourceLease
}

// NewMultiDimensionalResourcePool creates a pool with maximum capacity.
func NewMultiDimensionalResourcePool(capacity ResourceVector) (*MultiDimensionalResourcePool, error) {
	if capacity.CPUMillicores <= 0 || capacity.MemoryMB <= 0 {
		return nil, fmt.Errorf("capacity must have positive CPU and memory: %+v", capacity)
	}
	return &MultiDimensionalResourcePool{
		capacity: capacity,
		leases:   make(map[string]*ResourceLease),
	}, nil
}

// TryAllocate attempts to reserve the requested resources. Returns lease on success.
func (p *MultiDimensionalResourcePool) TryAllocate(taskID string, req ResourceVector) (*ResourceLease, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Check if sufficient resources remain
	if p.allocated.CPUMillicores+req.CPUMillicores > p.capacity.CPUMillicores {
		return nil, fmt.Errorf("insufficient CPU: requested %dm, available %dm",
			req.CPUMillicores, p.capacity.CPUMillicores-p.allocated.CPUMillicores)
	}
	if p.allocated.MemoryMB+req.MemoryMB > p.capacity.MemoryMB {
		return nil, fmt.Errorf("insufficient memory: requested %dMB, available %dMB",
			req.MemoryMB, p.capacity.MemoryMB-p.allocated.MemoryMB)
	}
	if p.allocated.GPUCount+req.GPUCount > p.capacity.GPUCount {
		return nil, fmt.Errorf("insufficient GPU: requested %d, available %d",
			req.GPUCount, p.capacity.GPUCount-p.allocated.GPUCount)
	}

	leaseID := fmt.Sprintf("lease-%s-%d", taskID, time.Now().UnixNano())
	lease := &ResourceLease{
		LeaseID:    leaseID,
		TaskID:     taskID,
		Allocated:  req,
		AcquiredAt: time.Now(),
	}

	p.allocated.CPUMillicores += req.CPUMillicores
	p.allocated.MemoryMB += req.MemoryMB
	p.allocated.GPUCount += req.GPUCount
	p.leases[leaseID] = lease

	return lease, nil
}

// Release frees resources held by a lease.
func (p *MultiDimensionalResourcePool) Release(leaseID string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	lease, exists := p.leases[leaseID]
	if !exists {
		return false
	}

	p.allocated.CPUMillicores -= lease.Allocated.CPUMillicores
	p.allocated.MemoryMB -= lease.Allocated.MemoryMB
	p.allocated.GPUCount -= lease.Allocated.GPUCount
	delete(p.leases, leaseID)

	return true
}

// Utilization returns percentage utilization for CPU, memory, and GPU.
func (p *MultiDimensionalResourcePool) Utilization() (cpuPct, memPct, gpuPct float64) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.capacity.CPUMillicores > 0 {
		cpuPct = (float64(p.allocated.CPUMillicores) / float64(p.capacity.CPUMillicores)) * 100.0
	}
	if p.capacity.MemoryMB > 0 {
		memPct = (float64(p.allocated.MemoryMB) / float64(p.capacity.MemoryMB)) * 100.0
	}
	if p.capacity.GPUCount > 0 {
		gpuPct = (float64(p.allocated.GPUCount) / float64(p.capacity.GPUCount)) * 100.0
	}
	return cpuPct, memPct, gpuPct
}
