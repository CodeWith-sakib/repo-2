package core

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
