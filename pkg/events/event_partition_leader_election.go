package events

import (
	"context"
	"sync"
	"time"
)

// PartitionLease records active leader node and lease duration.
type PartitionLease struct {
	PartitionID int       `json:"partition_id"`
	LeaderNode  string    `json:"leader_node"`
	LeaseTTL    time.Duration `json:"lease_ttl"`
	GrantedAt   time.Time `json:"granted_at"`
}

// PartitionLeaderElection coordinates leadership leases per event stream partition.
type PartitionLeaderElection struct {
	mu     sync.RWMutex
	leases map[int]PartitionLease
}

// NewPartitionLeaderElection creates an in-memory lease coordinator.
func NewPartitionLeaderElection() *PartitionLeaderElection {
	return &PartitionLeaderElection{
		leases: make(map[int]PartitionLease),
	}
}

// TryAcquireLease attempts to claim leadership for a partition.
func (e *PartitionLeaderElection) TryAcquireLease(ctx context.Context, partitionID int, nodeID string, ttl time.Duration) bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	now := time.Now()
	if current, exists := e.leases[partitionID]; exists {
		// If current lease is still valid and held by another node, fail
		if now.Sub(current.GrantedAt) < current.LeaseTTL && current.LeaderNode != nodeID {
			return false
		}
	}

	e.leases[partitionID] = PartitionLease{
		PartitionID: partitionID,
		LeaderNode:  nodeID,
		LeaseTTL:    ttl,
		GrantedAt:   now,
	}
	return true
}

// GetLeader returns current active leader for partition if lease is not expired.
func (e *PartitionLeaderElection) GetLeader(partitionID int) (string, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	lease, exists := e.leases[partitionID]
	if !exists {
		return "", false
	}

	if time.Since(lease.GrantedAt) >= lease.LeaseTTL {
		return "", false // expired
	}

	return lease.LeaderNode, true
}
