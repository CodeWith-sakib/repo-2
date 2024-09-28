package memory

import (
	"context"
	"sort"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/storage"
)

func (s *Store) EnqueueTask(ctx context.Context, task *storage.QueuedTask) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if task.ID.IsEmpty() {
		task.ID = core.NewID("task")
	}
	if task.ScheduledAt.IsZero() {
		task.ScheduledAt = time.Now().UTC()
	}

	copyTask := *task
	s.queuedTasks[task.ID] = &copyTask
	return nil
}

func (s *Store) DequeueTasks(ctx context.Context, workerID string, limit int, leaseDuration time.Duration) ([]*storage.QueuedTask, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	candidates := make([]*storage.QueuedTask, 0)

	for _, task := range s.queuedTasks {
		// Ready if unassigned or lease expired, and scheduled time reached
		isUnassigned := task.LeaseWorker == ""
		isExpired := task.LeaseUntil != nil && task.LeaseUntil.Before(now)
		isTime := !task.ScheduledAt.After(now)

		if (isUnassigned || isExpired) && isTime {
			candidates = append(candidates, task)
		}
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].Priority != candidates[j].Priority {
			return candidates[i].Priority > candidates[j].Priority
		}
		return candidates[i].ScheduledAt.Before(candidates[j].ScheduledAt)
	})

	if limit > 0 && len(candidates) > limit {
		candidates = candidates[:limit]
	}

	result := make([]*storage.QueuedTask, 0, len(candidates))
	leaseEnd := now.Add(leaseDuration)
	for _, task := range candidates {
		task.LeaseWorker = workerID
		task.LeaseUntil = &leaseEnd

		copyTask := *task
		result = append(result, &copyTask)
	}

	return result, nil
}

func (s *Store) RenewLease(ctx context.Context, taskID core.ID, workerID string, extendBy time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.queuedTasks[taskID]
	if !exists {
		return core.ErrNotFound
	}
	if task.LeaseWorker != workerID {
		return core.ErrConflict
	}

	now := time.Now().UTC()
	if task.LeaseUntil != nil && task.LeaseUntil.Before(now) {
		return core.ErrLeaseExpired
	}

	newExpiry := now.Add(extendBy)
	task.LeaseUntil = &newExpiry
	return nil
}

func (s *Store) AckTask(ctx context.Context, taskID core.ID, workerID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.queuedTasks[taskID]
	if !exists {
		return core.ErrNotFound
	}
	if task.LeaseWorker != workerID {
		return core.ErrConflict
	}

	delete(s.queuedTasks, taskID)
	return nil
}

func (s *Store) NackTask(ctx context.Context, taskID core.ID, workerID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.queuedTasks[taskID]
	if !exists {
		return core.ErrNotFound
	}
	if task.LeaseWorker != workerID {
		return core.ErrConflict
	}

	task.LeaseWorker = ""
	task.LeaseUntil = nil
	task.Attempt++
	return nil
}

func (s *Store) RequeueOrphaned(ctx context.Context) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	count := 0
	for _, task := range s.queuedTasks {
		if task.LeaseWorker != "" && task.LeaseUntil != nil && task.LeaseUntil.Before(now) {
			task.LeaseWorker = ""
			task.LeaseUntil = nil
			task.Attempt++
			count++
		}
	}
	return count, nil
}

func (s *Store) AppendEvent(ctx context.Context, event *core.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if event.ID.IsEmpty() {
		event.ID = core.NewID("event")
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	copyEvent := *event
	s.events[event.RunID] = append(s.events[event.RunID], &copyEvent)
	return nil
}

func (s *Store) ListEvents(ctx context.Context, runID core.ID) ([]*core.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	events := s.events[runID]
	result := make([]*core.Event, len(events))
	for i, e := range events {
		copyEvent := *e
		result[i] = &copyEvent
	}
	return result, nil
}

// In-memory transaction wrapper
type memTx struct {
	*Store
}

func (tx *memTx) Commit(ctx context.Context) error {
	return nil
}

func (tx *memTx) Rollback(ctx context.Context) error {
	return nil
}

func (s *Store) BeginTx(ctx context.Context) (storage.Transaction, error) {
	return &memTx{Store: s}, nil
}
