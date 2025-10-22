package postgres

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

// LockWaitEdge represents a transaction waiting on a lock held by another transaction.
type LockWaitEdge struct {
	WaiterTx  string
	HolderTx  string
	LockID    string
	WaitSince time.Time
}

// DeadlockCycle contains the list of transactions forming a cyclic dependency.
type DeadlockCycle struct {
	Transactions []string
	VictimTx     string
	DetectedAt   time.Time
}

// DeadlockDetector tracks active lock dependencies and detects circular waits.
type DeadlockDetector struct {
	mu    sync.RWMutex
	edges map[string]map[string]string // waiter -> holder -> lockID
	txAge map[string]time.Time
}

// NewDeadlockDetector creates a deadlock detector.
func NewDeadlockDetector() *DeadlockDetector {
	return &DeadlockDetector{
		edges: make(map[string]map[string]string),
		txAge: make(map[string]time.Time),
	}
}

// RegisterTx registers a transaction start time for victim selection.
func (d *DeadlockDetector) RegisterTx(txID string, startedAt time.Time) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.txAge[txID] = startedAt
}

// UnregisterTx removes a transaction and its associated wait edges.
func (d *DeadlockDetector) UnregisterTx(txID string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	delete(d.txAge, txID)
	delete(d.edges, txID)

	for _, holders := range d.edges {
		delete(holders, txID)
	}
}

// AddWaitEdge records that waiterTx is blocked waiting for holderTx to release lockID.
func (d *DeadlockDetector) AddWaitEdge(waiterTx, holderTx, lockID string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, ok := d.edges[waiterTx]; !ok {
		d.edges[waiterTx] = make(map[string]string)
	}
	d.edges[waiterTx][holderTx] = lockID
}

// RemoveWaitEdge removes the specific lock wait dependency.
func (d *DeadlockDetector) RemoveWaitEdge(waiterTx, holderTx string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if holders, ok := d.edges[waiterTx]; ok {
		delete(holders, holderTx)
		if len(holders) == 0 {
			delete(d.edges, waiterTx)
		}
	}
}

// DetectCycles searches for circular wait paths and selects the youngest transaction as victim.
func (d *DeadlockDetector) DetectCycles() []*DeadlockCycle {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var cycles []*DeadlockCycle
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	parent := make(map[string]string)

	// Collect all tx nodes
	nodes := make(map[string]bool)
	for w, holders := range d.edges {
		nodes[w] = true
		for h := range holders {
			nodes[h] = true
		}
	}

	var sortedNodes []string
	for n := range nodes {
		sortedNodes = append(sortedNodes, n)
	}
	sort.Strings(sortedNodes)

	for _, node := range sortedNodes {
		if !visited[node] {
			d.dfs(node, visited, recStack, parent, &cycles)
		}
	}

	return cycles
}

func (d *DeadlockDetector) dfs(curr string, visited, recStack map[string]bool, parent map[string]string, cycles *[]*DeadlockCycle) {
	visited[curr] = true
	recStack[curr] = true

	for next := range d.edges[curr] {
		if !visited[next] {
			parent[next] = curr
			d.dfs(next, visited, recStack, parent, cycles)
		} else if recStack[next] {
			// Cycle found! Backtrack cycle path
			cyclePath := []string{next}
			for p := curr; p != next && p != ""; p = parent[p] {
				cyclePath = append([]string{p}, cyclePath...)
			}

			victim := d.selectVictim(cyclePath)
			*cycles = append(*cycles, &DeadlockCycle{
				Transactions: cyclePath,
				VictimTx:     victim,
				DetectedAt:   time.Now(),
			})
		}
	}

	recStack[curr] = false
}

// selectVictim picks the newest transaction (highest start time) to minimize rollback work.
func (d *DeadlockDetector) selectVictim(cycle []string) string {
	if len(cycle) == 0 {
		return ""
	}

	victim := cycle[0]
	youngestTime := d.txAge[victim]

	for _, tx := range cycle[1:] {
		t, ok := d.txAge[tx]
		if ok && t.After(youngestTime) {
			victim = tx
			youngestTime = t
		}
	}

	return victim
}

// FormatCycle returns a readable string description of the deadlock.
func (c *DeadlockCycle) FormatCycle() string {
	return fmt.Sprintf("Deadlock cycle: %v (recommended victim to abort: %s)", c.Transactions, c.VictimTx)
}
