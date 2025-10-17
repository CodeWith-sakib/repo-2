package postgres

import (
	"strings"
	"sync"
	"time"
)

// NodeRole indicates primary or read-replica role.
type NodeRole string

const (
	RolePrimary NodeRole = "PRIMARY"
	RoleReplica NodeRole = "REPLICA"
)

// DatabaseNode holds connectivity and telemetry info for a database instance.
type DatabaseNode struct {
	ID          string
	DSN         string
	Role        NodeRole
	IsHealthy   bool
	LagBytes    int64
	ActiveConns int
}

// RouterConfig configures routing policies.
type RouterConfig struct {
	MaxLagBytes    int64         // max allowed replication lag
	StickyDuration time.Duration // read-your-writes session stickiness window
}

// DefaultRouterConfig returns standard production defaults.
func DefaultRouterConfig() RouterConfig {
	return RouterConfig{
		MaxLagBytes:    10 * 1024 * 1024, // 10MB
		StickyDuration: 5 * time.Second,
	}
}

// ReplicaRouter manages read-write splitting across primary and read replicas.
type ReplicaRouter struct {
	mu           sync.RWMutex
	primary      *DatabaseNode
	replicas     []*DatabaseNode
	cfg          RouterConfig
	sessionWrite map[string]time.Time // sessionID -> lastWriteAt
	roundRobin   int
}

// NewReplicaRouter creates a router with a primary node.
func NewReplicaRouter(primary DatabaseNode, cfg RouterConfig) *ReplicaRouter {
	primary.Role = RolePrimary
	primary.IsHealthy = true
	return &ReplicaRouter{
		primary:      &primary,
		cfg:          cfg,
		sessionWrite: make(map[string]time.Time),
	}
}

// AddReplica registers a read-replica node.
func (r *ReplicaRouter) AddReplica(replica DatabaseNode) {
	r.mu.Lock()
	defer r.mu.Unlock()
	replica.Role = RoleReplica
	replica.IsHealthy = true
	r.replicas = append(r.replicas, &replica)
}

// RecordSessionWrite notes that a session performed a write operation (activates read-your-writes stickiness).
func (r *ReplicaRouter) RecordSessionWrite(sessionID string) {
	if sessionID == "" {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessionWrite[sessionID] = time.Now()
}

// RouteQuery determines whether a query should be executed on Primary or a Read Replica.
func (r *ReplicaRouter) RouteQuery(sql string, sessionID string) (*DatabaseNode, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	trimmed := strings.TrimSpace(strings.ToUpper(sql))

	// Writes and locking reads always go to Primary
	isWrite := strings.HasPrefix(trimmed, "INSERT") ||
		strings.HasPrefix(trimmed, "UPDATE") ||
		strings.HasPrefix(trimmed, "DELETE") ||
		strings.HasPrefix(trimmed, "CREATE") ||
		strings.HasPrefix(trimmed, "DROP") ||
		strings.HasPrefix(trimmed, "ALTER") ||
		strings.Contains(trimmed, "FOR UPDATE") ||
		strings.Contains(trimmed, "FOR SHARE")

	if isWrite {
		if sessionID != "" {
			r.sessionWrite[sessionID] = time.Now()
		}
		return r.primary, nil
	}

	// Check read-your-writes session stickiness
	if sessionID != "" {
		if lastWrite, ok := r.sessionWrite[sessionID]; ok {
			if time.Since(lastWrite) < r.cfg.StickyDuration {
				// Sticky window active: route read to primary to guarantee consistency
				return r.primary, nil
			}
			// Sticky window expired
			delete(r.sessionWrite, sessionID)
		}
	}

	// Read query: route to healthy read replica within lag threshold
	var candidates []*DatabaseNode
	for _, rep := range r.replicas {
		if rep.IsHealthy && rep.LagBytes <= r.cfg.MaxLagBytes {
			candidates = append(candidates, rep)
		}
	}

	if len(candidates) == 0 {
		// Fallback to primary if no healthy replicas available
		return r.primary, nil
	}

	// Round-robin selection across candidates
	idx := r.roundRobin % len(candidates)
	r.roundRobin++
	return candidates[idx], nil
}

// UpdateReplicaHealth updates the health and lag metrics of a replica.
func (r *ReplicaRouter) UpdateReplicaHealth(id string, healthy bool, lagBytes int64) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, rep := range r.replicas {
		if rep.ID == id {
			rep.IsHealthy = healthy
			rep.LagBytes = lagBytes
			break
		}
	}
}
