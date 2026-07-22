package core

import (
	"context"
	"sync"
	"time"
)

// DeadLetterStepRecord captures details of a permanently failed step after exhausting retry limits.
type DeadLetterStepRecord struct {
	WorkflowID    string    `json:"workflow_id"`
	RunID         string    `json:"run_id"`
	StepID        string    `json:"step_id"`
	TotalRetries  int       `json:"total_retries"`
	TerminalError string    `json:"terminal_error"`
	PayloadDump   string    `json:"payload_dump"`
	RoutedAt      time.Time `json:"routed_at"`
}

// DeadLetterQueueRouter safely holds failed executions for human intervention and forensics.
type DeadLetterQueueRouter struct {
	mu           sync.RWMutex
	records      []DeadLetterStepRecord
	maxRetained  int
	deadLetterDL []string
}

// NewDeadLetterQueueRouter initializes a dead-letter queue collector.
func NewDeadLetterQueueRouter(maxRetained int) *DeadLetterQueueRouter {
	if maxRetained <= 0 {
		maxRetained = 5000
	}
	return &DeadLetterQueueRouter{
		records:     make([]DeadLetterStepRecord, 0),
		maxRetained: maxRetained,
	}
}

// RouteToDLQ logs a poison pill or fatally faulted task execution.
func (r *DeadLetterQueueRouter) RouteToDLQ(ctx context.Context, rec DeadLetterStepRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	rec.RoutedAt = time.Now()
	r.records = append(r.records, rec)

	if len(r.records) > r.maxRetained {
		r.records = r.records[len(r.records)-r.maxRetained:]
	}
	return nil
}

// GetRecords returns copies of currently buffered dead-letter items.
func (r *DeadLetterQueueRouter) GetRecords() []DeadLetterStepRecord {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]DeadLetterStepRecord, len(r.records))
	copy(out, r.records)
	return out
}

// Count returns the number of currently retained dead-letter items.
func (r *DeadLetterQueueRouter) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.records)
}
