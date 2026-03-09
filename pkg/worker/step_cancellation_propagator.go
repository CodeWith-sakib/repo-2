package worker

import (
	"context"
	"sync"
)

// StepCancellationPropagator coordinates cooperative cancellation tokens across parent-child workflow steps.
type StepCancellationPropagator struct {
	mu       sync.RWMutex
	cancels  map[string]context.CancelFunc
	children map[string][]string // parentRunID -> childRunIDs
}

// NewStepCancellationPropagator creates a cancellation propagator.
func NewStepCancellationPropagator() *StepCancellationPropagator {
	return &StepCancellationPropagator{
		cancels:  make(map[string]context.CancelFunc),
		children: make(map[string][]string),
	}
}

// RegisterStep registers a cancellable context for an executing run/step ID.
func (p *StepCancellationPropagator) RegisterStep(runID, stepID string, cancel context.CancelFunc) {
	p.mu.Lock()
	defer p.mu.Unlock()

	key := runID + ":" + stepID
	p.cancels[key] = cancel
}

// LinkChildRun associates a nested child workflow run with a parent run.
func (p *StepCancellationPropagator) LinkChildRun(parentRunID, childRunID string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.children[parentRunID] = append(p.children[parentRunID], childRunID)
}

// CancelStep cancels a specific step immediately.
func (p *StepCancellationPropagator) CancelStep(runID, stepID string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	key := runID + ":" + stepID
	if cancel, exists := p.cancels[key]; exists {
		cancel()
		delete(p.cancels, key)
		return true
	}
	return false
}

// CancelCascade cancels all active steps in a run, and cascades recursively to all linked child runs.
func (p *StepCancellationPropagator) CancelCascade(runID string) int {
	p.mu.Lock()
	defer p.mu.Unlock()

	cancelled := 0
	prefix := runID + ":"

	for key, cancel := range p.cancels {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			cancel()
			delete(p.cancels, key)
			cancelled++
		}
	}

	// Cascade to children
	for _, childID := range p.children[runID] {
		childPrefix := childID + ":"
		for key, cancel := range p.cancels {
			if len(key) >= len(childPrefix) && key[:len(childPrefix)] == childPrefix {
				cancel()
				delete(p.cancels, key)
				cancelled++
			}
		}
	}

	return cancelled
}
