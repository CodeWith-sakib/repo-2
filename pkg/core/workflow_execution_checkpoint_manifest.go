package core

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"
)

// WorkflowStepCheckpoint records snapshot state of a step for deterministic resumption.
type WorkflowStepCheckpoint struct {
	StepID            string            `json:"step_id"`
	Status            string            `json:"status"`
	StateSnapshotJSON string            `json:"state_snapshot_json"`
	StateChecksum     string            `json:"state_checksum"`
	CompletedAt       time.Time         `json:"completed_at"`
	ExecutionAttempt  int               `json:"execution_attempt"`
	OutputsMetadata   map[string]string `json:"outputs_metadata"`
}

// WorkflowExecutionCheckpointManifest stores consistent global checkpoint state across all DAG nodes.
type WorkflowExecutionCheckpointManifest struct {
	mu          sync.RWMutex
	WorkflowID  string                            `json:"workflow_id"`
	RunID       string                            `json:"run_id"`
	Epoch       int64                             `json:"epoch"`
	Checkpoints map[string]WorkflowStepCheckpoint `json:"checkpoints"`
	LastUpdated time.Time                         `json:"last_updated"`
}

// NewWorkflowExecutionCheckpointManifest creates a checkpoint state registry.
func NewWorkflowExecutionCheckpointManifest(workflowID, runID string, epoch int64) *WorkflowExecutionCheckpointManifest {
	return &WorkflowExecutionCheckpointManifest{
		WorkflowID:  workflowID,
		RunID:       runID,
		Epoch:       epoch,
		Checkpoints: make(map[string]WorkflowStepCheckpoint),
		LastUpdated: time.Now(),
	}
}

// RecordStepCheckpoint saves an immutable state snapshot for a completed step.
func (m *WorkflowExecutionCheckpointManifest) RecordStepCheckpoint(stepID, status, snapshotJSON string, attempt int, meta map[string]string) WorkflowStepCheckpoint {
	m.mu.Lock()
	defer m.mu.Unlock()

	hash := sha256.Sum256([]byte(snapshotJSON))
	chk := WorkflowStepCheckpoint{
		StepID:            stepID,
		Status:            status,
		StateSnapshotJSON: snapshotJSON,
		StateChecksum:     hex.EncodeToString(hash[:16]),
		CompletedAt:       time.Now(),
		ExecutionAttempt:  attempt,
		OutputsMetadata:   meta,
	}

	m.Checkpoints[stepID] = chk
	m.LastUpdated = time.Now()
	return chk
}

// GetCheckpoint retrieves checkpoint record if previously saved.
func (m *WorkflowExecutionCheckpointManifest) GetCheckpoint(stepID string) (WorkflowStepCheckpoint, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	chk, ok := m.Checkpoints[stepID]
	return chk, ok
}

// CompletedStepIDs returns all steps successfully checkpointed.
func (m *WorkflowExecutionCheckpointManifest) CompletedStepIDs() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ids := make([]string, 0, len(m.Checkpoints))
	for id, chk := range m.Checkpoints {
		if chk.Status == "COMPLETED" || chk.Status == "SUCCESS" {
			ids = append(ids, id)
		}
	}
	return ids
}
