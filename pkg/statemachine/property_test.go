package statemachine

import (
	"testing"
	"testing/quick"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

func TestPropertyStateMachineTerminalImmutability(t *testing.T) {
	engine := NewEngine()
	terminalStates := []core.RunState{
		core.RunStateCompleted,
		core.RunStateFailed,
		core.RunStateCancelled,
	}

	property := func(targetStateIndex uint8) bool {
		targetStates := []core.RunState{
			core.RunStatePending,
			core.RunStateRunning,
			core.RunStateSuspended,
			core.RunStateCompleted,
			core.RunStateFailed,
			core.RunStateCancelled,
		}
		target := targetStates[int(targetStateIndex)%len(targetStates)]

		for _, term := range terminalStates {
			// Invariant: No transition out of a terminal state is ever permitted
			if engine.CanTransitionWorkflow(term, target) {
				return false
			}
		}
		return true
	}

	if err := quick.Check(property, nil); err != nil {
		t.Fatalf("property test failed: %v", err)
	}
}

func TestPropertyStepStateTerminalImmutability(t *testing.T) {
	engine := NewEngine()
	terminalStepStates := []core.StepState{
		core.StepStateCompleted,
		core.StepStateFailed,
		core.StepStateSkipped,
		core.StepStateCancelled,
	}

	allStepStates := []core.StepState{
		core.StepStatePending,
		core.StepStateQueued,
		core.StepStateRunning,
		core.StepStateRetrying,
		core.StepStateCompleted,
		core.StepStateFailed,
		core.StepStateSkipped,
		core.StepStateCancelled,
	}

	property := func(idx uint8) bool {
		target := allStepStates[int(idx)%len(allStepStates)]
		for _, term := range terminalStepStates {
			if engine.CanTransitionStep(term, target) {
				return false
			}
		}
		return true
	}

	if err := quick.Check(property, nil); err != nil {
		t.Fatalf("step state terminal immutability property failed: %v", err)
	}
}
