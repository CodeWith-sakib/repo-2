package storage

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
