package worker

import (
	"hash/fnv"
	"sync"
)

// TenantAffinityLoadBalancer routes tasks to persistent worker nodes based on consistent hashing of tenant IDs.
type TenantAffinityLoadBalancer struct {
	mu      sync.RWMutex
	workers []string
}

// NewTenantAffinityLoadBalancer creates an affinity balancer with worker cluster nodes.
func NewTenantAffinityLoadBalancer(workers []string) *TenantAffinityLoadBalancer {
	return &TenantAffinityLoadBalancer{
		workers: append([]string(nil), workers...),
	}
}

// SelectWorker maps a tenant or workflow key to an assigned affinity worker node.
func (b *TenantAffinityLoadBalancer) SelectWorker(key string) string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if len(b.workers) == 0 {
		return ""
	}

	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	idx := int(h.Sum32()) % len(b.workers)
	if idx < 0 {
		idx = -idx
	}

	return b.workers[idx]
}

// UpdateWorkers updates the active pool of available worker nodes.
func (b *TenantAffinityLoadBalancer) UpdateWorkers(workers []string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.workers = append([]string(nil), workers...)
}
