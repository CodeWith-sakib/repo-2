package worker

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// SandboxIsolationLimits defines resource boundaries for an isolated task workspace.
type SandboxIsolationLimits struct {
	MaxDiskMB      int64
	MaxRuntimeSec  int
	AllowedEnvVars []string
}

// ExecutionSandboxJail sets up temporary quarantined workspace directories with cleanup guarantees.
type ExecutionSandboxJail struct {
	mu         sync.Mutex
	baseDir    string
	activeRuns map[string]string // runID -> isolated dir path
}

// NewExecutionSandboxJail creates a sandboxed directory isolation manager.
func NewExecutionSandboxJail(baseDir string) (*ExecutionSandboxJail, error) {
	if baseDir == "" {
		baseDir = filepath.Join(os.TempDir(), "kf_sandboxes")
	}
	if err := os.MkdirAll(baseDir, 0700); err != nil {
		return nil, err
	}
	return &ExecutionSandboxJail{
		baseDir:    baseDir,
		activeRuns: make(map[string]string),
	}, nil
}

// CreateJail provisions an isolated directory for task execution.
func (j *ExecutionSandboxJail) CreateJail(runID, stepID string) (string, error) {
	if runID == "" || stepID == "" {
		return "", errors.New("runID and stepID are required")
	}

	j.mu.Lock()
	defer j.mu.Unlock()

	jailPath := filepath.Join(j.baseDir, runID+"_"+stepID)
	if err := os.MkdirAll(jailPath, 0700); err != nil {
		return "", err
	}

	j.activeRuns[runID+"_"+stepID] = jailPath
	return jailPath, nil
}

// DestroyJail removes the quarantine workspace directory.
func (j *ExecutionSandboxJail) DestroyJail(runID, stepID string) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	key := runID + "_" + stepID
	path, exists := j.activeRuns[key]
	if !exists {
		return nil
	}

	delete(j.activeRuns, key)
	return os.RemoveAll(path)
}

// RunInJail executes a task action inside the protected jail directory with timeout enforcement.
func (j *ExecutionSandboxJail) RunInJail(ctx context.Context, runID, stepID string, timeout time.Duration, fn func(jailDir string) error) error {
	jailDir, err := j.CreateJail(runID, stepID)
	if err != nil {
		return err
	}
	defer func() {
		_ = j.DestroyJail(runID, stepID)
	}()

	ctxTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- fn(jailDir)
	}()

	select {
	case <-ctxTimeout.Done():
		return ctxTimeout.Err()
	case err := <-done:
		return err
	}
}
