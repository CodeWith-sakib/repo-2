package postgres

import (
	"fmt"
	"sync"
	"time"
)

// ConnState represents the current activity state of a tracked pool connection.
type ConnState string

const (
	ConnStateIdle    ConnState = "IDLE"
	ConnStateActive  ConnState = "ACTIVE"
	ConnStateInTx    ConnState = "IN_TRANSACTION"
	ConnStateSuspect ConnState = "SUSPECT"
	ConnStateDead    ConnState = "DEAD"
)

// TrackedConnection contains diagnostic health metadata for a single database connection.
type TrackedConnection struct {
	ID             int64
	State          ConnState
	CreatedAt      time.Time
	LastActiveAt   time.Time
	LeasedAt       time.Time
	LeaseHolder    string
	ConsecutiveErr int
}

// ConnectionHealthConfig defines timeout and reaper thresholds.
type ConnectionHealthConfig struct {
	MaxLeaseDuration    time.Duration
	MaxIdleDuration     time.Duration
	MaxConsecutiveError int
}

// DefaultConnectionHealthConfig returns standard production settings.
func DefaultConnectionHealthConfig() ConnectionHealthConfig {
	return ConnectionHealthConfig{
		MaxLeaseDuration:    5 * time.Minute,
		MaxIdleDuration:     15 * time.Minute,
		MaxConsecutiveError: 3,
	}
}

// ConnectionHealthReaper monitors connection lifespans and detects leaked or stale connections.
type ConnectionHealthReaper struct {
	mu          sync.RWMutex
	conns       map[int64]*TrackedConnection
	cfg         ConnectionHealthConfig
	leaksReaped int64
}

// NewConnectionHealthReaper initializes a new connection reaper.
func NewConnectionHealthReaper(cfg ConnectionHealthConfig) *ConnectionHealthReaper {
	return &ConnectionHealthReaper{
		conns: make(map[int64]*TrackedConnection),
		cfg:   cfg,
	}
}

// RegisterConnection adds a connection to the health monitor.
func (r *ConnectionHealthReaper) RegisterConnection(id int64) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	r.conns[id] = &TrackedConnection{
		ID:           id,
		State:        ConnStateIdle,
		CreatedAt:    now,
		LastActiveAt: now,
	}
}

// UnregisterConnection removes a connection from the monitor.
func (r *ConnectionHealthReaper) UnregisterConnection(id int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.conns, id)
}

// RecordAcquire marks a connection as active under a lease holder.
func (r *ConnectionHealthReaper) RecordAcquire(id int64, holder string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	conn, exists := r.conns[id]
	if !exists {
		return fmt.Errorf("connection %d not tracked", id)
	}

	conn.State = ConnStateActive
	conn.LeaseHolder = holder
	conn.LeasedAt = time.Now()
	conn.LastActiveAt = time.Now()
	return nil
}

// RecordRelease returns a connection to idle state.
func (r *ConnectionHealthReaper) RecordRelease(id int64) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if conn, exists := r.conns[id]; exists {
		conn.State = ConnStateIdle
		conn.LeaseHolder = ""
		conn.LastActiveAt = time.Now()
		conn.ConsecutiveErr = 0
	}
}

// RecordError increments consecutive errors for a connection, marking it suspect if threshold exceeded.
func (r *ConnectionHealthReaper) RecordError(id int64) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if conn, exists := r.conns[id]; exists {
		conn.ConsecutiveErr++
		if conn.ConsecutiveErr >= r.cfg.MaxConsecutiveError {
			conn.State = ConnStateSuspect
		}
	}
}

// ReapStaleConnections scans all tracked connections and identifies leaked or dead connection IDs.
func (r *ConnectionHealthReaper) ReapStaleConnections() (leakedIDs []int64, deadIDs []int64) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	for id, conn := range r.conns {
		// 1. Check for leaked active lease
		if conn.State == ConnStateActive || conn.State == ConnStateInTx {
			if !conn.LeasedAt.IsZero() && now.Sub(conn.LeasedAt) > r.cfg.MaxLeaseDuration {
				conn.State = ConnStateDead
				leakedIDs = append(leakedIDs, id)
				r.leaksReaped++
				continue
			}
		}

		// 2. Check for dead/suspect connections
		if conn.State == ConnStateSuspect {
			conn.State = ConnStateDead
			deadIDs = append(deadIDs, id)
		}
	}

	return leakedIDs, deadIDs
}

// ReapedLeakCount returns total leaked connections detected over lifetime.
func (r *ConnectionHealthReaper) ReapedLeakCount() int64 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.leaksReaped
}
