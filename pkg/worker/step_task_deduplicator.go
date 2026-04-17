package worker

import (
	"sync"
	"time"
)

// TaskExecutionLock represents in-flight task execution idempotency token.
type TaskExecutionLock struct {
	TaskKey   string
	AcquiredAt time.Time
}

// StepTaskDeduplicator prevents redundant concurrent executions of identical tasks.
type StepTaskDeduplicator struct {
	mu     sync.Mutex
	locks  map[string]time.Time
	lockTTL time.Duration
}

// NewStepTaskDeduplicator creates a task execution deduplicator.
func NewStepTaskDeduplicator(ttl time.Duration) *StepTaskDeduplicator {
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	return &StepTaskDeduplicator{
		locks:   make(map[string]time.Time),
		lockTTL: ttl,
	}
}

// TryAcquire attempts to obtain execution rights for a task deduplication key.
func (d *StepTaskDeduplicator) TryAcquire(key string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now().UTC()
	if acquiredAt, exists := d.locks[key]; exists {
		if now.Sub(acquiredAt) < d.lockTTL {
			return false // already executing
		}
	}

	d.locks[key] = now
	return true
}

// Release yields the task execution lock.
func (d *StepTaskDeduplicator) Release(key string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.locks, key)
}

// IsExecuting checks if a task key is currently locked.
func (d *StepTaskDeduplicator) IsExecuting(key string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now().UTC()
	if acquiredAt, exists := d.locks[key]; exists {
		return now.Sub(acquiredAt) < d.lockTTL
	}
	return false
}
