package worker

import (
	"sync"
	"time"
)

type FlightRecord struct {
	TaskID    string
	WorkerID  string
	StartedAt time.Time
}

type TaskFlightTracker struct {
	mu      sync.RWMutex
	inflight map[string]*FlightRecord
}

func NewTaskFlightTracker() *TaskFlightTracker {
	return &TaskFlightTracker{
		inflight: make(map[string]*FlightRecord),
	}
}

func (t *TaskFlightTracker) Track(taskID, workerID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.inflight[taskID] = &FlightRecord{
		TaskID:    taskID,
		WorkerID:  workerID,
		StartedAt: time.Now().UTC(),
	}
}

func (t *TaskFlightTracker) Complete(taskID string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if _, exists := t.inflight[taskID]; exists {
		delete(t.inflight, taskID)
		return true
	}
	return false
}

func (t *TaskFlightTracker) InFlightCount() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.inflight)
}

func (t *TaskFlightTracker) ActiveTasksForWorker(workerID string) []string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var tasks []string
	for id, rec := range t.inflight {
		if rec.WorkerID == workerID {
			tasks = append(tasks, id)
		}
	}
	return tasks
}
