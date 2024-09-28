package statemachine

import (
	"github.com/kestrelflow/kestrelflow/pkg/core"
)

type State = core.RunState
type StepState = core.StepState

var LegalWorkflowTransitions = map[core.RunState][]core.RunState{
	core.RunStatePending: {
		core.RunStateRunning,
		core.RunStateCancelled,
	},
	core.RunStateRunning: {
		core.RunStateCompleted,
		core.RunStateFailed,
		core.RunStateCancelled,
		core.RunStateSuspended,
	},
	core.RunStateSuspended: {
		core.RunStateRunning,
		core.RunStateCancelled,
	},
	core.RunStateCompleted: {},
	core.RunStateFailed:    {},
	core.RunStateCancelled: {},
}

var LegalStepTransitions = map[core.StepState][]core.StepState{
	core.StepStatePending: {
		core.StepStateQueued,
		core.StepStateSkipped,
		core.StepStateCancelled,
	},
	core.StepStateQueued: {
		core.StepStateRunning,
		core.StepStateCancelled,
	},
	core.StepStateRunning: {
		core.StepStateCompleted,
		core.StepStateFailed,
		core.StepStateRetrying,
		core.StepStateCancelled,
	},
	core.StepStateRetrying: {
		core.StepStateQueued,
		core.StepStateFailed,
		core.StepStateCancelled,
	},
	core.StepStateCompleted: {},
	core.StepStateFailed:    {},
	core.StepStateSkipped:   {},
	core.StepStateCancelled: {},
}
