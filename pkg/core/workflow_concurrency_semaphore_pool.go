package core

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrSemaphoreAcquisitionTimeout = errors.New("timeout acquiring workflow concurrency semaphore permit")
)

// WorkflowSemaphorePool limits the number of simultaneously executing workflows per tenant.
type WorkflowSemaphorePool struct {
	mu         sync.Mutex
	capacities map[string]int
	active     map[string]int
}

// NewWorkflowSemaphorePool creates a multi-tenant concurrency pool.
func NewWorkflowSemaphorePool() *WorkflowSemaphorePool {
	return &WorkflowSemaphorePool{
		capacities: make(map[string]int),
		active:     make(map[string]int),
	}
}

// SetTenantLimit defines the maximum parallel workflows allowed for tenantID.
func (p *WorkflowSemaphorePool) SetTenantLimit(tenantID string, limit int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if limit <= 0 {
		limit = 10
	}
	p.capacities[tenantID] = limit
}

// TryAcquire attempts to allocate a concurrency slot immediately without blocking.
func (p *WorkflowSemaphorePool) TryAcquire(tenantID string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	limit, ok := p.capacities[tenantID]
	if !ok {
		limit = 10 // default fallback
	}

	if p.active[tenantID] >= limit {
		return false
	}

	p.active[tenantID]++
	return true
}

// Acquire waits for an available slot or fails if context deadline is exceeded.
func (p *WorkflowSemaphorePool) Acquire(ctx context.Context, tenantID string) error {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		if p.TryAcquire(tenantID) {
			return nil
		}
		select {
		case <-ctx.Done():
			return ErrSemaphoreAcquisitionTimeout
		case <-ticker.C:
		}
	}
}

// Release yields an active slot back to the tenant pool.
func (p *WorkflowSemaphorePool) Release(tenantID string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.active[tenantID] > 0 {
		p.active[tenantID]--
	}
}

// ActiveCount returns current concurrent executions for tenant.
func (p *WorkflowSemaphorePool) ActiveCount(tenantID string) int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.active[tenantID]
}
