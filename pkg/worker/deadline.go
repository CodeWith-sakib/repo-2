package worker

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrDeadlineExceeded = errors.New("task execution deadline exceeded")
)

type ExecutionBudget struct {
	TotalBudget   time.Duration
	StepTimeout   time.Duration
	WarnThreshold time.Duration
}

type DeadlineCoordinator struct {
	mu            sync.Mutex
	activeBudgets map[string]*taskTimer
}

type taskTimer struct {
	cancel context.CancelFunc
	budget ExecutionBudget
	start  time.Time
}

func NewDeadlineCoordinator() *DeadlineCoordinator {
	return &DeadlineCoordinator{
		activeBudgets: make(map[string]*taskTimer),
	}
}

func (dc *DeadlineCoordinator) WrapContext(parent context.Context, taskID string, budget ExecutionBudget) (context.Context, context.CancelFunc) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	timeout := budget.StepTimeout
	if timeout <= 0 {
		timeout = budget.TotalBudget
	}
	if timeout <= 0 {
		timeout = 5 * time.Minute
	}

	ctx, cancel := context.WithTimeout(parent, timeout)
	timer := &taskTimer{
		cancel: cancel,
		budget: budget,
		start:  time.Now().UTC(),
	}
	dc.activeBudgets[taskID] = timer

	cleanup := func() {
		cancel()
		dc.mu.Lock()
		delete(dc.activeBudgets, taskID)
		dc.mu.Unlock()
	}

	return ctx, cleanup
}

func (dc *DeadlineCoordinator) RemainingTime(taskID string) (time.Duration, bool) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	timer, exists := dc.activeBudgets[taskID]
	if !exists {
		return 0, false
	}

	timeout := timer.budget.StepTimeout
	if timeout <= 0 {
		timeout = timer.budget.TotalBudget
	}
	elapsed := time.Since(timer.start)
	if elapsed >= timeout {
		return 0, true
	}
	return timeout - elapsed, true
}
