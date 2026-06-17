package postgres

import (
	"sync"
	"time"
)

// InMemLockWaitEdge models a transaction waiting for a lock held by another transaction.
type InMemLockWaitEdge struct {
	WaitingTXID int64     `json:"waiting_txid"`
	HoldingTXID int64     `json:"holding_txid"`
	ResourceID  string    `json:"resource_id"`
	RequestedAt time.Time `json:"requested_at"`
}

// DeadlockDetectorGraph tracks lock wait graphs and detects cyclic dependencies.
type DeadlockDetectorGraph struct {
	mu    sync.RWMutex
	edges []InMemLockWaitEdge
}

// NewDeadlockDetectorGraph creates an in-memory lock wait dependency graph.
func NewDeadlockDetectorGraph() *DeadlockDetectorGraph {
	return &DeadlockDetectorGraph{
		edges: make([]InMemLockWaitEdge, 0),
	}
}

// AddWaitEdge records that waitingTX is blocked waiting for holdingTX.
func (g *DeadlockDetectorGraph) AddWaitEdge(waitingTX, holdingTX int64, resourceID string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.edges = append(g.edges, InMemLockWaitEdge{
		WaitingTXID: waitingTX,
		HoldingTXID: holdingTX,
		ResourceID:  resourceID,
		RequestedAt: time.Now(),
	})
}

// ClearTXEdges removes all wait edges associated with a completed transaction.
func (g *DeadlockDetectorGraph) ClearTXEdges(txID int64) {
	g.mu.Lock()
	defer g.mu.Unlock()

	retained := make([]InMemLockWaitEdge, 0, len(g.edges))
	for _, e := range g.edges {
		if e.WaitingTXID != txID && e.HoldingTXID != txID {
			retained = append(retained, e)
		}
	}
	g.edges = retained
}

// DetectCycles finds circular waits (deadlocks) in the lock dependency graph.
func (g *DeadlockDetectorGraph) DetectCycles() [][]int64 {
	g.mu.RLock()
	defer g.mu.RUnlock()

	adj := make(map[int64][]int64)
	allNodes := make(map[int64]bool)
	for _, e := range g.edges {
		adj[e.WaitingTXID] = append(adj[e.WaitingTXID], e.HoldingTXID)
		allNodes[e.WaitingTXID] = true
		allNodes[e.HoldingTXID] = true
	}

	visited := make(map[int64]bool)
	recStack := make(map[int64]bool)
	path := make([]int64, 0)
	var cycles [][]int64

	var dfs func(u int64)
	dfs = func(u int64) {
		visited[u] = true
		recStack[u] = true
		path = append(path, u)

		for _, v := range adj[u] {
			if !visited[v] {
				dfs(v)
			} else if recStack[v] {
				for i, node := range path {
					if node == v {
						cycle := make([]int64, len(path[i:]))
						copy(cycle, path[i:])
						cycles = append(cycles, cycle)
						break
					}
				}
			}
		}

		path = path[:len(path)-1]
		recStack[u] = false
	}

	for node := range allNodes {
		if !visited[node] {
			dfs(node)
		}
	}

	return cycles
}
