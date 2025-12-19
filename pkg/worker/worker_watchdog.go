package worker

import (
	"sync"
	"time"
)

// WorkerLiveness represents the heartbeat and execution status of a registered worker.
type WorkerLiveness struct {
	WorkerID        string
	LastHeartbeatAt time.Time
	ActiveJobCount  int
	IsIsolated      bool
	IsolationReason string
}

// WorkerWatchdog monitors worker heartbeats and cordons unresponsive workers.
type WorkerWatchdog struct {
	mu           sync.RWMutex
	workers      map[string]*WorkerLiveness
	heartbeatTTL time.Duration
}

// NewWorkerWatchdog constructs a watchdog with heartbeats expiration limit.
func NewWorkerWatchdog(heartbeatTTL time.Duration) *WorkerWatchdog {
	if heartbeatTTL <= 0 {
		heartbeatTTL = 30 * time.Second
	}
	return &WorkerWatchdog{
		workers:      make(map[string]*WorkerLiveness),
		heartbeatTTL: heartbeatTTL,
	}
}

// RecordHeartbeat updates worker liveness timestamp and active job count.
func (w *WorkerWatchdog) RecordHeartbeat(workerID string, activeJobs int) {
	w.mu.Lock()
	defer w.mu.Unlock()

	wl, exists := w.workers[workerID]
	if !exists {
		wl = &WorkerLiveness{
			WorkerID: workerID,
		}
		w.workers[workerID] = wl
	}

	wl.LastHeartbeatAt = time.Now().UTC()
	wl.ActiveJobCount = activeJobs
	if wl.IsIsolated && wl.IsolationReason == "missed_heartbeats" {
		// Auto-recover on fresh heartbeat
		wl.IsIsolated = false
		wl.IsolationReason = ""
	}
}

// Sweep detects unresponsive workers exceeding heartbeat TTL and flags them as isolated.
func (w *WorkerWatchdog) Sweep(now time.Time) []string {
	w.mu.Lock()
	defer w.mu.Unlock()

	var cordoned []string
	cutoff := now.Add(-w.heartbeatTTL)

	for id, wl := range w.workers {
		if !wl.IsIsolated && wl.LastHeartbeatAt.Before(cutoff) {
			wl.IsIsolated = true
			wl.IsolationReason = "missed_heartbeats"
			cordoned = append(cordoned, id)
		}
	}

	return cordoned
}

// IsAvailable checks if worker is healthy and ready to accept new tasks.
func (w *WorkerWatchdog) IsAvailable(workerID string) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()

	wl, exists := w.workers[workerID]
	if !exists {
		return false
	}
	return !wl.IsIsolated
}
