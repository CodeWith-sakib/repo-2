package core

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrDuplicateIdempotencyKey = errors.New("duplicate idempotency key encountered")
)

// IdempotencyRecord tracks workflow dispatch status for a specific idempotency key.
type IdempotencyRecord struct {
	Key        string    `json:"key"`
	WorkflowID string    `json:"workflow_id"`
	CreatedAt  time.Time `json:"created_at"`
	Status     string    `json:"status"`
}

// IdempotencyKeyEnforcer prevents duplicate workflow launches with identical client tokens.
type IdempotencyKeyEnforcer struct {
	mu      sync.RWMutex
	records map[string]IdempotencyRecord
	ttl     time.Duration
}

// NewIdempotencyKeyEnforcer initializes an idempotency controller.
func NewIdempotencyKeyEnforcer(ttl time.Duration) *IdempotencyKeyEnforcer {
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	return &IdempotencyKeyEnforcer{
		records: make(map[string]IdempotencyRecord),
		ttl:     ttl,
	}
}

// RegisterOrCheck acquires ownership of a key or returns existing workflow ID if previously registered.
func (e *IdempotencyKeyEnforcer) RegisterOrCheck(ctx context.Context, key, workflowID string) (string, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	now := time.Now()
	if rec, ok := e.records[key]; ok {
		if now.Sub(rec.CreatedAt) <= e.ttl {
			return rec.WorkflowID, ErrDuplicateIdempotencyKey
		}
	}

	e.records[key] = IdempotencyRecord{
		Key:        key,
		WorkflowID: workflowID,
		CreatedAt:  now,
		Status:     "SUBMITTED",
	}

	return workflowID, nil
}

// PurgeExpired cleans up idempotency records older than configured TTL.
func (e *IdempotencyKeyEnforcer) PurgeExpired(now time.Time) int {
	e.mu.Lock()
	defer e.mu.Unlock()

	purged := 0
	for k, rec := range e.records {
		if now.Sub(rec.CreatedAt) > e.ttl {
			delete(e.records, k)
			purged++
		}
	}
	return purged
}
