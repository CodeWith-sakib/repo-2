package postgres

import (
	"context"
	"hash/fnv"
	"sync"
)

// DBConnectionNode represents a target database endpoint.
type DBConnectionNode struct {
	ID        string `json:"id"`
	Host      string `json:"host"`
	Port      int    `json:"port"`
	IsPrimary bool   `json:"is_primary"`
}

// ConnectionAffinityRouter routes queries to read replicas or primary based on write-intensity and tenant affinity.
type ConnectionAffinityRouter struct {
	mu        sync.RWMutex
	primary   DBConnectionNode
	readPool  []DBConnectionNode
}

// NewConnectionAffinityRouter initializes an affinity router.
func NewConnectionAffinityRouter(primary DBConnectionNode, replicas []DBConnectionNode) *ConnectionAffinityRouter {
	return &ConnectionAffinityRouter{
		primary:  primary,
		readPool: replicas,
	}
}

// RouteTarget selects connection endpoint for given tenant and query intent.
func (r *ConnectionAffinityRouter) RouteTarget(ctx context.Context, tenantID string, isWrite bool) DBConnectionNode {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if isWrite || len(r.readPool) == 0 {
		return r.primary
	}

	h := fnv.New32a()
	h.Write([]byte(tenantID))
	idx := int(h.Sum32()) % len(r.readPool)

	return r.readPool[idx]
}

// AddReplica dynamically registers an additional read replica.
func (r *ConnectionAffinityRouter) AddReplica(replica DBConnectionNode) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.readPool = append(r.readPool, replica)
}
