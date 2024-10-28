package grpc

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/scheduler"
	"github.com/kestrelflow/kestrelflow/pkg/statemachine"
	"github.com/kestrelflow/kestrelflow/pkg/storage"
)

type SubmitRunReq struct {
	WorkflowID string
	Version    int
	Priority   int
	InputJSON  []byte
}

type RunStatusResp struct {
	RunID        string
	WorkflowID   string
	State        string
	ErrorMessage string
}

type CancelRunReq struct {
	RunID  string
	Reason string
}

type Service struct {
	store     storage.EngineStore
	scheduler *scheduler.Scheduler
	sm        *statemachine.Engine
}

func NewService(store storage.EngineStore, sched *scheduler.Scheduler) *Service {
	return &Service{
		store:     store,
		scheduler: sched,
		sm:        statemachine.NewEngine(),
	}
}

func (s *Service) SubmitWorkflowRun(ctx context.Context, req SubmitRunReq) (*RunStatusResp, error) {
	if req.WorkflowID == "" {
		return nil, fmt.Errorf("%w: missing workflow id", core.ErrValidationFailed)
	}

	wf, err := s.store.GetWorkflow(ctx, core.ID(req.WorkflowID))
	if err != nil {
		return nil, err
	}

	priority := core.PriorityNormal
	if req.Priority > 0 {
		priority = core.Priority(req.Priority)
	}

	run, err := s.scheduler.SubmitRun(ctx, wf, json.RawMessage(req.InputJSON), priority)
	if err != nil {
		return nil, err
	}

	return &RunStatusResp{
		RunID:      string(run.ID),
		WorkflowID: string(run.WorkflowID),
		State:      string(run.State),
	}, nil
}

func (s *Service) GetWorkflowRun(ctx context.Context, runID string) (*RunStatusResp, error) {
	if runID == "" {
		return nil, fmt.Errorf("%w: missing run id", core.ErrValidationFailed)
	}

	run, err := s.store.GetRun(ctx, core.ID(runID))
	if err != nil {
		return nil, err
	}

	return &RunStatusResp{
		RunID:        string(run.ID),
		WorkflowID:   string(run.WorkflowID),
		State:        string(run.State),
		ErrorMessage: run.ErrorMessage,
	}, nil
}

func (s *Service) CancelWorkflowRun(ctx context.Context, req CancelRunReq) (*RunStatusResp, error) {
	if req.RunID == "" {
		return nil, fmt.Errorf("%w: missing run id", core.ErrValidationFailed)
	}

	run, err := s.store.GetRun(ctx, core.ID(req.RunID))
	if err != nil {
		return nil, err
	}

	if run.State.IsTerminal() {
		return nil, fmt.Errorf("%w: run already terminal", core.ErrInvalidStateTransition)
	}

	if err := s.sm.TransitionWorkflow(run, core.RunStateCancelled); err != nil {
		return nil, err
	}

	_ = s.store.UpdateRun(ctx, run)
	return &RunStatusResp{
		RunID:      string(run.ID),
		WorkflowID: string(run.WorkflowID),
		State:      string(run.State),
	}, nil
}
