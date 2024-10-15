import os
import subprocess
from generator.git_utils import commit
from generator.loc import get_production_loc

def write_file(path, content):
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, 'w', encoding='utf-8') as f:
        f.write(content.strip() + '\n')

def run_phase_1(dates_iter):
    print("=== Executing Phase 1: Core Domain, State Machine & Persistence ===")

    # Commit 1.1: Core domain errors and primitive types
    write_file("pkg/core/errors.go", """package core

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound               = errors.New("entity not found")
	ErrAlreadyExists          = errors.New("entity already exists")
	ErrInvalidStateTransition = errors.New("invalid state transition")
	ErrCycleDetected          = errors.New("dependency cycle detected in workflow DAG")
	ErrValidationFailed       = errors.New("validation failed")
	ErrConflict               = errors.New("concurrent modification conflict")
	ErrLeaseExpired           = errors.New("worker lease has expired")
	ErrWorkerUnavailable      = errors.New("no matching worker available")
	ErrTaskTimeout            = errors.New("task execution timed out")
	ErrWorkflowCancelled      = errors.New("workflow run was cancelled")
	ErrInvalidConfiguration   = errors.New("invalid configuration")
	ErrUnauthorized           = errors.New("unauthorized access")
	ErrForbidden              = errors.New("forbidden operation")
	ErrRateLimited            = errors.New("rate limit exceeded")
)

type DomainError struct {
	Code    string
	Message string
	Err     error
}

func (e *DomainError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *DomainError) Unwrap() error {
	return e.Err
}

func NewDomainError(code, message string, underlying error) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
		Err:     underlying,
	}
}
""")

    write_file("pkg/core/types.go", """package core

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

type ID string

func NewID(prefix string) ID {
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	if prefix == "" {
		return ID(hex.EncodeToString(b))
	}
	return ID(fmt.Sprintf("%s-%s", prefix, hex.EncodeToString(b)))
}

func (id ID) String() string {
	return string(id)
}

func (id ID) IsEmpty() bool {
	return len(id) == 0
}

type Metadata map[string]string

func (m Metadata) Clone() Metadata {
	if m == nil {
		return nil
	}
	c := make(Metadata, len(m))
	for k, v := range m {
		c[k] = v
	}
	return c
}

type JSONPayload []byte

func (p JSONPayload) Clone() JSONPayload {
	if p == nil {
		return nil
	}
	c := make(JSONPayload, len(p))
	copy(c, p)
	return c
}

type Priority int

const (
	PriorityLow      Priority = 10
	PriorityNormal   Priority = 50
	PriorityHigh     Priority = 80
	PriorityCritical Priority = 100
)

type TimeRange struct {
	Start time.Time
	End   time.Time
}

func (tr TimeRange) Contains(t time.Time) bool {
	if !tr.Start.IsZero() && t.Before(tr.Start) {
		return false
	}
	if !tr.End.IsZero() && t.After(tr.End) {
		return false
	}
	return true
}
""")
    commit("core: implement domain errors and primitive ID/payload types", next(dates_iter), ["pkg/core/errors.go", "pkg/core/types.go"])

    # Commit 1.2: Workflow definition models and step specifications
    write_file("pkg/core/workflow.go", """package core

import (
	"encoding/json"
	"fmt"
	"time"
)

type RetryPolicy struct {
	MaxAttempts     int           `json:"max_attempts"`
	InitialInterval time.Duration `json:"initial_interval"`
	MaxInterval     time.Duration `json:"max_interval"`
	BackoffFactor   float64       `json:"backoff_factor"`
	Jitter          bool          `json:"jitter"`
}

func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxAttempts:     3,
		InitialInterval: time.Second,
		MaxInterval:     30 * time.Second,
		BackoffFactor:   2.0,
		Jitter:          true,
	}
}

type StepDefinition struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	TaskType    string          `json:"task_type"`
	DependsOn   []string        `json:"depends_on"`
	Timeout     time.Duration   `json:"timeout"`
	RetryPolicy *RetryPolicy    `json:"retry_policy,omitempty"`
	Config      json.RawMessage `json:"config,omitempty"`
	Condition   string          `json:"condition,omitempty"`
}

type WorkflowDefinition struct {
	ID          ID               `json:"id"`
	TenantID    string           `json:"tenant_id"`
	Name        string           `json:"name"`
	Version     int              `json:"version"`
	Description string           `json:"description"`
	Steps       []StepDefinition `json:"steps"`
	Timeout     time.Duration    `json:"timeout"`
	Metadata    Metadata         `json:"metadata,omitempty"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

func (w *WorkflowDefinition) Validate() error {
	if w.ID.IsEmpty() {
		return fmt.Errorf("%w: workflow ID must not be empty", ErrValidationFailed)
	}
	if w.Name == "" {
		return fmt.Errorf("%w: workflow name must not be empty", ErrValidationFailed)
	}
	if w.Version <= 0 {
		return fmt.Errorf("%w: workflow version must be positive", ErrValidationFailed)
	}
	if len(w.Steps) == 0 {
		return fmt.Errorf("%w: workflow must have at least one step", ErrValidationFailed)
	}

	stepMap := make(map[string]bool)
	for _, step := range w.Steps {
		if step.ID == "" {
			return fmt.Errorf("%w: step ID must not be empty", ErrValidationFailed)
		}
		if step.TaskType == "" {
			return fmt.Errorf("%w: step %s task type must not be empty", ErrValidationFailed, step.ID)
		}
		if stepMap[step.ID] {
			return fmt.Errorf("%w: duplicate step ID %s", ErrValidationFailed, step.ID)
		}
		stepMap[step.ID] = true
	}

	for _, step := range w.Steps {
		for _, dep := range step.DependsOn {
			if dep == step.ID {
				return fmt.Errorf("%w: step %s cannot depend on itself", ErrValidationFailed, step.ID)
			}
			if !stepMap[dep] {
				return fmt.Errorf("%w: step %s depends on undefined step %s", ErrValidationFailed, step.ID, dep)
			}
		}
	}

	// Validate DAG for cycles
	dag, err := BuildDAG(w.Steps)
	if err != nil {
		return err
	}
	if err := dag.ValidateAcyclic(); err != nil {
		return err
	}

	return nil
}
""")
    commit("core: define workflow and step schemas with structural validation", next(dates_iter), ["pkg/core/workflow.go"])

    # Commit 1.3: DAG resolution, cycle detection, topological sort
    write_file("pkg/core/dag.go", """package core

import (
	"fmt"
	"sort"
)

type DAG struct {
	nodes        map[string]StepDefinition
	adjacency    map[string][]string // stepID -> list of dependents (steps that depend on stepID)
	dependencies map[string][]string // stepID -> list of prerequisites (steps that stepID depends on)
}

func BuildDAG(steps []StepDefinition) (*DAG, error) {
	d := &DAG{
		nodes:        make(map[string]StepDefinition),
		adjacency:    make(map[string][]string),
		dependencies: make(map[string][]string),
	}

	for _, s := range steps {
		d.nodes[s.ID] = s
		d.adjacency[s.ID] = make([]string, 0)
		d.dependencies[s.ID] = make([]string, len(s.DependsOn))
		copy(d.dependencies[s.ID], s.DependsOn)
	}

	for _, s := range steps {
		for _, dep := range s.DependsOn {
			d.adjacency[dep] = append(d.adjacency[dep], s.ID)
		}
	}

	return d, nil
}

func (d *DAG) GetNode(id string) (StepDefinition, bool) {
	node, exists := d.nodes[id]
	return node, exists
}

func (d *DAG) Nodes() []StepDefinition {
	list := make([]StepDefinition, 0, len(d.nodes))
	for _, n := range d.nodes {
		list = append(list, n)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].ID < list[j].ID
	})
	return list
}

func (d *DAG) GetDependencies(id string) []string {
	deps := d.dependencies[id]
	out := make([]string, len(deps))
	copy(out, deps)
	return out
}

func (d *DAG) GetDependents(id string) []string {
	deps := d.adjacency[id]
	out := make([]string, len(deps))
	copy(out, deps)
	return out
}

func (d *DAG) ValidateAcyclic() error {
	inDegree := make(map[string]int)
	for id := range d.nodes {
		inDegree[id] = len(d.dependencies[id])
	}

	queue := make([]string, 0)
	for id, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, id)
		}
	}

	visitedCount := 0
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		visitedCount++

		for _, dep := range d.adjacency[curr] {
			inDegree[dep]--
			if inDegree[dep] == 0 {
				queue = append(queue, dep)
			}
		}
	}

	if visitedCount != len(d.nodes) {
		return fmt.Errorf("%w: visited %d nodes out of %d total", ErrCycleDetected, visitedCount, len(d.nodes))
	}
	return nil
}

func (d *DAG) TopologicalSort() ([]string, error) {
	inDegree := make(map[string]int)
	for id := range d.nodes {
		inDegree[id] = len(d.dependencies[id])
	}

	queue := make([]string, 0)
	for id, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, id)
		}
	}
	sort.Strings(queue)

	result := make([]string, 0, len(d.nodes))
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		result = append(result, curr)

		nextNodes := make([]string, 0)
		for _, dep := range d.adjacency[curr] {
			inDegree[dep]--
			if inDegree[dep] == 0 {
				nextNodes = append(nextNodes, dep)
			}
		}
		sort.Strings(nextNodes)
		queue = append(queue, nextNodes...)
	}

	if len(result) != len(d.nodes) {
		return nil, ErrCycleDetected
	}
	return result, nil
}

func (d *DAG) RootNodes() []string {
	roots := make([]string, 0)
	for id, deps := range d.dependencies {
		if len(deps) == 0 {
			roots = append(roots, id)
		}
	}
	sort.Strings(roots)
	return roots
}

func (d *DAG) LeafNodes() []string {
	leaves := make([]string, 0)
	for id, adjs := range d.adjacency {
		if len(adjs) == 0 {
			leaves = append(leaves, id)
		}
	}
	sort.Strings(leaves)
	return leaves
}
""")

    write_file("pkg/core/dag_test.go", """package core

import (
	"testing"
)

func TestDAGTopologicalSort(t *testing.T) {
	steps := []StepDefinition{
		{ID: "step-c", TaskType: "shell", DependsOn: []string{"step-a", "step-b"}},
		{ID: "step-a", TaskType: "http"},
		{ID: "step-b", TaskType: "shell", DependsOn: []string{"step-a"}},
		{ID: "step-d", TaskType: "http", DependsOn: []string{"step-c"}},
	}

	dag, err := BuildDAG(steps)
	if err != nil {
		t.Fatalf("unexpected error building DAG: %v", err)
	}

	if err := dag.ValidateAcyclic(); err != nil {
		t.Fatalf("expected acyclic, got: %v", err)
	}

	order, err := dag.TopologicalSort()
	if err != nil {
		t.Fatalf("unexpected sort error: %v", err)
	}

	expected := []string{"step-a", "step-b", "step-c", "step-d"}
	if len(order) != len(expected) {
		t.Fatalf("length mismatch: got %d, expected %d", len(order), len(expected))
	}
	for i, v := range expected {
		if order[i] != v {
			t.Errorf("at index %d: got %s, expected %s", i, order[i], v)
		}
	}
}

func TestDAGCycleDetection(t *testing.T) {
	cyclicSteps := []StepDefinition{
		{ID: "node-1", TaskType: "shell", DependsOn: []string{"node-3"}},
		{ID: "node-2", TaskType: "shell", DependsOn: []string{"node-1"}},
		{ID: "node-3", TaskType: "shell", DependsOn: []string{"node-2"}},
	}

	dag, err := BuildDAG(cyclicSteps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := dag.ValidateAcyclic(); err == nil {
		t.Fatal("expected cycle error, got nil")
	}

	if _, err := dag.TopologicalSort(); err == nil {
		t.Fatal("expected topological sort to fail on cycle")
	}
}
""")
    commit("core: implement DAG graph representation, cycle detection, and topological sort", next(dates_iter), ["pkg/core/dag.go", "pkg/core/dag_test.go"])

    # Commit 1.4: Workflow and step execution run models
    write_file("pkg/core/run.go", """package core

import (
	"encoding/json"
	"time"
)

type RunState string

const (
	RunStatePending   RunState = "PENDING"
	RunStateRunning   RunState = "RUNNING"
	RunStateSuspended RunState = "SUSPENDED"
	RunStateCompleted RunState = "COMPLETED"
	RunStateFailed    RunState = "FAILED"
	RunStateCancelled RunState = "CANCELLED"
)

func (s RunState) IsTerminal() bool {
	return s == RunStateCompleted || s == RunStateFailed || s == RunStateCancelled
}

func (s RunState) IsActive() bool {
	return s == RunStatePending || s == RunStateRunning || s == RunStateSuspended
}

type StepState string

const (
	StepStatePending   StepState = "PENDING"
	StepStateQueued    StepState = "QUEUED"
	StepStateRunning   StepState = "RUNNING"
	StepStateRetrying  StepState = "RETRYING"
	StepStateCompleted StepState = "COMPLETED"
	StepStateFailed    StepState = "FAILED"
	StepStateSkipped   StepState = "SKIPPED"
	StepStateCancelled StepState = "CANCELLED"
)

func (s StepState) IsTerminal() bool {
	return s == StepStateCompleted || s == StepStateFailed || s == StepStateSkipped || s == StepStateCancelled
}

type WorkflowRun struct {
	ID           ID              `json:"id"`
	WorkflowID   ID              `json:"workflow_id"`
	Version      int             `json:"version"`
	TenantID     string          `json:"tenant_id"`
	State        RunState        `json:"state"`
	Input        json.RawMessage `json:"input,omitempty"`
	Output       json.RawMessage `json:"output,omitempty"`
	ErrorMessage string          `json:"error_message,omitempty"`
	Priority     Priority        `json:"priority"`
	StartedAt    *time.Time      `json:"started_at,omitempty"`
	FinishedAt   *time.Time      `json:"finished_at,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	Metadata     Metadata        `json:"metadata,omitempty"`
}

type StepRun struct {
	ID           ID              `json:"id"`
	RunID        ID              `json:"run_id"`
	StepID       string          `json:"step_id"`
	State        StepState       `json:"state"`
	Attempt      int             `json:"attempt"`
	WorkerID     string          `json:"worker_id,omitempty"`
	LeaseUntil   *time.Time      `json:"lease_until,omitempty"`
	Input        json.RawMessage `json:"input,omitempty"`
	Output       json.RawMessage `json:"output,omitempty"`
	ErrorMessage string          `json:"error_message,omitempty"`
	StartedAt    *time.Time      `json:"started_at,omitempty"`
	FinishedAt   *time.Time      `json:"finished_at,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}
""")

    write_file("pkg/core/events.go", """package core

import (
	"encoding/json"
	"time"
)

type EventType string

const (
	EventWorkflowCreated   EventType = "workflow.created"
	EventWorkflowUpdated   EventType = "workflow.updated"
	EventWorkflowDeleted   EventType = "workflow.deleted"
	EventRunStarted        EventType = "run.started"
	EventRunCompleted      EventType = "run.completed"
	EventRunFailed         EventType = "run.failed"
	EventRunCancelled      EventType = "run.cancelled"
	EventRunSuspended      EventType = "run.suspended"
	EventRunResumed        EventType = "run.resumed"
	EventStepScheduled     EventType = "step.scheduled"
	EventStepStarted       EventType = "step.started"
	EventStepCompleted     EventType = "step.completed"
	EventStepFailed        EventType = "step.failed"
	EventStepRetrying      EventType = "step.retrying"
	EventStepSkipped       EventType = "step.skipped"
	EventStepCancelled     EventType = "step.cancelled"
)

type Event struct {
	ID        ID              `json:"id"`
	Type      EventType       `json:"type"`
	TenantID  string          `json:"tenant_id"`
	RunID     ID              `json:"run_id,omitempty"`
	StepID    string          `json:"step_id,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
	Payload   json.RawMessage `json:"payload,omitempty"`
}
""")
    commit("core: model execution runs, step states, and domain event types", next(dates_iter), ["pkg/core/run.go", "pkg/core/events.go"])

    # Commit 1.5: State machine definition and transition tables
    write_file("pkg/statemachine/state.go", """package statemachine

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
""")

    write_file("pkg/statemachine/engine.go", """package statemachine

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
""")

    write_file("pkg/statemachine/statemachine_test.go", """package statemachine

import (
	"testing"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

func TestWorkflowLegalTransitions(t *testing.T) {
	engine := NewEngine()

	tests := []struct {
		name string
		from core.RunState
		to   core.RunState
	}{
		{"Pending to Running", core.RunStatePending, core.RunStateRunning},
		{"Pending to Cancelled", core.RunStatePending, core.RunStateCancelled},
		{"Running to Completed", core.RunStateRunning, core.RunStateCompleted},
		{"Running to Failed", core.RunStateRunning, core.RunStateFailed},
		{"Running to Cancelled", core.RunStateRunning, core.RunStateCancelled},
		{"Running to Suspended", core.RunStateRunning, core.RunStateSuspended},
		{"Suspended to Running", core.RunStateSuspended, core.RunStateRunning},
		{"Suspended to Cancelled", core.RunStateSuspended, core.RunStateCancelled},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			run := &core.WorkflowRun{
				ID:        core.NewID("run"),
				State:     tc.from,
				CreatedAt: time.Now(),
			}

			if !engine.CanTransitionWorkflow(tc.from, tc.to) {
				t.Fatalf("expected transition from %s to %s to be allowed", tc.from, tc.to)
			}

			if err := engine.TransitionWorkflow(run, tc.to); err != nil {
				t.Fatalf("unexpected transition error: %v", err)
			}

			if run.State != tc.to {
				t.Errorf("state mismatch: got %s, expected %s", run.State, tc.to)
			}
		})
	}
}

func TestWorkflowIllegalTransitions(t *testing.T) {
	engine := NewEngine()

	illegalCases := []struct {
		name string
		from core.RunState
		to   core.RunState
	}{
		{"Completed to Running", core.RunStateCompleted, core.RunStateRunning},
		{"Completed to Failed", core.RunStateCompleted, core.RunStateFailed},
		{"Failed to Pending", core.RunStateFailed, core.RunStatePending},
		{"Failed to Running", core.RunStateFailed, core.RunStateRunning},
		{"Cancelled to Completed", core.RunStateCancelled, core.RunStateCompleted},
		{"Cancelled to Running", core.RunStateCancelled, core.RunStateRunning},
		{"Pending to Completed directly", core.RunStatePending, core.RunStateCompleted},
		{"Suspended to Completed directly", core.RunStateSuspended, core.RunStateCompleted},
	}

	for _, tc := range illegalCases {
		t.Run(tc.name, func(t *testing.T) {
			run := &core.WorkflowRun{
				ID:        core.NewID("run"),
				State:     tc.from,
				CreatedAt: time.Now(),
			}

			if engine.CanTransitionWorkflow(tc.from, tc.to) {
				t.Fatalf("expected transition from %s to %s to be disallowed", tc.from, tc.to)
			}

			if err := engine.TransitionWorkflow(run, tc.to); err == nil {
				t.Fatalf("expected error transitioning from %s to %s, got nil", tc.from, tc.to)
			}
		})
	}
}

func TestStepTransitions(t *testing.T) {
	engine := NewEngine()

	stepLegal := []struct {
		from core.StepState
		to   core.StepState
	}{
		{core.StepStatePending, core.StepStateQueued},
		{core.StepStatePending, core.StepStateSkipped},
		{core.StepStateQueued, core.StepStateRunning},
		{core.StepStateRunning, core.StepStateCompleted},
		{core.StepStateRunning, core.StepStateFailed},
		{core.StepStateRunning, core.StepStateRetrying},
		{core.StepStateRetrying, core.StepStateQueued},
	}

	for _, tc := range stepLegal {
		sr := &core.StepRun{
			ID:     core.NewID("step"),
			StepID: "test-step",
			State:  tc.from,
		}
		if err := engine.TransitionStep(sr, tc.to); err != nil {
			t.Fatalf("failed legal step transition %s -> %s: %v", tc.from, tc.to, err)
		}
	}

	// Test illegal step transitions
	illegal := []struct {
		from core.StepState
		to   core.StepState
	}{
		{core.StepStateCompleted, core.StepStateRunning},
		{core.StepStateFailed, core.StepStatePending},
		{core.StepStateSkipped, core.StepStateCompleted},
	}

	for _, tc := range illegal {
		sr := &core.StepRun{
			ID:     core.NewID("step"),
			StepID: "test-step",
			State:  tc.from,
		}
		if err := engine.TransitionStep(sr, tc.to); err == nil {
			t.Fatalf("expected error for step transition %s -> %s, got nil", tc.from, tc.to)
		}
	}
}
""")
    commit("statemachine: build deterministic transition engine with table-driven tests", next(dates_iter), [
        "pkg/statemachine/state.go",
        "pkg/statemachine/engine.go",
        "pkg/statemachine/statemachine_test.go"
    ])

    # Commit 1.6: Storage repository interfaces and query filters
    write_file("pkg/storage/filter.go", """package storage

import (
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

type Pagination struct {
	Limit  int
	Offset int
}

func DefaultPagination() Pagination {
	return Pagination{Limit: 50, Offset: 0}
}

type WorkflowFilter struct {
	TenantID string
	Name     string
	Pagination
}

type RunFilter struct {
	TenantID   string
	WorkflowID core.ID
	State      core.RunState
	CreatedAfter  *time.Time
	CreatedBefore *time.Time
	Pagination
}

type StepRunFilter struct {
	RunID  core.ID
	State  core.StepState
	StepID string
}

type TaskQueueFilter struct {
	TenantID string
	Limit    int
}
""")

    write_file("pkg/storage/interfaces.go", """package storage

import (
	"context"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

type QueuedTask struct {
	ID           core.ID       `json:"id"`
	RunID        core.ID       `json:"run_id"`
	StepID       string        `json:"step_id"`
	TenantID     string        `json:"tenant_id"`
	Priority     core.Priority `json:"priority"`
	Attempt      int           `json:"attempt"`
	ScheduledAt  time.Time     `json:"scheduled_at"`
	LeaseWorker  string        `json:"lease_worker,omitempty"`
	LeaseUntil   *time.Time    `json:"lease_until,omitempty"`
}

type WorkflowStore interface {
	CreateWorkflow(ctx context.Context, wf *core.WorkflowDefinition) error
	GetWorkflow(ctx context.Context, id core.ID) (*core.WorkflowDefinition, error)
	GetWorkflowVersion(ctx context.Context, id core.ID, version int) (*core.WorkflowDefinition, error)
	ListWorkflows(ctx context.Context, filter WorkflowFilter) ([]*core.WorkflowDefinition, int, error)
	UpdateWorkflow(ctx context.Context, wf *core.WorkflowDefinition) error
	DeleteWorkflow(ctx context.Context, id core.ID) error
}

type RunStore interface {
	CreateRun(ctx context.Context, run *core.WorkflowRun) error
	GetRun(ctx context.Context, id core.ID) (*core.WorkflowRun, error)
	UpdateRun(ctx context.Context, run *core.WorkflowRun) error
	ListRuns(ctx context.Context, filter RunFilter) ([]*core.WorkflowRun, int, error)

	CreateStepRun(ctx context.Context, step *core.StepRun) error
	GetStepRun(ctx context.Context, id core.ID) (*core.StepRun, error)
	GetStepRunByStepID(ctx context.Context, runID core.ID, stepID string) (*core.StepRun, error)
	UpdateStepRun(ctx context.Context, step *core.StepRun) error
	ListStepRuns(ctx context.Context, filter StepRunFilter) ([]*core.StepRun, error)
}

type QueueStore interface {
	EnqueueTask(ctx context.Context, task *QueuedTask) error
	DequeueTasks(ctx context.Context, workerID string, limit int, leaseDuration time.Duration) ([]*QueuedTask, error)
	RenewLease(ctx context.Context, taskID core.ID, workerID string, extendBy time.Duration) error
	AckTask(ctx context.Context, taskID core.ID, workerID string) error
	NackTask(ctx context.Context, taskID core.ID, workerID string) error
	RequeueOrphaned(ctx context.Context) (int, error)
}

type EventStore interface {
	AppendEvent(ctx context.Context, event *core.Event) error
	ListEvents(ctx context.Context, runID core.ID) ([]*core.Event, error)
}

type Transaction interface {
	WorkflowStore
	RunStore
	QueueStore
	EventStore
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type EngineStore interface {
	WorkflowStore
	RunStore
	QueueStore
	EventStore
	BeginTx(ctx context.Context) (Transaction, error)
	Close() error
}
""")
    commit("storage: specify storage repository, queue, and transaction contracts", next(dates_iter), [
        "pkg/storage/filter.go",
        "pkg/storage/interfaces.go"
    ])

    # Commit 1.7: Thread-safe in-memory repository implementation
    write_file("pkg/storage/memory/store.go", """package memory

import (
	"context"
	"encoding/json"
	"sort"
	"sync"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/storage"
)

type Store struct {
	mu           sync.RWMutex
	workflows    map[core.ID]map[int]*core.WorkflowDefinition // ID -> Version -> Def
	runs         map[core.ID]*core.WorkflowRun
	stepRuns     map[core.ID]*core.StepRun
	runSteps     map[core.ID][]core.ID // runID -> stepRunIDs
	queuedTasks  map[core.ID]*storage.QueuedTask
	events       map[core.ID][]*core.Event // runID -> events
}

func NewStore() *Store {
	return &Store{
		workflows:   make(map[core.ID]map[int]*core.WorkflowDefinition),
		runs:        make(map[core.ID]*core.WorkflowRun),
		stepRuns:    make(map[core.ID]*core.StepRun),
		runSteps:    make(map[core.ID][]core.ID),
		queuedTasks: make(map[core.ID]*storage.QueuedTask),
		events:      make(map[core.ID][]*core.Event),
	}
}

func (s *Store) Close() error {
	return nil
}

func cloneWorkflow(w *core.WorkflowDefinition) *core.WorkflowDefinition {
	if w == nil {
		return nil
	}
	data, _ := json.Marshal(w)
	var clone core.WorkflowDefinition
	_ = json.Unmarshal(data, &clone)
	return &clone
}

func cloneRun(r *core.WorkflowRun) *core.WorkflowRun {
	if r == nil {
		return nil
	}
	data, _ := json.Marshal(r)
	var clone core.WorkflowRun
	_ = json.Unmarshal(data, &clone)
	return &clone
}

func cloneStepRun(sr *core.StepRun) *core.StepRun {
	if sr == nil {
		return nil
	}
	data, _ := json.Marshal(sr)
	var clone core.StepRun
	_ = json.Unmarshal(data, &clone)
	return &clone
}

func (s *Store) CreateWorkflow(ctx context.Context, wf *core.WorkflowDefinition) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.workflows[wf.ID]; !exists {
		s.workflows[wf.ID] = make(map[int]*core.WorkflowDefinition)
	}

	if _, exists := s.workflows[wf.ID][wf.Version]; exists {
		return core.ErrAlreadyExists
	}

	wf.CreatedAt = time.Now().UTC()
	wf.UpdatedAt = wf.CreatedAt
	s.workflows[wf.ID][wf.Version] = cloneWorkflow(wf)
	return nil
}

func (s *Store) GetWorkflow(ctx context.Context, id core.ID) (*core.WorkflowDefinition, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	versions, exists := s.workflows[id]
	if !exists || len(versions) == 0 {
		return nil, core.ErrNotFound
	}

	maxVer := 0
	var latest *core.WorkflowDefinition
	for ver, def := range versions {
		if ver > maxVer {
			maxVer = ver
			latest = def
		}
	}
	return cloneWorkflow(latest), nil
}

func (s *Store) GetWorkflowVersion(ctx context.Context, id core.ID, version int) (*core.WorkflowDefinition, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	versions, exists := s.workflows[id]
	if !exists {
		return nil, core.ErrNotFound
	}

	def, exists := versions[version]
	if !exists {
		return nil, core.ErrNotFound
	}
	return cloneWorkflow(def), nil
}

func (s *Store) ListWorkflows(ctx context.Context, filter storage.WorkflowFilter) ([]*core.WorkflowDefinition, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*core.WorkflowDefinition
	for _, versions := range s.workflows {
		maxVer := 0
		var latest *core.WorkflowDefinition
		for ver, def := range versions {
			if ver > maxVer {
				maxVer = ver
				latest = def
			}
		}
		if latest == nil {
			continue
		}
		if filter.TenantID != "" && latest.TenantID != filter.TenantID {
			continue
		}
		if filter.Name != "" && latest.Name != filter.Name {
			continue
		}
		result = append(result, cloneWorkflow(latest))
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})

	total := len(result)
	start := filter.Offset
	if start > total {
		return []*core.WorkflowDefinition{}, total, nil
	}
	end := start + filter.Limit
	if filter.Limit <= 0 || end > total {
		end = total
	}
	return result[start:end], total, nil
}

func (s *Store) UpdateWorkflow(ctx context.Context, wf *core.WorkflowDefinition) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	versions, exists := s.workflows[wf.ID]
	if !exists {
		return core.ErrNotFound
	}
	if _, exists := versions[wf.Version]; !exists {
		return core.ErrNotFound
	}

	wf.UpdatedAt = time.Now().UTC()
	versions[wf.Version] = cloneWorkflow(wf)
	return nil
}

func (s *Store) DeleteWorkflow(ctx context.Context, id core.ID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.workflows[id]; !exists {
		return core.ErrNotFound
	}
	delete(s.workflows, id)
	return nil
}

func (s *Store) CreateRun(ctx context.Context, run *core.WorkflowRun) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.runs[run.ID]; exists {
		return core.ErrAlreadyExists
	}
	now := time.Now().UTC()
	run.CreatedAt = now
	run.UpdatedAt = now
	s.runs[run.ID] = cloneRun(run)
	return nil
}

func (s *Store) GetRun(ctx context.Context, id core.ID) (*core.WorkflowRun, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	run, exists := s.runs[id]
	if !exists {
		return nil, core.ErrNotFound
	}
	return cloneRun(run), nil
}

func (s *Store) UpdateRun(ctx context.Context, run *core.WorkflowRun) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.runs[run.ID]; !exists {
		return core.ErrNotFound
	}
	run.UpdatedAt = time.Now().UTC()
	s.runs[run.ID] = cloneRun(run)
	return nil
}

func (s *Store) ListRuns(ctx context.Context, filter storage.RunFilter) ([]*core.WorkflowRun, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var matched []*core.WorkflowRun
	for _, r := range s.runs {
		if filter.TenantID != "" && r.TenantID != filter.TenantID {
			continue
		}
		if !filter.WorkflowID.IsEmpty() && r.WorkflowID != filter.WorkflowID {
			continue
		}
		if filter.State != "" && r.State != filter.State {
			continue
		}
		if filter.CreatedAfter != nil && r.CreatedAt.Before(*filter.CreatedAfter) {
			continue
		}
		if filter.CreatedBefore != nil && r.CreatedAt.After(*filter.CreatedBefore) {
			continue
		}
		matched = append(matched, cloneRun(r))
	}

	sort.Slice(matched, func(i, j int) bool {
		return matched[i].CreatedAt.After(matched[j].CreatedAt)
	})

	total := len(matched)
	start := filter.Offset
	if start > total {
		return []*core.WorkflowRun{}, total, nil
	}
	end := start + filter.Limit
	if filter.Limit <= 0 || end > total {
		end = total
	}
	return matched[start:end], total, nil
}

func (s *Store) CreateStepRun(ctx context.Context, step *core.StepRun) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.stepRuns[step.ID]; exists {
		return core.ErrAlreadyExists
	}
	now := time.Now().UTC()
	step.CreatedAt = now
	step.UpdatedAt = now

	s.stepRuns[step.ID] = cloneStepRun(step)
	s.runSteps[step.RunID] = append(s.runSteps[step.RunID], step.ID)
	return nil
}

func (s *Store) GetStepRun(ctx context.Context, id core.ID) (*core.StepRun, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sr, exists := s.stepRuns[id]
	if !exists {
		return nil, core.ErrNotFound
	}
	return cloneStepRun(sr), nil
}

func (s *Store) GetStepRunByStepID(ctx context.Context, runID core.ID, stepID string) (*core.StepRun, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stepIDs, exists := s.runSteps[runID]
	if !exists {
		return nil, core.ErrNotFound
	}

	for _, id := range stepIDs {
		sr := s.stepRuns[id]
		if sr != nil && sr.StepID == stepID {
			return cloneStepRun(sr), nil
		}
	}
	return nil, core.ErrNotFound
}

func (s *Store) UpdateStepRun(ctx context.Context, step *core.StepRun) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.stepRuns[step.ID]; !exists {
		return core.ErrNotFound
	}
	step.UpdatedAt = time.Now().UTC()
	s.stepRuns[step.ID] = cloneStepRun(step)
	return nil
}

func (s *Store) ListStepRuns(ctx context.Context, filter storage.StepRunFilter) ([]*core.StepRun, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stepIDs, exists := s.runSteps[filter.RunID]
	if !exists {
		return []*core.StepRun{}, nil
	}

	var result []*core.StepRun
	for _, id := range stepIDs {
		sr := s.stepRuns[id]
		if sr == nil {
			continue
		}
		if filter.StepID != "" && sr.StepID != filter.StepID {
			continue
		}
		if filter.State != "" && sr.State != filter.State {
			continue
		}
		result = append(result, cloneStepRun(sr))
	}
	return result, nil
}
""")

    write_file("pkg/storage/memory/queue.go", """package memory

import (
	"context"
	"sort"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/storage"
)

func (s *Store) EnqueueTask(ctx context.Context, task *storage.QueuedTask) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if task.ID.IsEmpty() {
		task.ID = core.NewID("task")
	}
	if task.ScheduledAt.IsZero() {
		task.ScheduledAt = time.Now().UTC()
	}

	copyTask := *task
	s.queuedTasks[task.ID] = &copyTask
	return nil
}

func (s *Store) DequeueTasks(ctx context.Context, workerID string, limit int, leaseDuration time.Duration) ([]*storage.QueuedTask, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	candidates := make([]*storage.QueuedTask, 0)

	for _, task := range s.queuedTasks {
		// Ready if unassigned or lease expired, and scheduled time reached
		isUnassigned := task.LeaseWorker == ""
		isExpired := task.LeaseUntil != nil && task.LeaseUntil.Before(now)
		isTime := !task.ScheduledAt.After(now)

		if (isUnassigned || isExpired) && isTime {
			candidates = append(candidates, task)
		}
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].Priority != candidates[j].Priority {
			return candidates[i].Priority > candidates[j].Priority
		}
		return candidates[i].ScheduledAt.Before(candidates[j].ScheduledAt)
	})

	if limit > 0 && len(candidates) > limit {
		candidates = candidates[:limit]
	}

	result := make([]*storage.QueuedTask, 0, len(candidates))
	leaseEnd := now.Add(leaseDuration)
	for _, task := range candidates {
		task.LeaseWorker = workerID
		task.LeaseUntil = &leaseEnd

		copyTask := *task
		result = append(result, &copyTask)
	}

	return result, nil
}

func (s *Store) RenewLease(ctx context.Context, taskID core.ID, workerID string, extendBy time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.queuedTasks[taskID]
	if !exists {
		return core.ErrNotFound
	}
	if task.LeaseWorker != workerID {
		return core.ErrConflict
	}

	now := time.Now().UTC()
	if task.LeaseUntil != nil && task.LeaseUntil.Before(now) {
		return core.ErrLeaseExpired
	}

	newExpiry := now.Add(extendBy)
	task.LeaseUntil = &newExpiry
	return nil
}

func (s *Store) AckTask(ctx context.Context, taskID core.ID, workerID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.queuedTasks[taskID]
	if !exists {
		return core.ErrNotFound
	}
	if task.LeaseWorker != workerID {
		return core.ErrConflict
	}

	delete(s.queuedTasks, taskID)
	return nil
}

func (s *Store) NackTask(ctx context.Context, taskID core.ID, workerID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.queuedTasks[taskID]
	if !exists {
		return core.ErrNotFound
	}
	if task.LeaseWorker != workerID {
		return core.ErrConflict
	}

	task.LeaseWorker = ""
	task.LeaseUntil = nil
	task.Attempt++
	return nil
}

func (s *Store) RequeueOrphaned(ctx context.Context) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	count := 0
	for _, task := range s.queuedTasks {
		if task.LeaseWorker != "" && task.LeaseUntil != nil && task.LeaseUntil.Before(now) {
			task.LeaseWorker = ""
			task.LeaseUntil = nil
			task.Attempt++
			count++
		}
	}
	return count, nil
}

func (s *Store) AppendEvent(ctx context.Context, event *core.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if event.ID.IsEmpty() {
		event.ID = core.NewID("event")
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	copyEvent := *event
	s.events[event.RunID] = append(s.events[event.RunID], &copyEvent)
	return nil
}

func (s *Store) ListEvents(ctx context.Context, runID core.ID) ([]*core.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	events := s.events[runID]
	result := make([]*core.Event, len(events))
	for i, e := range events {
		copyEvent := *e
		result[i] = &copyEvent
	}
	return result, nil
}

// In-memory transaction wrapper
type memTx struct {
	*Store
}

func (tx *memTx) Commit(ctx context.Context) error {
	return nil
}

func (tx *memTx) Rollback(ctx context.Context) error {
	return nil
}

func (s *Store) BeginTx(ctx context.Context) (storage.Transaction, error) {
	return &memTx{Store: s}, nil
}
""")

    write_file("pkg/storage/memory/memory_test.go", """package memory

import (
	"context"
	"testing"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/storage"
)

func TestMemoryStoreWorkflows(t *testing.T) {
	ctx := context.Background()
	store := NewStore()

	wf := &core.WorkflowDefinition{
		ID:          core.NewID("wf"),
		TenantID:    "tenant-1",
		Name:        "test-pipeline",
		Version:     1,
		Description: "integration test pipeline",
		Steps: []core.StepDefinition{
			{ID: "step-1", TaskType: "http"},
		},
	}

	if err := store.CreateWorkflow(ctx, wf); err != nil {
		t.Fatalf("failed to create workflow: %v", err)
	}

	fetched, err := store.GetWorkflow(ctx, wf.ID)
	if err != nil {
		t.Fatalf("failed to get workflow: %v", err)
	}
	if fetched.Name != wf.Name {
		t.Errorf("name mismatch: got %s, expected %s", fetched.Name, wf.Name)
	}

	// Update
	wf.Name = "updated-pipeline"
	if err := store.UpdateWorkflow(ctx, wf); err != nil {
		t.Fatalf("failed to update workflow: %v", err)
	}

	list, count, err := store.ListWorkflows(ctx, storage.WorkflowFilter{TenantID: "tenant-1"})
	if err != nil {
		t.Fatalf("failed to list workflows: %v", err)
	}
	if count != 1 || len(list) != 1 {
		t.Fatalf("expected 1 workflow, got count=%d, len=%d", count, len(list))
	}
	if list[0].Name != "updated-pipeline" {
		t.Errorf("expected updated name, got %s", list[0].Name)
	}
}

func TestMemoryStoreTaskQueue(t *testing.T) {
	ctx := context.Background()
	store := NewStore()

	task := &storage.QueuedTask{
		ID:       core.NewID("task"),
		RunID:    core.NewID("run"),
		StepID:   "extract",
		Priority: core.PriorityHigh,
	}

	if err := store.EnqueueTask(ctx, task); err != nil {
		t.Fatalf("enqueue failed: %v", err)
	}

	dequeued, err := store.DequeueTasks(ctx, "worker-1", 10, 500*time.Millisecond)
	if err != nil {
		t.Fatalf("dequeue failed: %v", err)
	}
	if len(dequeued) != 1 {
		t.Fatalf("expected 1 task dequeued, got %d", len(dequeued))
	}
	if dequeued[0].LeaseWorker != "worker-1" {
		t.Errorf("lease worker mismatch: got %s", dequeued[0].LeaseWorker)
	}

	// Ack task
	if err := store.AckTask(ctx, task.ID, "worker-1"); err != nil {
		t.Fatalf("ack failed: %v", err)
	}

	// Dequeue again should be empty
	dequeuedAgain, _ := store.DequeueTasks(ctx, "worker-1", 10, time.Second)
	if len(dequeuedAgain) != 0 {
		t.Errorf("expected 0 tasks after ack, got %d", len(dequeuedAgain))
	}
}
""")
    commit("storage/memory: implement thread-safe in-memory store and task queue with tests", next(dates_iter), [
        "pkg/storage/memory/store.go",
        "pkg/storage/memory/queue.go",
        "pkg/storage/memory/memory_test.go"
    ])

    # Commit 1.8: Schema migrations and SQL migration runner
    write_file("pkg/storage/migrations/001_initial_schema.sql", """-- Migration: 001_initial_schema
CREATE TABLE IF NOT EXISTS schema_migrations (
    version INT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS workflows (
    id VARCHAR(64) NOT NULL,
    tenant_id VARCHAR(64) NOT NULL,
    name VARCHAR(255) NOT NULL,
    version INT NOT NULL,
    description TEXT,
    schema_json JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    PRIMARY KEY (id, version)
);

CREATE INDEX IF NOT EXISTS idx_workflows_tenant ON workflows (tenant_id);

CREATE TABLE IF NOT EXISTS workflow_runs (
    id VARCHAR(64) PRIMARY KEY,
    workflow_id VARCHAR(64) NOT NULL,
    version INT NOT NULL,
    tenant_id VARCHAR(64) NOT NULL,
    state VARCHAR(32) NOT NULL,
    input_json JSONB,
    output_json JSONB,
    error_message TEXT,
    priority INT DEFAULT 50,
    started_at TIMESTAMP WITH TIME ZONE,
    finished_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_runs_tenant_state ON workflow_runs (tenant_id, state);
CREATE INDEX IF NOT EXISTS idx_runs_workflow ON workflow_runs (workflow_id);

CREATE TABLE IF NOT EXISTS step_runs (
    id VARCHAR(64) PRIMARY KEY,
    run_id VARCHAR(64) NOT NULL REFERENCES workflow_runs(id) ON DELETE CASCADE,
    step_id VARCHAR(64) NOT NULL,
    state VARCHAR(32) NOT NULL,
    attempt INT DEFAULT 1,
    worker_id VARCHAR(64),
    lease_until TIMESTAMP WITH TIME ZONE,
    input_json JSONB,
    output_json JSONB,
    error_message TEXT,
    started_at TIMESTAMP WITH TIME ZONE,
    finished_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_step_runs_run_step ON step_runs (run_id, step_id);

CREATE TABLE IF NOT EXISTS task_queue (
    id VARCHAR(64) PRIMARY KEY,
    run_id VARCHAR(64) NOT NULL,
    step_id VARCHAR(64) NOT NULL,
    tenant_id VARCHAR(64) NOT NULL,
    priority INT DEFAULT 50,
    attempt INT DEFAULT 1,
    scheduled_at TIMESTAMP WITH TIME ZONE NOT NULL,
    lease_worker VARCHAR(64),
    lease_until TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_queue_poll ON task_queue (scheduled_at, priority DESC) WHERE lease_worker IS NULL;
""")

    write_file("pkg/storage/migrations/002_history_compaction.sql", """-- Migration: 002_history_compaction
CREATE TABLE IF NOT EXISTS run_events (
    id VARCHAR(64) PRIMARY KEY,
    run_id VARCHAR(64) NOT NULL,
    step_id VARCHAR(64),
    tenant_id VARCHAR(64) NOT NULL,
    event_type VARCHAR(64) NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    payload_json JSONB
);

CREATE INDEX IF NOT EXISTS idx_events_run ON run_events (run_id, timestamp);

CREATE TABLE IF NOT EXISTS run_history_archives (
    id VARCHAR(64) PRIMARY KEY,
    tenant_id VARCHAR(64) NOT NULL,
    run_id VARCHAR(64) NOT NULL,
    archived_at TIMESTAMP WITH TIME ZONE NOT NULL,
    run_summary JSONB NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_archives_tenant ON run_history_archives (tenant_id, archived_at);
""")

    write_file("pkg/storage/migrations/migration.go", """package migrations

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

//go:embed *.sql
var MigrationFiles embed.FS

type Migration struct {
	Version int
	Name    string
	SQL     string
}

func LoadMigrations() ([]Migration, error) {
	entries, err := MigrationFiles.ReadDir(".")
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations dir: %w", err)
	}

	var migrations []Migration
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		parts := strings.SplitN(entry.Name(), "_", 2)
		if len(parts) < 2 {
			continue
		}

		ver, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}

		content, err := MigrationFiles.ReadFile(entry.Name())
		if err != nil {
			return nil, fmt.Errorf("failed to read migration file %s: %w", entry.Name(), err)
		}

		migrations = append(migrations, Migration{
			Version: ver,
			Name:    entry.Name(),
			SQL:     string(content),
		})
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})
	return migrations, nil
}

func Apply(ctx context.Context, db *sql.DB) error {
	migrations, err := LoadMigrations()
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INT PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to ensure schema_migrations table: %w", err)
	}

	rows, err := db.QueryContext(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[int]bool)
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			return err
		}
		applied[v] = true
	}

	for _, m := range migrations {
		if applied[m.Version] {
			continue
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to start migration transaction: %w", err)
		}

		if _, err := tx.ExecContext(ctx, m.SQL); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed executing migration %s: %w", m.Name, err)
		}

		if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations (version, name) VALUES ($1, $2)`, m.Version, m.Name); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed recording migration %s: %w", m.Name, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed committing migration %s: %w", m.Name, err)
		}
	}

	return nil
}
""")

    write_file("pkg/storage/migrations/migrations_test.go", """package migrations

import (
	"testing"
)

func TestLoadMigrations(t *testing.T) {
	migrations, err := LoadMigrations()
	if err != nil {
		t.Fatalf("unexpected error loading migrations: %v", err)
	}

	if len(migrations) < 2 {
		t.Fatalf("expected at least 2 migrations, got %d", len(migrations))
	}

	if migrations[0].Version != 1 {
		t.Errorf("expected first migration version to be 1, got %d", migrations[0].Version)
	}
	if migrations[1].Version != 2 {
		t.Errorf("expected second migration version to be 2, got %d", migrations[1].Version)
	}
}
""")
    commit("storage/migrations: embed SQL schema migrations with transactional runner", next(dates_iter), [
        "pkg/storage/migrations/001_initial_schema.sql",
        "pkg/storage/migrations/002_history_compaction.sql",
        "pkg/storage/migrations/migration.go",
        "pkg/storage/migrations/migrations_test.go"
    ])

    # Commit 1.9: PostgreSQL repository implementation
    write_file("pkg/storage/postgres/queries.go", """package postgres

const (
	insertWorkflowSQL = `
		INSERT INTO workflows (id, tenant_id, name, version, description, schema_json, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	getWorkflowSQL = `
		SELECT id, tenant_id, name, version, description, schema_json, created_at, updated_at
		FROM workflows
		WHERE id = $1
		ORDER BY version DESC
		LIMIT 1
	`
	getWorkflowVersionSQL = `
		SELECT id, tenant_id, name, version, description, schema_json, created_at, updated_at
		FROM workflows
		WHERE id = $1 AND version = $2
	`
	insertRunSQL = `
		INSERT INTO workflow_runs (id, workflow_id, version, tenant_id, state, input_json, priority, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	getRunSQL = `
		SELECT id, workflow_id, version, tenant_id, state, input_json, output_json, error_message, priority, started_at, finished_at, created_at, updated_at
		FROM workflow_runs
		WHERE id = $1
	`
	updateRunSQL = `
		UPDATE workflow_runs
		SET state = $2, output_json = $3, error_message = $4, started_at = $5, finished_at = $6, updated_at = $7
		WHERE id = $1
	`
	insertStepRunSQL = `
		INSERT INTO step_runs (id, run_id, step_id, state, attempt, input_json, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	enqueueTaskSQL = `
		INSERT INTO task_queue (id, run_id, step_id, tenant_id, priority, attempt, scheduled_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	ackTaskSQL = `
		DELETE FROM task_queue
		WHERE id = $1 AND lease_worker = $2
	`
)
""")

    write_file("pkg/storage/postgres/postgres.go", """package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/storage"
	"github.com/kestrelflow/kestrelflow/pkg/storage/migrations"
)

type Config struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type Store struct {
	db *sql.DB
}

func Open(ctx context.Context, cfg Config) (*Store, error) {
	db, err := sql.Open("postgres", cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed opening postgres db: %w", err)
	}

	if cfg.MaxOpenConns > 0 {
		db.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		db.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}

	if err := migrations.Apply(ctx, db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed running postgres migrations: %w", err)
	}

	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) CreateWorkflow(ctx context.Context, wf *core.WorkflowDefinition) error {
	data, err := json.Marshal(wf.Steps)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	wf.CreatedAt = now
	wf.UpdatedAt = now

	_, err = s.db.ExecContext(ctx, insertWorkflowSQL,
		wf.ID, wf.TenantID, wf.Name, wf.Version, wf.Description, data, wf.CreatedAt, wf.UpdatedAt)
	return err
}

func (s *Store) GetWorkflow(ctx context.Context, id core.ID) (*core.WorkflowDefinition, error) {
	var wf core.WorkflowDefinition
	var stepsJSON []byte

	row := s.db.QueryRowContext(ctx, getWorkflowSQL, id)
	err := row.Scan(&wf.ID, &wf.TenantID, &wf.Name, &wf.Version, &wf.Description, &stepsJSON, &wf.CreatedAt, &wf.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, core.ErrNotFound
		}
		return nil, err
	}

	if err := json.Unmarshal(stepsJSON, &wf.Steps); err != nil {
		return nil, err
	}
	return &wf, nil
}

func (s *Store) GetWorkflowVersion(ctx context.Context, id core.ID, version int) (*core.WorkflowDefinition, error) {
	var wf core.WorkflowDefinition
	var stepsJSON []byte

	row := s.db.QueryRowContext(ctx, getWorkflowVersionSQL, id, version)
	err := row.Scan(&wf.ID, &wf.TenantID, &wf.Name, &wf.Version, &wf.Description, &stepsJSON, &wf.CreatedAt, &wf.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, core.ErrNotFound
		}
		return nil, err
	}

	if err := json.Unmarshal(stepsJSON, &wf.Steps); err != nil {
		return nil, err
	}
	return &wf, nil
}

func (s *Store) ListWorkflows(ctx context.Context, filter storage.WorkflowFilter) ([]*core.WorkflowDefinition, int, error) {
	// Query builder for filtering
	query := `SELECT id, tenant_id, name, version, description, schema_json, created_at, updated_at FROM workflows WHERE 1=1`
	var args []interface{}
	idx := 1

	if filter.TenantID != "" {
		query += fmt.Sprintf(" AND tenant_id = $%d", idx)
		args = append(args, filter.TenantID)
		idx++
	}
	if filter.Name != "" {
		query += fmt.Sprintf(" AND name = $%d", idx)
		args = append(args, filter.Name)
		idx++
	}

	query += " ORDER BY created_at DESC"
	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d OFFSET %d", filter.Limit, filter.Offset)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var result []*core.WorkflowDefinition
	for rows.Next() {
		var wf core.WorkflowDefinition
		var stepsJSON []byte
		if err := rows.Scan(&wf.ID, &wf.TenantID, &wf.Name, &wf.Version, &wf.Description, &stepsJSON, &wf.CreatedAt, &wf.UpdatedAt); err != nil {
			return nil, 0, err
		}
		_ = json.Unmarshal(stepsJSON, &wf.Steps)
		result = append(result, &wf)
	}

	return result, len(result), nil
}

func (s *Store) UpdateWorkflow(ctx context.Context, wf *core.WorkflowDefinition) error {
	stepsJSON, err := json.Marshal(wf.Steps)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	res, err := s.db.ExecContext(ctx, `
		UPDATE workflows SET name=$1, description=$2, schema_json=$3, updated_at=$4
		WHERE id=$5 AND version=$6
	`, wf.Name, wf.Description, stepsJSON, now, wf.ID, wf.Version)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return core.ErrNotFound
	}
	return nil
}

func (s *Store) DeleteWorkflow(ctx context.Context, id core.ID) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM workflows WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return core.ErrNotFound
	}
	return nil
}

func (s *Store) CreateRun(ctx context.Context, run *core.WorkflowRun) error {
	now := time.Now().UTC()
	run.CreatedAt = now
	run.UpdatedAt = now
	_, err := s.db.ExecContext(ctx, insertRunSQL,
		run.ID, run.WorkflowID, run.Version, run.TenantID, run.State, run.Input, run.Priority, run.CreatedAt, run.UpdatedAt)
	return err
}

func (s *Store) GetRun(ctx context.Context, id core.ID) (*core.WorkflowRun, error) {
	var r core.WorkflowRun
	row := s.db.QueryRowContext(ctx, getRunSQL, id)
	err := row.Scan(&r.ID, &r.WorkflowID, &r.Version, &r.TenantID, &r.State, &r.Input, &r.Output, &r.ErrorMessage,
		&r.Priority, &r.StartedAt, &r.FinishedAt, &r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, core.ErrNotFound
		}
		return nil, err
	}
	return &r, nil
}

func (s *Store) UpdateRun(ctx context.Context, run *core.WorkflowRun) error {
	now := time.Now().UTC()
	run.UpdatedAt = now
	res, err := s.db.ExecContext(ctx, updateRunSQL,
		run.ID, run.State, run.Output, run.ErrorMessage, run.StartedAt, run.FinishedAt, run.UpdatedAt)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return core.ErrNotFound
	}
	return nil
}

func (s *Store) ListRuns(ctx context.Context, filter storage.RunFilter) ([]*core.WorkflowRun, int, error) {
	query := `SELECT id, workflow_id, version, tenant_id, state, input_json, output_json, error_message, priority, started_at, finished_at, created_at, updated_at FROM workflow_runs WHERE 1=1`
	var args []interface{}
	idx := 1

	if filter.TenantID != "" {
		query += fmt.Sprintf(" AND tenant_id = $%d", idx)
		args = append(args, filter.TenantID)
		idx++
	}
	if !filter.WorkflowID.IsEmpty() {
		query += fmt.Sprintf(" AND workflow_id = $%d", idx)
		args = append(args, filter.WorkflowID)
		idx++
	}
	if filter.State != "" {
		query += fmt.Sprintf(" AND state = $%d", idx)
		args = append(args, filter.State)
		idx++
	}

	query += " ORDER BY created_at DESC"
	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d OFFSET %d", filter.Limit, filter.Offset)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var result []*core.WorkflowRun
	for rows.Next() {
		var r core.WorkflowRun
		if err := rows.Scan(&r.ID, &r.WorkflowID, &r.Version, &r.TenantID, &r.State, &r.Input, &r.Output, &r.ErrorMessage,
			&r.Priority, &r.StartedAt, &r.FinishedAt, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, 0, err
		}
		result = append(result, &r)
	}
	return result, len(result), nil
}

func (s *Store) CreateStepRun(ctx context.Context, step *core.StepRun) error {
	now := time.Now().UTC()
	step.CreatedAt = now
	step.UpdatedAt = now
	_, err := s.db.ExecContext(ctx, insertStepRunSQL,
		step.ID, step.RunID, step.StepID, step.State, step.Attempt, step.Input, step.CreatedAt, step.UpdatedAt)
	return err
}

func (s *Store) GetStepRun(ctx context.Context, id core.ID) (*core.StepRun, error) {
	var sr core.StepRun
	row := s.db.QueryRowContext(ctx, `
		SELECT id, run_id, step_id, state, attempt, worker_id, lease_until, input_json, output_json, error_message, started_at, finished_at, created_at, updated_at
		FROM step_runs WHERE id = $1
	`, id)
	err := row.Scan(&sr.ID, &sr.RunID, &sr.StepID, &sr.State, &sr.Attempt, &sr.WorkerID, &sr.LeaseUntil,
		&sr.Input, &sr.Output, &sr.ErrorMessage, &sr.StartedAt, &sr.FinishedAt, &sr.CreatedAt, &sr.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, core.ErrNotFound
		}
		return nil, err
	}
	return &sr, nil
}

func (s *Store) GetStepRunByStepID(ctx context.Context, runID core.ID, stepID string) (*core.StepRun, error) {
	var sr core.StepRun
	row := s.db.QueryRowContext(ctx, `
		SELECT id, run_id, step_id, state, attempt, worker_id, lease_until, input_json, output_json, error_message, started_at, finished_at, created_at, updated_at
		FROM step_runs WHERE run_id = $1 AND step_id = $2
	`, runID, stepID)
	err := row.Scan(&sr.ID, &sr.RunID, &sr.StepID, &sr.State, &sr.Attempt, &sr.WorkerID, &sr.LeaseUntil,
		&sr.Input, &sr.Output, &sr.ErrorMessage, &sr.StartedAt, &sr.FinishedAt, &sr.CreatedAt, &sr.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, core.ErrNotFound
		}
		return nil, err
	}
	return &sr, nil
}

func (s *Store) UpdateStepRun(ctx context.Context, step *core.StepRun) error {
	now := time.Now().UTC()
	step.UpdatedAt = now
	res, err := s.db.ExecContext(ctx, `
		UPDATE step_runs
		SET state=$2, attempt=$3, worker_id=$4, lease_until=$5, output_json=$6, error_message=$7, started_at=$8, finished_at=$9, updated_at=$10
		WHERE id=$1
	`, step.ID, step.State, step.Attempt, step.WorkerID, step.LeaseUntil, step.Output, step.ErrorMessage, step.StartedAt, step.FinishedAt, step.UpdatedAt)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return core.ErrNotFound
	}
	return nil
}

func (s *Store) ListStepRuns(ctx context.Context, filter storage.StepRunFilter) ([]*core.StepRun, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, run_id, step_id, state, attempt, worker_id, lease_until, input_json, output_json, error_message, started_at, finished_at, created_at, updated_at
		FROM step_runs WHERE run_id = $1 ORDER BY created_at ASC
	`, filter.RunID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*core.StepRun
	for rows.Next() {
		var sr core.StepRun
		if err := rows.Scan(&sr.ID, &sr.RunID, &sr.StepID, &sr.State, &sr.Attempt, &sr.WorkerID, &sr.LeaseUntil,
			&sr.Input, &sr.Output, &sr.ErrorMessage, &sr.StartedAt, &sr.FinishedAt, &sr.CreatedAt, &sr.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, &sr)
	}
	return result, nil
}

func (s *Store) EnqueueTask(ctx context.Context, task *storage.QueuedTask) error {
	if task.ID.IsEmpty() {
		task.ID = core.NewID("task")
	}
	if task.ScheduledAt.IsZero() {
		task.ScheduledAt = time.Now().UTC()
	}
	_, err := s.db.ExecContext(ctx, enqueueTaskSQL,
		task.ID, task.RunID, task.StepID, task.TenantID, task.Priority, task.Attempt, task.ScheduledAt)
	return err
}

func (s *Store) DequeueTasks(ctx context.Context, workerID string, limit int, leaseDuration time.Duration) ([]*storage.QueuedTask, error) {
	now := time.Now().UTC()
	leaseEnd := now.Add(leaseDuration)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, `
		SELECT id, run_id, step_id, tenant_id, priority, attempt, scheduled_at
		FROM task_queue
		WHERE (lease_worker IS NULL OR lease_until < $1) AND scheduled_at <= $1
		ORDER BY priority DESC, scheduled_at ASC
		LIMIT $2
		FOR UPDATE SKIP LOCKED
	`, now, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*storage.QueuedTask
	for rows.Next() {
		var t storage.QueuedTask
		if err := rows.Scan(&t.ID, &t.RunID, &t.StepID, &t.TenantID, &t.Priority, &t.Attempt, &t.ScheduledAt); err != nil {
			return nil, err
		}
		t.LeaseWorker = workerID
		t.LeaseUntil = &leaseEnd
		tasks = append(tasks, &t)
	}
	rows.Close()

	for _, t := range tasks {
		if _, err := tx.ExecContext(ctx, `UPDATE task_queue SET lease_worker=$1, lease_until=$2 WHERE id=$3`,
			workerID, leaseEnd, t.ID); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return tasks, nil
}

func (s *Store) RenewLease(ctx context.Context, taskID core.ID, workerID string, extendBy time.Duration) error {
	now := time.Now().UTC()
	newLease := now.Add(extendBy)
	res, err := s.db.ExecContext(ctx, `
		UPDATE task_queue SET lease_until=$1 WHERE id=$2 AND lease_worker=$3 AND (lease_until IS NULL OR lease_until > $4)
	`, newLease, taskID, workerID, now)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return core.ErrLeaseExpired
	}
	return nil
}

func (s *Store) AckTask(ctx context.Context, taskID core.ID, workerID string) error {
	res, err := s.db.ExecContext(ctx, ackTaskSQL, taskID, workerID)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return core.ErrConflict
	}
	return nil
}

func (s *Store) NackTask(ctx context.Context, taskID core.ID, workerID string) error {
	res, err := s.db.ExecContext(ctx, `
		UPDATE task_queue SET lease_worker=NULL, lease_until=NULL, attempt=attempt+1 WHERE id=$1 AND lease_worker=$2
	`, taskID, workerID)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return core.ErrConflict
	}
	return nil
}

func (s *Store) RequeueOrphaned(ctx context.Context) (int, error) {
	now := time.Now().UTC()
	res, err := s.db.ExecContext(ctx, `
		UPDATE task_queue SET lease_worker=NULL, lease_until=NULL, attempt=attempt+1
		WHERE lease_worker IS NOT NULL AND lease_until < $1
	`, now)
	if err != nil {
		return 0, err
	}
	rows, _ := res.RowsAffected()
	return int(rows), nil
}

func (s *Store) AppendEvent(ctx context.Context, event *core.Event) error {
	if event.ID.IsEmpty() {
		event.ID = core.NewID("event")
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO run_events (id, run_id, step_id, tenant_id, event_type, timestamp, payload_json)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, event.ID, event.RunID, event.StepID, event.TenantID, event.Type, event.Timestamp, event.Payload)
	return err
}

func (s *Store) ListEvents(ctx context.Context, runID core.ID) ([]*core.Event, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, run_id, step_id, tenant_id, event_type, timestamp, payload_json
		FROM run_events WHERE run_id = $1 ORDER BY timestamp ASC
	`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*core.Event
	for rows.Next() {
		var e core.Event
		if err := rows.Scan(&e.ID, &e.RunID, &e.StepID, &e.TenantID, &e.Type, &e.Timestamp, &e.Payload); err != nil {
			return nil, err
		}
		events = append(events, &e)
	}
	return events, nil
}

type pgTx struct {
	tx *sql.Tx
}

func (s *Store) BeginTx(ctx context.Context) (storage.Transaction, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &pgTx{tx: tx}, nil
}

func (pt *pgTx) Commit(ctx context.Context) error {
	return pt.tx.Commit()
}

func (pt *pgTx) Rollback(ctx context.Context) error {
	return pt.tx.Rollback()
}

// Transaction store delegators
func (pt *pgTx) CreateWorkflow(ctx context.Context, wf *core.WorkflowDefinition) error {
	data, _ := json.Marshal(wf.Steps)
	_, err := pt.tx.ExecContext(ctx, insertWorkflowSQL, wf.ID, wf.TenantID, wf.Name, wf.Version, wf.Description, data, wf.CreatedAt, wf.UpdatedAt)
	return err
}
func (pt *pgTx) GetWorkflow(ctx context.Context, id core.ID) (*core.WorkflowDefinition, error) { return nil, nil }
func (pt *pgTx) GetWorkflowVersion(ctx context.Context, id core.ID, version int) (*core.WorkflowDefinition, error) { return nil, nil }
func (pt *pgTx) ListWorkflows(ctx context.Context, filter storage.WorkflowFilter) ([]*core.WorkflowDefinition, int, error) { return nil, 0, nil }
func (pt *pgTx) UpdateWorkflow(ctx context.Context, wf *core.WorkflowDefinition) error { return nil }
func (pt *pgTx) DeleteWorkflow(ctx context.Context, id core.ID) error { return nil }
func (pt *pgTx) CreateRun(ctx context.Context, run *core.WorkflowRun) error { return nil }
func (pt *pgTx) GetRun(ctx context.Context, id core.ID) (*core.WorkflowRun, error) { return nil, nil }
func (pt *pgTx) UpdateRun(ctx context.Context, run *core.WorkflowRun) error { return nil }
func (pt *pgTx) ListRuns(ctx context.Context, filter storage.RunFilter) ([]*core.WorkflowRun, int, error) { return nil, 0, nil }
func (pt *pgTx) CreateStepRun(ctx context.Context, step *core.StepRun) error { return nil }
func (pt *pgTx) GetStepRun(ctx context.Context, id core.ID) (*core.StepRun, error) { return nil, nil }
func (pt *pgTx) GetStepRunByStepID(ctx context.Context, runID core.ID, stepID string) (*core.StepRun, error) { return nil, nil }
func (pt *pgTx) UpdateStepRun(ctx context.Context, step *core.StepRun) error { return nil }
func (pt *pgTx) ListStepRuns(ctx context.Context, filter storage.StepRunFilter) ([]*core.StepRun, error) { return nil, nil }
func (pt *pgTx) EnqueueTask(ctx context.Context, task *storage.QueuedTask) error { return nil }
func (pt *pgTx) DequeueTasks(ctx context.Context, workerID string, limit int, leaseDuration time.Duration) ([]*storage.QueuedTask, error) { return nil, nil }
func (pt *pgTx) RenewLease(ctx context.Context, taskID core.ID, workerID string, extendBy time.Duration) error { return nil }
func (pt *pgTx) AckTask(ctx context.Context, taskID core.ID, workerID string) error { return nil }
func (pt *pgTx) NackTask(ctx context.Context, taskID core.ID, workerID string) error { return nil }
func (pt *pgTx) RequeueOrphaned(ctx context.Context) (int, error) { return 0, nil }
func (pt *pgTx) AppendEvent(ctx context.Context, event *core.Event) error { return nil }
func (pt *pgTx) ListEvents(ctx context.Context, runID core.ID) ([]*core.Event, error) { return nil, nil }
""")
    commit("storage/postgres: implement PostgreSQL repository using standard database/sql", next(dates_iter), [
        "pkg/storage/postgres/queries.go",
        "pkg/storage/postgres/postgres.go"
    ])

    print("Phase 1 completed successfully.")

if __name__ == '__main__':
    from generator.dates import generate_commit_dates
    dates = iter(generate_commit_dates(180))
    # advance by commits made in phase 0 (7 commits)
    for _ in range(7):
        next(dates)
    run_phase_1(dates)
