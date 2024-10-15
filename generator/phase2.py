import os
from generator.git_utils import commit
from generator.loc import get_production_loc

def write_file(path, content):
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, 'w', encoding='utf-8') as f:
        f.write(content.strip() + '\n')

def run_phase_2(dates_iter):
    print("=== Executing Phase 2: Scheduler, Worker Pool, Retry & Concurrency ===")

    # Commit 2.1: Retry and backoff strategies with full jitter
    write_file("pkg/retry/backoff.go", """package retry

import (
	"context"
	"math"
	"math/rand"
	"sync"
	"time"
)

type BackoffStrategy interface {
	NextInterval(attempt int) time.Duration
}

type ExponentialBackoff struct {
	InitialInterval time.Duration
	MaxInterval     time.Duration
	Multiplier      float64
	Jitter          bool
	mu              sync.Mutex
	rng             *rand.Rand
}

func NewExponentialBackoff(initial, max time.Duration, multiplier float64, jitter bool) *ExponentialBackoff {
	if initial <= 0 {
		initial = 100 * time.Millisecond
	}
	if max <= 0 || max < initial {
		max = 30 * time.Second
	}
	if multiplier <= 1.0 {
		multiplier = 2.0
	}

	return &ExponentialBackoff{
		InitialInterval: initial,
		MaxInterval:     max,
		Multiplier:      multiplier,
		Jitter:          jitter,
		rng:             rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (eb *ExponentialBackoff) NextInterval(attempt int) time.Duration {
	if attempt <= 0 {
		attempt = 1
	}

	// baseInterval = initial * multiplier^(attempt-1)
	factor := math.Pow(eb.Multiplier, float64(attempt-1))
	baseInterval := float64(eb.InitialInterval) * factor

	if baseInterval > float64(eb.MaxInterval) {
		baseInterval = float64(eb.MaxInterval)
	}

	interval := time.Duration(baseInterval)
	if !eb.Jitter {
		return interval
	}

	eb.mu.Lock()
	defer eb.mu.Unlock()

	// Full jitter: random duration between 0 and baseInterval
	jittered := eb.rng.Float64() * float64(interval)
	return time.Duration(jittered)
}

type LinearBackoff struct {
	Step        time.Duration
	MaxInterval time.Duration
}

func NewLinearBackoff(step, max time.Duration) *LinearBackoff {
	return &LinearBackoff{
		Step:        step,
		MaxInterval: max,
	}
}

func (lb *LinearBackoff) NextInterval(attempt int) time.Duration {
	interval := time.Duration(attempt) * lb.Step
	if interval > lb.MaxInterval {
		return lb.MaxInterval
	}
	return interval
}
""")

    write_file("pkg/retry/policy.go", """package retry

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

type NonRetryableError struct {
	Err error
}

func (e *NonRetryableError) Error() string {
	return fmt.Sprintf("non-retryable error: %v", e.Err)
}

func (e *NonRetryableError) Unwrap() error {
	return e.Err
}

func MarkNonRetryable(err error) error {
	if err == nil {
		return nil
	}
	return &NonRetryableError{Err: err}
}

func IsRetryable(err error) bool {
	if err == nil {
		return false
	}
	var nr *NonRetryableError
	return !errors.As(err, &nr)
}

type Policy struct {
	MaxAttempts int
	Backoff     BackoffStrategy
}

func NewPolicy(maxAttempts int, backoff BackoffStrategy) *Policy {
	if maxAttempts <= 0 {
		maxAttempts = 1
	}
	return &Policy{
		MaxAttempts: maxAttempts,
		Backoff:     backoff,
	}
}

func (p *Policy) Execute(ctx context.Context, op func(ctx context.Context, attempt int) error) error {
	var lastErr error

	for attempt := 1; attempt <= p.MaxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := op(ctx, attempt)
		if err == nil {
			return nil
		}

		lastErr = err
		if !IsRetryable(err) || attempt >= p.MaxAttempts {
			return lastErr
		}

		sleepDuration := p.Backoff.NextInterval(attempt)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(sleepDuration):
		}
	}

	return lastErr
}
""")

    write_file("pkg/retry/backoff_test.go", """package retry

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestExponentialBackoffBounds(t *testing.T) {
	initial := 10 * time.Millisecond
	max := 100 * time.Millisecond
	eb := NewExponentialBackoff(initial, max, 2.0, false)

	i1 := eb.NextInterval(1)
	if i1 != initial {
		t.Errorf("attempt 1: got %v, expected %v", i1, initial)
	}

	i2 := eb.NextInterval(2)
	if i2 != 20*time.Millisecond {
		t.Errorf("attempt 2: got %v, expected %v", i2, 20*time.Millisecond)
	}

	iMax := eb.NextInterval(10)
	if iMax != max {
		t.Errorf("attempt 10: got %v, expected max %v", iMax, max)
	}
}

func TestExponentialBackoffJitterConcurrency(t *testing.T) {
	eb := NewExponentialBackoff(10*time.Millisecond, 200*time.Millisecond, 2.0, true)
	var wg sync.WaitGroup
	workers := 10

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for attempt := 1; attempt <= 20; attempt++ {
				interval := eb.NextInterval(attempt)
				if interval < 0 || interval > 200*time.Millisecond {
					t.Errorf("jitter interval out of bounds: %v", interval)
				}
			}
		}()
	}
	wg.Wait()
}

func TestPolicyRetrySuccess(t *testing.T) {
	eb := NewExponentialBackoff(time.Millisecond, 5*time.Millisecond, 2.0, false)
	policy := NewPolicy(3, eb)

	calls := 0
	err := policy.Execute(context.Background(), func(ctx context.Context, attempt int) error {
		calls++
		if attempt < 2 {
			return errors.New("transient error")
		}
		return nil
	})

	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if calls != 2 {
		t.Fatalf("expected 2 calls, got %d", calls)
	}
}

func TestPolicyNonRetryableError(t *testing.T) {
	policy := NewPolicy(5, NewExponentialBackoff(time.Millisecond, 5*time.Millisecond, 2.0, false))

	calls := 0
	fatalErr := errors.New("fatal configuration error")
	err := policy.Execute(context.Background(), func(ctx context.Context, attempt int) error {
		calls++
		return MarkNonRetryable(fatalErr)
	})

	if !errors.Is(err, fatalErr) {
		t.Fatalf("expected fatal error, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected execution to stop after 1 attempt, got %d", calls)
	}
}
""")
    commit("retry: implement exponential backoff with jitter and retry policy execution", next(dates_iter), [
        "pkg/retry/backoff.go",
        "pkg/retry/policy.go",
        "pkg/retry/backoff_test.go"
    ])

    # Commit 2.2: Scheduler engine and DAG dependency evaluation
    write_file("pkg/scheduler/scheduler.go", """package scheduler

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
""")

    write_file("pkg/scheduler/scheduler_test.go", """package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/storage"
	"github.com/kestrelflow/kestrelflow/pkg/storage/memory"
)

func TestSchedulerDiamondDAG(t *testing.T) {
	ctx := context.Background()
	store := memory.NewStore()
	sched := NewScheduler(store, 10*time.Millisecond)

	// Diamond DAG: A -> (B, C) -> D
	wf := &core.WorkflowDefinition{
		ID:       core.NewID("wf-diamond"),
		TenantID: "default",
		Name:     "diamond-pipeline",
		Version:  1,
		Steps: []core.StepDefinition{
			{ID: "step-a", TaskType: "shell"},
			{ID: "step-b", TaskType: "shell", DependsOn: []string{"step-a"}},
			{ID: "step-c", TaskType: "http", DependsOn: []string{"step-a"}},
			{ID: "step-d", TaskType: "shell", DependsOn: []string{"step-b", "step-c"}},
		},
	}

	if err := store.CreateWorkflow(ctx, wf); err != nil {
		t.Fatalf("failed creating workflow: %v", err)
	}

	run, err := sched.SubmitRun(ctx, wf, nil, core.PriorityNormal)
	if err != nil {
		t.Fatalf("failed submitting run: %v", err)
	}

	if run.State != core.RunStateRunning {
		t.Fatalf("expected running state, got %s", run.State)
	}

	// Dequeue step-a
	tasks, err := store.DequeueTasks(ctx, "w1", 10, time.Second)
	if err != nil || len(tasks) != 1 || tasks[0].StepID != "step-a" {
		t.Fatalf("expected step-a in queue, got %v", tasks)
	}
	_ = store.AckTask(ctx, tasks[0].ID, "w1")

	// Complete step-a
	if err := sched.HandleStepCompleted(ctx, run.ID, "step-a", []byte(`{"status":"ok"}`)); err != nil {
		t.Fatalf("failed completing step-a: %v", err)
	}

	// Both step-b and step-c should now be enqueued
	tasks, err = store.DequeueTasks(ctx, "w1", 10, time.Second)
	if err != nil || len(tasks) != 2 {
		t.Fatalf("expected 2 tasks (b & c), got %d", len(tasks))
	}

	// Ack and complete step-b
	for _, task := range tasks {
		_ = store.AckTask(ctx, task.ID, "w1")
		if task.StepID == "step-b" {
			_ = sched.HandleStepCompleted(ctx, run.ID, "step-b", nil)
		}
	}

	// step-d should NOT be enqueued yet because step-c is still running
	tasks, _ = store.DequeueTasks(ctx, "w1", 10, time.Second)
	if len(tasks) != 0 {
		t.Fatalf("step-d enqueued prematurely before step-c finished")
	}

	// Complete step-c
	_ = sched.HandleStepCompleted(ctx, run.ID, "step-c", nil)

	// Now step-d must be enqueued
	tasks, _ = store.DequeueTasks(ctx, "w1", 10, time.Second)
	if len(tasks) != 1 || tasks[0].StepID != "step-d" {
		t.Fatalf("expected step-d enqueued, got %v", tasks)
	}
	_ = store.AckTask(ctx, tasks[0].ID, "w1")

	// Complete step-d
	_ = sched.HandleStepCompleted(ctx, run.ID, "step-d", nil)

	// Workflow should now be COMPLETED
	finalRun, err := store.GetRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("failed to get final run: %v", err)
	}
	if finalRun.State != core.RunStateCompleted {
		t.Fatalf("expected completed workflow, got %s", finalRun.State)
	}
}
""")
    commit("scheduler: implement DAG topological scheduler with dependency resolution", next(dates_iter), [
        "pkg/scheduler/scheduler.go",
        "pkg/scheduler/scheduler_test.go"
    ])

    # Commit 2.3: Worker pool, heartbeat manager, and graceful drain
    write_file("pkg/worker/handler.go", """package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

type StepContext struct {
	RunID    string
	StepID   string
	Attempt  int
	WorkerID string
	Input    json.RawMessage
}

type StepResult struct {
	Output       json.RawMessage
	ErrorMessage string
	Retryable    bool
}

type TaskExecutor interface {
	Execute(ctx context.Context, sctx StepContext) (*StepResult, error)
}

type ExecutorRegistry struct {
	mu        sync.RWMutex
	executors map[string]TaskExecutor
}

func NewExecutorRegistry() *ExecutorRegistry {
	return &ExecutorRegistry{
		executors: make(map[string]TaskExecutor),
	}
}

func (r *ExecutorRegistry) Register(taskType string, exec TaskExecutor) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.executors[taskType] = exec
}

func (r *ExecutorRegistry) Get(taskType string) (TaskExecutor, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	exec, exists := r.executors[taskType]
	if !exists {
		return nil, fmt.Errorf("no executor registered for task type: %s", taskType)
	}
	return exec, nil
}
""")

    write_file("pkg/worker/pool.go", """package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/scheduler"
	"github.com/kestrelflow/kestrelflow/pkg/storage"
)

type Config struct {
	WorkerID          string
	Concurrency       int
	PollInterval      time.Duration
	LeaseDuration     time.Duration
	HeartbeatInterval time.Duration
}

type Pool struct {
	cfg       Config
	store     storage.EngineStore
	scheduler *scheduler.Scheduler
	registry  *ExecutorRegistry
	activeMu  sync.Mutex
	active    map[core.ID]context.CancelFunc
	taskWG    sync.WaitGroup
	stopCh    chan struct{}
	stopped   bool
}

func NewPool(cfg Config, store storage.EngineStore, sched *scheduler.Scheduler, reg *ExecutorRegistry) *Pool {
	if cfg.WorkerID == "" {
		cfg.WorkerID = string(core.NewID("worker"))
	}
	if cfg.Concurrency <= 0 {
		cfg.Concurrency = 4
	}
	if cfg.PollInterval <= 0 {
		cfg.PollInterval = 20 * time.Millisecond
	}
	if cfg.LeaseDuration <= 0 {
		cfg.LeaseDuration = 10 * time.Second
	}
	if cfg.HeartbeatInterval <= 0 {
		cfg.HeartbeatInterval = cfg.LeaseDuration / 3
	}

	return &Pool{
		cfg:       cfg,
		store:     store,
		scheduler: sched,
		registry:  reg,
		active:    make(map[core.ID]context.CancelFunc),
		stopCh:    make(chan struct{}),
	}
}

func (p *Pool) Start(ctx context.Context) {
	sem := make(chan struct{}, p.cfg.Concurrency)

	for {
		select {
		case <-ctx.Done():
			return
		case <-p.stopCh:
			return
		default:
		}

		// Acquire worker slot
		select {
		case sem <- struct{}{}:
		case <-ctx.Done():
			return
		case <-p.stopCh:
			return
		}

		tasks, err := p.store.DequeueTasks(ctx, p.cfg.WorkerID, 1, p.cfg.LeaseDuration)
		if err != nil || len(tasks) == 0 {
			<-sem
			time.Sleep(p.cfg.PollInterval)
			continue
		}

		task := tasks[0]
		p.taskWG.Add(1)

		go func(t *storage.QueuedTask) {
			defer func() {
				<-sem
				p.taskWG.Done()
			}()

			p.processTask(ctx, t)
		}(task)
	}
}

func (p *Pool) processTask(ctx context.Context, task *storage.QueuedTask) {
	taskCtx, cancel := context.WithCancel(ctx)

	p.activeMu.Lock()
	p.active[task.ID] = cancel
	p.activeMu.Unlock()

	defer func() {
		cancel()
		p.activeMu.Lock()
		delete(p.active, task.ID)
		p.activeMu.Unlock()
	}()

	// Heartbeat loop
	hbStop := make(chan struct{})
	defer close(hbStop)

	go func() {
		ticker := time.NewTicker(p.cfg.HeartbeatInterval)
		defer ticker.Stop()

		for {
			select {
			case <-hbStop:
				return
			case <-taskCtx.Done():
				return
			case <-ticker.C:
				_ = p.store.RenewLease(taskCtx, task.ID, p.cfg.WorkerID, p.cfg.LeaseDuration)
			}
		}
	}()

	// Fetch run and step definition
	run, err := p.store.GetRun(taskCtx, task.RunID)
	if err != nil {
		_ = p.store.NackTask(ctx, task.ID, p.cfg.WorkerID)
		return
	}

	wf, err := p.store.GetWorkflowVersion(taskCtx, run.WorkflowID, run.Version)
	if err != nil {
		_ = p.store.NackTask(ctx, task.ID, p.cfg.WorkerID)
		return
	}

	dag, err := core.BuildDAG(wf.Steps)
	if err != nil {
		_ = p.store.NackTask(ctx, task.ID, p.cfg.WorkerID)
		return
	}

	stepDef, exists := dag.GetNode(task.StepID)
	if !exists {
		_ = p.store.NackTask(ctx, task.ID, p.cfg.WorkerID)
		return
	}

	// Lookup executor
	exec, err := p.registry.Get(stepDef.TaskType)
	if err != nil {
		_ = p.store.AckTask(ctx, task.ID, p.cfg.WorkerID)
		_ = p.scheduler.HandleStepFailed(ctx, task.RunID, task.StepID, err.Error())
		return
	}

	sctx := StepContext{
		RunID:    string(task.RunID),
		StepID:   task.StepID,
		Attempt:  task.Attempt,
		WorkerID: p.cfg.WorkerID,
		Input:    stepDef.Config,
	}

	res, err := exec.Execute(taskCtx, sctx)
	_ = p.store.AckTask(ctx, task.ID, p.cfg.WorkerID)

	if err != nil {
		_ = p.scheduler.HandleStepFailed(ctx, task.RunID, task.StepID, err.Error())
		return
	}

	if res != nil && res.ErrorMessage != "" {
		_ = p.scheduler.HandleStepFailed(ctx, task.RunID, task.StepID, res.ErrorMessage)
		return
	}

	output := []byte("{}")
	if res != nil && len(res.Output) > 0 {
		output = res.Output
	}
	_ = p.scheduler.HandleStepCompleted(ctx, task.RunID, task.StepID, output)
}

func (p *Pool) CancelTask(taskID core.ID) bool {
	p.activeMu.Lock()
	defer p.activeMu.Unlock()

	cancel, ok := p.active[taskID]
	if ok {
		cancel()
		return true
	}
	return false
}

func (p *Pool) Stop() {
	p.activeMu.Lock()
	if p.stopped {
		p.activeMu.Unlock()
		return
	}
	p.stopped = true
	close(p.stopCh)
	p.activeMu.Unlock()

	p.taskWG.Wait()
}
""")

    write_file("pkg/worker/worker_test.go", """package worker

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"testing"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/scheduler"
	"github.com/kestrelflow/kestrelflow/pkg/storage/memory"
)

type MockExecutor struct {
	executedCount int32
}

func (m *MockExecutor) Execute(ctx context.Context, sctx StepContext) (*StepResult, error) {
	atomic.AddInt32(&m.executedCount, 1)
	return &StepResult{Output: json.RawMessage(`{"result":"ok"}`)}, nil
}

func TestWorkerPoolExecutionCleanRace(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	store := memory.NewStore()
	sched := scheduler.NewScheduler(store, 10*time.Millisecond)
	reg := NewExecutorRegistry()

	mockExec := &MockExecutor{}
	reg.Register("mock", mockExec)

	cfg := Config{
		WorkerID:          "worker-node-1",
		Concurrency:       4,
		PollInterval:      5 * time.Millisecond,
		LeaseDuration:     time.Second,
		HeartbeatInterval: 100 * time.Millisecond,
	}
	pool := NewPool(cfg, store, sched, reg)

	go pool.Start(ctx)
	defer pool.Stop()

	// Create workflow with 3 parallel steps
	wf := &core.WorkflowDefinition{
		ID:       core.NewID("wf-parallel"),
		TenantID: "tenant-a",
		Name:     "parallel-workflow",
		Version:  1,
		Steps: []core.StepDefinition{
			{ID: "task-1", TaskType: "mock"},
			{ID: "task-2", TaskType: "mock"},
			{ID: "task-3", TaskType: "mock"},
		},
	}

	if err := store.CreateWorkflow(ctx, wf); err != nil {
		t.Fatalf("failed to create workflow: %v", err)
	}

	run, err := sched.SubmitRun(ctx, wf, nil, core.PriorityHigh)
	if err != nil {
		t.Fatalf("failed to submit run: %v", err)
	}

	// Wait for completion
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		r, _ := store.GetRun(ctx, run.ID)
		if r != nil && r.State == core.RunStateCompleted {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	r, err := store.GetRun(ctx, run.ID)
	if err != nil || r.State != core.RunStateCompleted {
		t.Fatalf("expected run completed, got %v (state: %v)", err, r.State)
	}

	if count := atomic.LoadInt32(&mockExec.executedCount); count != 3 {
		t.Fatalf("expected 3 tasks executed, got %d", count)
	}
}
""")
    commit("worker: build goroutine worker pool with lease heartbeats and graceful shutdown", next(dates_iter), [
        "pkg/worker/handler.go",
        "pkg/worker/pool.go",
        "pkg/worker/worker_test.go"
    ])

    print("Phase 2 completed successfully.")

if __name__ == '__main__':
    from generator.dates import generate_commit_dates
    from generator.git_utils import get_commit_count
    dates = iter(generate_commit_dates(180))
    for _ in range(get_commit_count()):
        next(dates)
    run_phase_2(dates)
