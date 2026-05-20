package core

import (
	"context"
	"fmt"
	"sync"
)

// CompensatingAction defines a rollback step to execute if subsequent workflow steps fail.
type CompensatingAction struct {
	StepID     string
	ActionName string
	Payload    map[string]interface{}
}

// CompensatingActionHandler defines the signature for executing a compensation.
type CompensatingActionHandler func(ctx context.Context, action CompensatingAction) error

// CompensatingActionEngine manages LIFO rollback chains (Saga pattern) for transactional workflows.
type CompensatingActionEngine struct {
	mu      sync.Mutex
	actions []CompensatingAction
}

// NewCompensatingActionEngine creates a saga rollback coordinator.
func NewCompensatingActionEngine() *CompensatingActionEngine {
	return &CompensatingActionEngine{
		actions: make([]CompensatingAction, 0, 10),
	}
}

// RegisterAction pushes a completed step's undo compensation onto the rollback stack.
func (e *CompensatingActionEngine) RegisterAction(action CompensatingAction) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.actions = append(e.actions, action)
}

// Rollback executes all registered compensating actions in reverse order (LIFO).
func (e *CompensatingActionEngine) Rollback(ctx context.Context, handler CompensatingActionHandler) ([]string, error) {
	e.mu.Lock()
	actions := make([]CompensatingAction, len(e.actions))
	copy(actions, e.actions)
	e.mu.Unlock()

	if handler == nil {
		return nil, fmt.Errorf("compensation handler cannot be nil")
	}

	var executed []string
	// Execute in reverse order
	for i := len(actions) - 1; i >= 0; i-- {
		act := actions[i]
		if err := handler(ctx, act); err != nil {
			return executed, fmt.Errorf("failed executing compensation for step %s: %w", act.StepID, err)
		}
		executed = append(executed, act.StepID)
	}

	return executed, nil
}

// ActionCount returns total registered compensating actions in the stack.
func (e *CompensatingActionEngine) ActionCount() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return len(e.actions)
}
