package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

type RunSnapshot struct {
	RunID        core.ID                 `json:"run_id"`
	WorkflowID   core.ID                 `json:"workflow_id"`
	Version      int                     `json:"version"`
	State        core.RunState           `json:"state"`
	StepStates   map[string]core.StepState `json:"step_states"`
	StepOutputs  map[string]json.RawMessage `json:"step_outputs,omitempty"`
	CheckpointAt time.Time               `json:"checkpoint_at"`
}

type SnapshotManager struct {
	mu        sync.RWMutex
	snapshots map[core.ID][]*RunSnapshot
}

func NewSnapshotManager() *SnapshotManager {
	return &SnapshotManager{
		snapshots: make(map[core.ID][]*RunSnapshot),
	}
}

func (sm *SnapshotManager) CreateCheckpoint(ctx context.Context, run *core.WorkflowRun, steps []*core.StepRun) (*RunSnapshot, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	stepStates := make(map[string]core.StepState)
	stepOutputs := make(map[string]json.RawMessage)

	for _, s := range steps {
		stepStates[s.StepID] = s.State
		if len(s.Output) > 0 {
			stepOutputs[s.StepID] = s.Output
		}
	}

	snap := &RunSnapshot{
		RunID:        run.ID,
		WorkflowID:   run.WorkflowID,
		Version:      run.Version,
		State:        run.State,
		StepStates:   stepStates,
		StepOutputs:  stepOutputs,
		CheckpointAt: time.Now().UTC(),
	}

	sm.snapshots[run.ID] = append(sm.snapshots[run.ID], snap)
	return snap, nil
}

func (sm *SnapshotManager) GetLatest(runID core.ID) (*RunSnapshot, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	list, exists := sm.snapshots[runID]
	if !exists || len(list) == 0 {
		return nil, fmt.Errorf("%w: no snapshot found for run %s", core.ErrNotFound, runID)
	}

	return list[len(list)-1], nil
}
