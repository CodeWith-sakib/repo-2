package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/statemachine"
	"github.com/kestrelflow/kestrelflow/pkg/storage"
)

type Scheduler struct {
	mu           sync.Mutex
	store        storage.EngineStore
	sm           *statemachine.Engine
	pollInterval time.Duration
	stopCh       chan struct{}
}

func NewScheduler(store storage.EngineStore, pollInterval time.Duration) *Scheduler {
	if pollInterval <= 0 {
		pollInterval = 50 * time.Millisecond
	}
	return &Scheduler{
		store:        store,
		sm:           statemachine.NewEngine(),
		pollInterval: pollInterval,
		stopCh:       make(chan struct{}),
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.reconcile(ctx)
		}
	}
}

func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	select {
	case <-s.stopCh:
	default:
		close(s.stopCh)
	}
}

func (s *Scheduler) reconcile(ctx context.Context) {
	// Requeue any orphaned tasks where worker lease expired
	_, _ = s.store.RequeueOrphaned(ctx)
}

func (s *Scheduler) SubmitRun(ctx context.Context, wf *core.WorkflowDefinition, input json.RawMessage, priority core.Priority) (*core.WorkflowRun, error) {
	if err := wf.Validate(); err != nil {
		return nil, fmt.Errorf("workflow validation failed: %w", err)
	}

	dag, err := core.BuildDAG(wf.Steps)
	if err != nil {
		return nil, err
	}

	runID := core.NewID("run")
	now := time.Now().UTC()

	run := &core.WorkflowRun{
		ID:         runID,
		WorkflowID: wf.ID,
		Version:    wf.Version,
		TenantID:   wf.TenantID,
		State:      core.RunStatePending,
		Input:      input,
		Priority:   priority,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := s.store.CreateRun(ctx, run); err != nil {
		return nil, fmt.Errorf("failed creating run: %w", err)
	}

	// Create step runs
	for _, step := range wf.Steps {
		sr := &core.StepRun{
			ID:        core.NewID("steprun"),
			RunID:     runID,
			StepID:    step.ID,
			State:     core.StepStatePending,
			Attempt:   0,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := s.store.CreateStepRun(ctx, sr); err != nil {
			return nil, fmt.Errorf("failed creating step run %s: %w", step.ID, err)
		}
	}

	// Transition run to RUNNING
	if err := s.sm.TransitionWorkflow(run, core.RunStateRunning); err != nil {
		return nil, err
	}
	if err := s.store.UpdateRun(ctx, run); err != nil {
		return nil, err
	}

	_ = s.store.AppendEvent(ctx, &core.Event{
		Type:      core.EventRunStarted,
		TenantID:  wf.TenantID,
		RunID:     run.ID,
		Timestamp: now,
	})

	// Enqueue all root nodes (steps with no dependencies)
	roots := dag.RootNodes()
	for _, rootID := range roots {
		stepDef, _ := dag.GetNode(rootID)
		if err := s.enqueueStep(ctx, run, &stepDef); err != nil {
			return nil, err
		}
	}

	return run, nil
}

func (s *Scheduler) enqueueStep(ctx context.Context, run *core.WorkflowRun, step *core.StepDefinition) error {
	sr, err := s.store.GetStepRunByStepID(ctx, run.ID, step.ID)
	if err != nil {
		return err
	}

	if err := s.sm.TransitionStep(sr, core.StepStateQueued); err != nil {
		return err
	}
	sr.Attempt++
	if err := s.store.UpdateStepRun(ctx, sr); err != nil {
		return err
	}

	task := &storage.QueuedTask{
		ID:          core.NewID("task"),
		RunID:       run.ID,
		StepID:      step.ID,
		TenantID:    run.TenantID,
		Priority:    run.Priority,
		Attempt:     sr.Attempt,
		ScheduledAt: time.Now().UTC(),
	}

	_ = s.store.AppendEvent(ctx, &core.Event{
		Type:      core.EventStepScheduled,
		TenantID:  run.TenantID,
		RunID:     run.ID,
		StepID:    step.ID,
		Timestamp: time.Now().UTC(),
	})

	return s.store.EnqueueTask(ctx, task)
}


func (s *Scheduler) HandleStepStarted(ctx context.Context, runID core.ID, stepID string, workerID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sr, err := s.store.GetStepRunByStepID(ctx, runID, stepID)
	if err != nil {
		return err
	}

	if err := s.sm.TransitionStep(sr, core.StepStateRunning); err != nil {
		return err
	}
	sr.WorkerID = workerID
	now := time.Now().UTC()
	sr.StartedAt = &now
	if err := s.store.UpdateStepRun(ctx, sr); err != nil {
		return err
	}

	_ = s.store.AppendEvent(ctx, &core.Event{
		Type:      core.EventStepStarted,
		TenantID:  sr.StepID,
		RunID:     runID,
		StepID:    stepID,
		Timestamp: now,
	})
	return nil
}

func (s *Scheduler) HandleStepCompleted(ctx context.Context, runID core.ID, stepID string, output json.RawMessage) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	run, err := s.store.GetRun(ctx, runID)
	if err != nil {
		return err
	}
	if run.State.IsTerminal() {
		return nil // Run already terminated
	}

	wf, err := s.store.GetWorkflowVersion(ctx, run.WorkflowID, run.Version)
	if err != nil {
		return err
	}

	sr, err := s.store.GetStepRunByStepID(ctx, runID, stepID)
	if err != nil {
		return err
	}

	if err := s.sm.TransitionStep(sr, core.StepStateCompleted); err != nil {
		return err
	}
	sr.Output = output
	if err := s.store.UpdateStepRun(ctx, sr); err != nil {
		return err
	}

	_ = s.store.AppendEvent(ctx, &core.Event{
		Type:      core.EventStepCompleted,
		TenantID:  run.TenantID,
		RunID:     run.ID,
		StepID:    stepID,
		Timestamp: time.Now().UTC(),
	})

	// Check if downstream steps are now unblocked
	dag, err := core.BuildDAG(wf.Steps)
	if err != nil {
		return err
	}

	allSteps, err := s.store.ListStepRuns(ctx, storage.StepRunFilter{RunID: runID})
	if err != nil {
		return err
	}

	completedMap := make(map[string]bool)
	allCompleted := true
	for _, stepRun := range allSteps {
		if stepRun.State == core.StepStateCompleted {
			completedMap[stepRun.StepID] = true
		} else if !stepRun.State.IsTerminal() {
			allCompleted = false
		}
	}

	// Find dependents of this step
	dependents := dag.GetDependents(stepID)
	for _, depID := range dependents {
		depStepRun, err := s.store.GetStepRunByStepID(ctx, runID, depID)
		if err != nil || depStepRun.State != core.StepStatePending {
			continue
		}

		// Verify all prerequisites of depID are completed
		prereqs := dag.GetDependencies(depID)
		canRun := true
		for _, p := range prereqs {
			if !completedMap[p] {
				canRun = false
				break
			}
		}

		if canRun {
			depStepDef, _ := dag.GetNode(depID)
			_ = s.enqueueStep(ctx, run, &depStepDef)
		}
	}

	// If all steps in the entire workflow have completed successfully
	if allCompleted && len(completedMap) == len(wf.Steps) {
		if err := s.sm.TransitionWorkflow(run, core.RunStateCompleted); err != nil {
			return err
		}
		_ = s.store.UpdateRun(ctx, run)
		_ = s.store.AppendEvent(ctx, &core.Event{
			Type:      core.EventRunCompleted,
			TenantID:  run.TenantID,
			RunID:     run.ID,
			Timestamp: time.Now().UTC(),
		})
	}

	return nil
}

func (s *Scheduler) HandleStepFailed(ctx context.Context, runID core.ID, stepID string, errMsg string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	run, err := s.store.GetRun(ctx, runID)
	if err != nil {
		return err
	}
	if run.State.IsTerminal() {
		return nil
	}

	sr, err := s.store.GetStepRunByStepID(ctx, runID, stepID)
	if err != nil {
		return err
	}

	wf, err := s.store.GetWorkflowVersion(ctx, run.WorkflowID, run.Version)
	if err != nil {
		return err
	}

	dag, err := core.BuildDAG(wf.Steps)
	if err != nil {
		return err
	}
	stepDef, _ := dag.GetNode(stepID)

	// Check retry policy
	maxAttempts := 1
	if stepDef.RetryPolicy != nil && stepDef.RetryPolicy.MaxAttempts > 1 {
		maxAttempts = stepDef.RetryPolicy.MaxAttempts
	}

	if sr.Attempt < maxAttempts {
		if err := s.sm.TransitionStep(sr, core.StepStateRetrying); err != nil {
			return err
		}
		sr.ErrorMessage = errMsg
		_ = s.store.UpdateStepRun(ctx, sr)

		_ = s.store.AppendEvent(ctx, &core.Event{
			Type:      core.EventStepRetrying,
			TenantID:  run.TenantID,
			RunID:     run.ID,
			StepID:    stepID,
			Timestamp: time.Now().UTC(),
		})

		// Re-enqueue after backoff
		return s.enqueueStep(ctx, run, &stepDef)
	}

	// Retries exhausted: mark step failed
	if err := s.sm.TransitionStep(sr, core.StepStateFailed); err != nil {
		return err
	}
	sr.ErrorMessage = errMsg
	_ = s.store.UpdateStepRun(ctx, sr)

	_ = s.store.AppendEvent(ctx, &core.Event{
		Type:      core.EventStepFailed,
		TenantID:  run.TenantID,
		RunID:     run.ID,
		StepID:    stepID,
		Timestamp: time.Now().UTC(),
	})

	// Fail the workflow run and skip unstarted dependent steps
	if err := s.sm.TransitionWorkflow(run, core.RunStateFailed); err != nil {
		return err
	}
	run.ErrorMessage = fmt.Sprintf("step %s failed: %s", stepID, errMsg)
	_ = s.store.UpdateRun(ctx, run)

	_ = s.store.AppendEvent(ctx, &core.Event{
		Type:      core.EventRunFailed,
		TenantID:  run.TenantID,
		RunID:     run.ID,
		Timestamp: time.Now().UTC(),
	})

	// Mark all non-terminal steps as SKIPPED
	allSteps, _ := s.store.ListStepRuns(ctx, storage.StepRunFilter{RunID: runID})
	for _, otherStep := range allSteps {
		if otherStep.State == core.StepStatePending || otherStep.State == core.StepStateQueued {
			_ = s.sm.TransitionStep(otherStep, core.StepStateSkipped)
			_ = s.store.UpdateStepRun(ctx, otherStep)
		}
	}

	return nil
}
