package events

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

// DLQEventRecord contains failed event metadata, original error, and delivery attempts count.
type DLQEventRecord struct {
	Event            *core.Event `json:"event"`
	OriginalError    string      `json:"original_error"`
	DeliveryAttempts int         `json:"delivery_attempts"`
	RoutedAt         time.Time   `json:"routed_at"`
}

// DLQEventRouter captures undeliverable events and routes them into dead-letter storage.
type DLQEventRouter struct {
	mu      sync.Mutex
	records []DLQEventRecord
	maxSize int
}

// NewDLQEventRouter creates a dead-letter queue router.
func NewDLQEventRouter(maxSize int) *DLQEventRouter {
	if maxSize <= 0 {
		maxSize = 1000
	}
	return &DLQEventRouter{
		records: make([]DLQEventRecord, 0, maxSize),
		maxSize: maxSize,
	}
}

// RouteToDLQ routes a failed event with error reason into the dead-letter queue.
func (r *DLQEventRouter) RouteToDLQ(ctx context.Context, event *core.Event, err error, attempts int) error {
	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	rec := DLQEventRecord{
		Event:            event,
		OriginalError:    err.Error(),
		DeliveryAttempts: attempts,
		RoutedAt:         time.Now().UTC(),
	}

	if len(r.records) >= r.maxSize {
		r.records = r.records[1:] // evict oldest
	}
	r.records = append(r.records, rec)
	return nil
}

// PendingCount returns total stored dead-letter records.
func (r *DLQEventRouter) PendingCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.records)
}

// PopBatch extracts and clears up to batchSize dead-letter events for reprocessing.
func (r *DLQEventRouter) PopBatch(batchSize int) []DLQEventRecord {
	r.mu.Lock()
	defer r.mu.Unlock()

	n := len(r.records)
	if batchSize <= 0 || batchSize > n {
		batchSize = n
	}

	batch := make([]DLQEventRecord, batchSize)
	copy(batch, r.records[:batchSize])
	r.records = r.records[batchSize:]
	return batch
}
