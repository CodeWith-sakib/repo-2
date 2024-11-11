package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

type TaskHandler interface {
	Type() string
	Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error)
	ValidateConfig(config json.RawMessage) error
}

type Registry struct {
	mu       sync.RWMutex
	handlers map[string]TaskHandler
}

func NewRegistry() *Registry {
	return &Registry{
		handlers: make(map[string]TaskHandler),
	}
}

func (r *Registry) Register(h TaskHandler) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	taskType := h.Type()
	if taskType == "" {
		return fmt.Errorf("plugin task type cannot be empty")
	}
	if _, exists := r.handlers[taskType]; exists {
		return fmt.Errorf("plugin already registered for type: %s", taskType)
	}

	r.handlers[taskType] = h
	return nil
}

func (r *Registry) Get(taskType string) (TaskHandler, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	h, exists := r.handlers[taskType]
	if !exists {
		return nil, fmt.Errorf("no plugin registered for task type: %s", taskType)
	}
	return h, nil
}

func (r *Registry) PopulateWorkerRegistry(wreg *worker.ExecutorRegistry) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for k, h := range r.handlers {
		wreg.Register(k, h)
	}
}
