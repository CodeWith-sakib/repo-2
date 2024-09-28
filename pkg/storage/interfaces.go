package storage

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
