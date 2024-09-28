package statemachine

import (
	"fmt"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

type Engine struct{}

func NewEngine() *Engine {
	return &Engine{}
}

func (e *Engine) CanTransitionWorkflow(from, to core.RunState) bool {
	allowed, ok := LegalWorkflowTransitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}

func (e *Engine) CanTransitionStep(from, to core.StepState) bool {
	allowed, ok := LegalStepTransitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}

func (e *Engine) TransitionWorkflow(run *core.WorkflowRun, to core.RunState) error {
	if !e.CanTransitionWorkflow(run.State, to) {
		return fmt.Errorf("%w: cannot transition workflow %s from %s to %s",
			core.ErrInvalidStateTransition, run.ID, run.State, to)
	}

	now := time.Now().UTC()
	run.State = to
	run.UpdatedAt = now

	if to == core.RunStateRunning && run.StartedAt == nil {
		run.StartedAt = &now
	}
	if to.IsTerminal() && run.FinishedAt == nil {
		run.FinishedAt = &now
	}
	return nil
}

func (e *Engine) TransitionStep(step *core.StepRun, to core.StepState) error {
	if !e.CanTransitionStep(step.State, to) {
		return fmt.Errorf("%w: cannot transition step %s (%s) from %s to %s",
			core.ErrInvalidStateTransition, step.ID, step.StepID, step.State, to)
	}

	now := time.Now().UTC()
	step.State = to
	step.UpdatedAt = now

	if to == core.StepStateRunning && step.StartedAt == nil {
		step.StartedAt = &now
	}
	if to.IsTerminal() && step.FinishedAt == nil {
		step.FinishedAt = &now
	}
	return nil
}
