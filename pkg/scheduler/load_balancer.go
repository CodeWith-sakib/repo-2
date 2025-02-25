package scheduler

import (
	"sync"
	"time"
)

type WorkerLoadStats struct {
	WorkerID      string
	Capacity      int
	ActiveTasks   int
	LastHeartbeat time.Time
}

type WorkerLoadBalancer struct {
	mu      sync.RWMutex
	workers map[string]*WorkerLoadStats
}

func NewWorkerLoadBalancer() *WorkerLoadBalancer {
	return &WorkerLoadBalancer{
		workers: make(map[string]*WorkerLoadStats),
	}
}

func (lb *WorkerLoadBalancer) RegisterWorker(id string, capacity int) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	lb.workers[id] = &WorkerLoadStats{
		WorkerID:      id,
		Capacity:      capacity,
		ActiveTasks:   0,
		LastHeartbeat: time.Now().UTC(),
	}
}

func (lb *WorkerLoadBalancer) RecordHeartbeat(id string) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	if w, exists := lb.workers[id]; exists {
		w.LastHeartbeat = time.Now().UTC()
	}
}

func (lb *WorkerLoadBalancer) SelectWorker() (string, bool) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	var bestWorker string
	lowestLoadRatio := 1.0
	found := false
	cutoff := time.Now().UTC().Add(-30 * time.Second)

	for id, w := range lb.workers {
		if w.LastHeartbeat.Before(cutoff) {
			continue // Dead worker
		}
		if w.Capacity <= 0 {
			continue
		}
		ratio := float64(w.ActiveTasks) / float64(w.Capacity)
		if ratio < 1.0 && (!found || ratio < lowestLoadRatio) {
			lowestLoadRatio = ratio
			bestWorker = id
			found = true
		}
	}

	if found {
		lb.workers[bestWorker].ActiveTasks++
		return bestWorker, true
	}
	return "", false
}

func (lb *WorkerLoadBalancer) ReleaseTask(workerID string) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	if w, exists := lb.workers[workerID]; exists && w.ActiveTasks > 0 {
		w.ActiveTasks--
	}
}
