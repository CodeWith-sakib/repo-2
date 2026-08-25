package postgres

import (
	"fmt"
	"sync"
	"time"
)

// IndexRebuildTask represents a single SQL REINDEX statement scheduled for concurrency.
type IndexRebuildTask struct {
	IndexName        string    `json:"index_name"`
	TableName        string    `json:"table_name"`
	EstimatedBloatMB float64   `json:"estimated_bloat_mb"`
	RebuildSQL       string    `json:"rebuild_sql"`
	Priority         int       `json:"priority"` // 1 = Highest
	ScheduledAt      time.Time `json:"scheduled_at"`
}

// IndexRebuildPlanner prioritizes fragmented indexes for off-peak concurrent rebuilds.
type IndexRebuildPlanner struct {
	mu    sync.RWMutex
	tasks []IndexRebuildTask
}

// NewIndexRebuildPlanner creates a planner.
func NewIndexRebuildPlanner() *IndexRebuildPlanner {
	return &IndexRebuildPlanner{
		tasks: make([]IndexRebuildTask, 0),
	}
}

// ScheduleReindex enqueues an index for REINDEX CONCURRENTLY maintenance.
func (p *IndexRebuildPlanner) ScheduleReindex(tableName, indexName string, bloatMB float64) IndexRebuildTask {
	p.mu.Lock()
	defer p.mu.Unlock()

	priority := 3
	if bloatMB > 1000.0 {
		priority = 1
	} else if bloatMB > 100.0 {
		priority = 2
	}

	task := IndexRebuildTask{
		IndexName:        indexName,
		TableName:        tableName,
		EstimatedBloatMB: bloatMB,
		RebuildSQL:       fmt.Sprintf("REINDEX INDEX CONCURRENTLY %s;", indexName),
		Priority:         priority,
		ScheduledAt:      time.Now(),
	}

	p.tasks = append(p.tasks, task)
	return task
}

// PendingTasks returns queued index rebuild tasks.
func (p *IndexRebuildPlanner) PendingTasks() []IndexRebuildTask {
	p.mu.RLock()
	defer p.mu.RUnlock()

	out := make([]IndexRebuildTask, len(p.tasks))
	copy(out, p.tasks)
	return out
}

// Clear removes all pending tasks after maintenance window completes.
func (p *IndexRebuildPlanner) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.tasks = make([]IndexRebuildTask, 0)
}
