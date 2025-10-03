package statemachine

import (
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

// WorkflowSnapshot records a full point-in-time state of a workflow run.
type WorkflowSnapshot struct {
	SnapshotID string                    `json:"snapshot_id"`
	RunID      string                    `json:"run_id"`
	WorkflowID string                    `json:"workflow_id"`
	State      core.RunState             `json:"state"`
	StepStates map[string]core.StepState `json:"step_states"`
	Variables  map[string]interface{}    `json:"variables"`
	Version    int64                     `json:"version"`
	CreatedAt  time.Time                 `json:"created_at"`
	Checksum   string                    `json:"checksum"`
}

// SnapshotManager coordinates snapshot creation, retention, and restoration.
type SnapshotManager struct {
	mu        sync.RWMutex
	snapshots map[string][]*WorkflowSnapshot // runID -> snapshots
	maxPerRun int
}

// NewSnapshotManager creates a snapshot manager with max snapshots to retain per run.
func NewSnapshotManager(maxPerRun int) *SnapshotManager {
	if maxPerRun <= 0 {
		maxPerRun = 10
	}
	return &SnapshotManager{
		snapshots: make(map[string][]*WorkflowSnapshot),
		maxPerRun: maxPerRun,
	}
}

// SaveSnapshot stores a new snapshot for a run, maintaining retention limits.
func (m *SnapshotManager) SaveSnapshot(snap *WorkflowSnapshot) error {
	if snap.RunID == "" {
		return fmt.Errorf("run ID cannot be empty")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	list := m.snapshots[snap.RunID]
	list = append(list, snap)

	// Sort by CreatedAt ascending
	sort.Slice(list, func(i, j int) bool {
		return list[i].CreatedAt.Before(list[j].CreatedAt)
	})

	// Prune older snapshots if exceeding max
	if len(list) > m.maxPerRun {
		overflow := len(list) - m.maxPerRun
		list = list[overflow:]
	}

	m.snapshots[snap.RunID] = list
	return nil
}

// GetLatest returns the most recent snapshot for a run.
func (m *SnapshotManager) GetLatest(runID string) (*WorkflowSnapshot, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list, ok := m.snapshots[runID]
	if !ok || len(list) == 0 {
		return nil, fmt.Errorf("no snapshots found for run %s", runID)
	}

	return list[len(list)-1], nil
}

// ListSnapshots returns all stored snapshots for a run in chronological order.
func (m *SnapshotManager) ListSnapshots(runID string) []*WorkflowSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list := m.snapshots[runID]
	res := make([]*WorkflowSnapshot, len(list))
	copy(res, list)
	return res
}

// RestoreState parses snapshot JSON and reconstructs a WorkflowSnapshot.
func RestoreState(data []byte) (*WorkflowSnapshot, error) {
	var snap WorkflowSnapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return nil, fmt.Errorf("failed to restore snapshot: %w", err)
	}
	return &snap, nil
}
