package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// TaskCheckpoint represents saved progress state of an iterative task.
type TaskCheckpoint struct {
	TaskID      string          `json:"task_id"`
	ProgressPct float64         `json:"progress_pct"`
	Stage       string          `json:"stage"`
	Data        json.RawMessage `json:"data"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// CheckpointRecorder provides an interface for tasks to store checkpoints.
type CheckpointRecorder interface {
	SaveCheckpoint(cp TaskCheckpoint) error
	LoadCheckpoint(taskID string) (*TaskCheckpoint, bool)
}

// MemoryCheckpointRecorder stores checkpoints in memory.
type MemoryCheckpointRecorder struct {
	mu          sync.RWMutex
	checkpoints map[string]TaskCheckpoint
}

// NewMemoryCheckpointRecorder creates an in-memory checkpoint store.
func NewMemoryCheckpointRecorder() *MemoryCheckpointRecorder {
	return &MemoryCheckpointRecorder{
		checkpoints: make(map[string]TaskCheckpoint),
	}
}

// SaveCheckpoint updates stored progress for a task.
func (r *MemoryCheckpointRecorder) SaveCheckpoint(cp TaskCheckpoint) error {
	if cp.TaskID == "" {
		return fmt.Errorf("task ID cannot be empty")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	cp.UpdatedAt = time.Now()
	r.checkpoints[cp.TaskID] = cp
	return nil
}

// LoadCheckpoint returns the last checkpoint recorded for a task.
func (r *MemoryCheckpointRecorder) LoadCheckpoint(taskID string) (*TaskCheckpoint, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cp, ok := r.checkpoints[taskID]
	if !ok {
		return nil, false
	}
	return &cp, true
}

// CooperativeExecutionSession coordinates checkpoints and cooperative cancellation for a single task run.
type CooperativeExecutionSession struct {
	ctx      context.Context
	taskID   string
	recorder CheckpointRecorder
}

// NewCooperativeExecutionSession initializes a session for a task run.
func NewCooperativeExecutionSession(ctx context.Context, taskID string, rec CheckpointRecorder) *CooperativeExecutionSession {
	return &CooperativeExecutionSession{
		ctx:      ctx,
		taskID:   taskID,
		recorder: rec,
	}
}

// CheckCancellation checks if task should abort cooperatively.
func (s *CooperativeExecutionSession) CheckCancellation() error {
	select {
	case <-s.ctx.Done():
		return s.ctx.Err()
	default:
		return nil
	}
}

// ReportProgress saves a progress percentage and state payload to the recorder.
func (s *CooperativeExecutionSession) ReportProgress(pct float64, stage string, state interface{}) error {
	if err := s.CheckCancellation(); err != nil {
		return err
	}

	bytes, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	cp := TaskCheckpoint{
		TaskID:      s.taskID,
		ProgressPct: pct,
		Stage:       stage,
		Data:        bytes,
	}
	return s.recorder.SaveCheckpoint(cp)
}

// ResumeState loads the last recorded checkpoint for this task, if any.
func (s *CooperativeExecutionSession) ResumeState() (*TaskCheckpoint, bool) {
	return s.recorder.LoadCheckpoint(s.taskID)
}
