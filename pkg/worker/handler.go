package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

type StepContext struct {
	RunID    string
	StepID   string
	Attempt  int
	WorkerID string
	Input    json.RawMessage
}

type StepResult struct {
	Output       json.RawMessage
	ErrorMessage string
	Retryable    bool
}

type TaskExecutor interface {
	Execute(ctx context.Context, sctx StepContext) (*StepResult, error)
}

type ExecutorRegistry struct {
	mu        sync.RWMutex
	executors map[string]TaskExecutor
}

func NewExecutorRegistry() *ExecutorRegistry {
	return &ExecutorRegistry{
		executors: make(map[string]TaskExecutor),
	}
}

func (r *ExecutorRegistry) Register(taskType string, exec TaskExecutor) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.executors[taskType] = exec
}

func (r *ExecutorRegistry) Get(taskType string) (TaskExecutor, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	exec, exists := r.executors[taskType]
	if !exists {
		return nil, fmt.Errorf("no executor registered for task type: %s", taskType)
	}
	return exec, nil
}
