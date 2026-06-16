package core

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrTimeBudgetExhausted = errors.New("workflow total execution time budget exhausted")
)

// WorkflowTimeBudgetController governs cumulative and per-step execution timeouts across distributed DAG runs.
type WorkflowTimeBudgetController struct {
	mu           sync.RWMutex
	workflowID   string
	totalBudget  time.Duration
	spentTime    time.Duration
	stepLimits   map[string]time.Duration
	startTime    time.Time
}

// NewWorkflowTimeBudgetController initializes a budget controller.
func NewWorkflowTimeBudgetController(workflowID string, totalBudget time.Duration) *WorkflowTimeBudgetController {
	if totalBudget <= 0 {
		totalBudget = 30 * time.Minute
	}
	return &WorkflowTimeBudgetController{
		workflowID:  workflowID,
		totalBudget: totalBudget,
		stepLimits:  make(map[string]time.Duration),
		startTime:   time.Now(),
	}
}

// SetStepLimit sets a specific max duration for a step.
func (c *WorkflowTimeBudgetController) SetStepLimit(stepID string, limit time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.stepLimits[stepID] = limit
}

// AllocateStepContext creates a child context bounded by both the remaining global budget and step-specific limit.
func (c *WorkflowTimeBudgetController) AllocateStepContext(parent context.Context, stepID string) (context.Context, context.CancelFunc, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elapsed := time.Since(c.startTime)
	if elapsed >= c.totalBudget {
		return nil, nil, ErrTimeBudgetExhausted
	}

	remainingTotal := c.totalBudget - elapsed
	timeout := remainingTotal

	if stepLimit, ok := c.stepLimits[stepID]; ok && stepLimit < timeout {
		timeout = stepLimit
	}

	ctx, cancel := context.WithTimeout(parent, timeout)
	return ctx, cancel, nil
}

// RecordStepElapsed updates the observed duration for accounting.
func (c *WorkflowTimeBudgetController) RecordStepElapsed(stepID string, d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.spentTime += d
}

// RemainingBudget returns remaining total execution time.
func (c *WorkflowTimeBudgetController) RemainingBudget() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()

	elapsed := time.Since(c.startTime)
	if elapsed >= c.totalBudget {
		return 0
	}
	return c.totalBudget - elapsed
}
