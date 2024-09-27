package core

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
