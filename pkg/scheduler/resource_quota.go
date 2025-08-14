package scheduler

import (
	"fmt"
	"sync"
)

// ResourceVector represents multi-dimensional compute requirements (CPU millicores, Memory MB, GPU units).
type ResourceVector struct {
	CPUMillicores int64 `json:"cpu_millicores"`
	MemoryMB      int64 `json:"memory_mb"`
	GPUUnits      int64 `json:"gpu_units"`
}

// Add sums two resource vectors.
func (r ResourceVector) Add(other ResourceVector) ResourceVector {
	return ResourceVector{
		CPUMillicores: r.CPUMillicores + other.CPUMillicores,
		MemoryMB:      r.MemoryMB + other.MemoryMB,
		GPUUnits:      r.GPUUnits + other.GPUUnits,
	}
}

// Sub subtracts a resource vector from another, bounding negative values at zero.
func (r ResourceVector) Sub(other ResourceVector) ResourceVector {
	cpu := r.CPUMillicores - other.CPUMillicores
	if cpu < 0 {
		cpu = 0
	}
	mem := r.MemoryMB - other.MemoryMB
	if mem < 0 {
		mem = 0
	}
	gpu := r.GPUUnits - other.GPUUnits
	if gpu < 0 {
		gpu = 0
	}
	return ResourceVector{
		CPUMillicores: cpu,
		MemoryMB:      mem,
		GPUUnits:      gpu,
	}
}

// FitsIn returns true if this vector does not exceed the capacity limit.
func (r ResourceVector) FitsIn(capacity ResourceVector) bool {
	return r.CPUMillicores <= capacity.CPUMillicores &&
		r.MemoryMB <= capacity.MemoryMB &&
		r.GPUUnits <= capacity.GPUUnits
}

// TenantQuotaManager tracks and bounds multi-dimensional resource usage per tenant.
type TenantQuotaManager struct {
	mu        sync.RWMutex
	limits    map[string]ResourceVector
	allocated map[string]ResourceVector
}

// NewTenantQuotaManager initializes a quota manager.
func NewTenantQuotaManager() *TenantQuotaManager {
	return &TenantQuotaManager{
		limits:    make(map[string]ResourceVector),
		allocated: make(map[string]ResourceVector),
	}
}

// SetLimit sets the maximum allowed resource quota for a tenant.
func (m *TenantQuotaManager) SetLimit(tenantID string, limit ResourceVector) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.limits[tenantID] = limit
}

// TryAcquire attempts to allocate resources for a tenant task without exceeding limits.
func (m *TenantQuotaManager) TryAcquire(tenantID string, req ResourceVector) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	limit, exists := m.limits[tenantID]
	if !exists {
		return fmt.Errorf("tenant %s has no configured resource quota", tenantID)
	}

	curr := m.allocated[tenantID]
	next := curr.Add(req)

	if !next.FitsIn(limit) {
		return fmt.Errorf("tenant %s resource limit exceeded (requested: %+v, current: %+v, limit: %+v)",
			tenantID, req, curr, limit)
	}

	m.allocated[tenantID] = next
	return nil
}

// Release frees previously acquired resources for a tenant.
func (m *TenantQuotaManager) Release(tenantID string, req ResourceVector) {
	m.mu.Lock()
	defer m.mu.Unlock()

	curr := m.allocated[tenantID]
	m.allocated[tenantID] = curr.Sub(req)
}

// GetUsage returns the current allocated resources for a tenant.
func (m *TenantQuotaManager) GetUsage(tenantID string) ResourceVector {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.allocated[tenantID]
}
