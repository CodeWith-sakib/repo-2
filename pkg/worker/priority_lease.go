package worker

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// PreemptiveLease represents an acquired execution lock on a task.
type PreemptiveLease struct {
	TaskID     string
	WorkerID   string
	Priority   int
	AcquiredAt time.Time
	ExpiresAt  time.Time
	cancelFn   context.CancelFunc
	Preempted  bool
}

// PreemptionConfig configures thresholds for preempting lower-priority work.
type PreemptionConfig struct {
	DefaultTTL         time.Duration
	MinPreemptDelta    int // e.g. priority must be at least 20 points higher to preempt
	PreemptGracePeriod time.Duration
}

// DefaultPreemptionConfig returns production default settings.
func DefaultPreemptionConfig() PreemptionConfig {
	return PreemptionConfig{
		DefaultTTL:         30 * time.Second,
		MinPreemptDelta:    20,
		PreemptGracePeriod: 5 * time.Second,
	}
}

// PriorityLeaseCoordinator manages preemptive execution locks on tasks.
type PriorityLeaseCoordinator struct {
	mu           sync.RWMutex
	leases       map[string]*PreemptiveLease // taskID -> lease
	cfg          PreemptionConfig
	preemptCount atomic.Int64
}

// NewPriorityLeaseCoordinator creates a coordinator with the given config.
func NewPriorityLeaseCoordinator(cfg PreemptionConfig) *PriorityLeaseCoordinator {
	return &PriorityLeaseCoordinator{
		leases: make(map[string]*PreemptiveLease),
		cfg:    cfg,
	}
}

// AcquireLease attempts to lock a task. If locked by lower priority, preempts it.
func (c *PriorityLeaseCoordinator) AcquireLease(parentCtx context.Context, taskID, workerID string, priority int) (*PreemptiveLease, context.Context, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	existing, exists := c.leases[taskID]

	if exists && now.Before(existing.ExpiresAt) && !existing.Preempted {
		// Existing active lease
		if existing.WorkerID == workerID {
			// Renewal by same worker
			existing.ExpiresAt = now.Add(c.cfg.DefaultTTL)
			ctx, cancel := context.WithCancel(parentCtx)
			existing.cancelFn = cancel
			return existing, ctx, nil
		}

		// Can candidate preempt existing?
		if priority-existing.Priority < c.cfg.MinPreemptDelta {
			return nil, nil, fmt.Errorf("task %s already leased by worker %s at priority %d", taskID, existing.WorkerID, existing.Priority)
		}

		// Preempt existing!
		existing.Preempted = true
		if existing.cancelFn != nil {
			existing.cancelFn() // signal cancellation to preempted worker
		}
		c.preemptCount.Add(1)
	}

	leaseCtx, cancel := context.WithCancel(parentCtx)
	lease := &PreemptiveLease{
		TaskID:     taskID,
		WorkerID:   workerID,
		Priority:   priority,
		AcquiredAt: now,
		ExpiresAt:  now.Add(c.cfg.DefaultTTL),
		cancelFn:   cancel,
		Preempted:  false,
	}

	c.leases[taskID] = lease
	return lease, leaseCtx, nil
}

// ReleaseLease unlocks a task if held by the given worker.
func (c *PriorityLeaseCoordinator) ReleaseLease(taskID, workerID string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if lease, exists := c.leases[taskID]; exists {
		if lease.WorkerID == workerID {
			if lease.cancelFn != nil {
				lease.cancelFn()
			}
			delete(c.leases, taskID)
			return true
		}
	}
	return false
}

// PreemptionsCount returns the number of preemption events recorded.
func (c *PriorityLeaseCoordinator) PreemptionsCount() int64 {
	return c.preemptCount.Load()
}
