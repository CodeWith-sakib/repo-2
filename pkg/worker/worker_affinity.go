package worker

import (
	"hash/fnv"
	"sort"
	"sync"
	"time"
)

// AffinityMapping records which worker holds warm cache for a workflow run.
type AffinityMapping struct {
	RunID      string
	WorkerID   string
	AssignedAt time.Time
	LastUsedAt time.Time
	WarmScore  int // higher score = more cached artifacts on this worker
}

// WorkerAffinityRouter routes tasks to workers with existing warmed state.
type WorkerAffinityRouter struct {
	mu          sync.RWMutex
	affinity    map[string]*AffinityMapping // runID -> mapping
	workerNodes []string
	ttl         time.Duration
}

// NewWorkerAffinityRouter creates an affinity router with mapping TTL.
func NewWorkerAffinityRouter(workers []string, ttl time.Duration) *WorkerAffinityRouter {
	if ttl <= 0 {
		ttl = 30 * time.Minute
	}
	sortedWorkers := make([]string, len(workers))
	copy(sortedWorkers, workers)
	sort.Strings(sortedWorkers)

	return &WorkerAffinityRouter{
		affinity:    make(map[string]*AffinityMapping),
		workerNodes: sortedWorkers,
		ttl:         ttl,
	}
}

// UpdateWorkers sets the active pool of worker node IDs.
func (r *WorkerAffinityRouter) UpdateWorkers(workers []string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	sorted := make([]string, len(workers))
	copy(sorted, workers)
	sort.Strings(sorted)
	r.workerNodes = sorted
}

// GetOrAssignWorker returns the preferred warm worker for a run, or assigns one via consistent hashing.
func (r *WorkerAffinityRouter) GetOrAssignWorker(runID string) (string, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	if mapping, exists := r.affinity[runID]; exists {
		if now.Sub(mapping.LastUsedAt) < r.ttl && r.isWorkerActive(mapping.WorkerID) {
			mapping.LastUsedAt = now
			mapping.WarmScore++
			return mapping.WorkerID, true // warm hit!
		}
		// Expired or worker inactive
		delete(r.affinity, runID)
	}

	if len(r.workerNodes) == 0 {
		return "", false
	}

	// Consistent hash assignment
	worker := r.hashToWorker(runID)
	r.affinity[runID] = &AffinityMapping{
		RunID:      runID,
		WorkerID:   worker,
		AssignedAt: now,
		LastUsedAt: now,
		WarmScore:  1,
	}

	return worker, false
}

// Invalidate clears affinity for a run.
func (r *WorkerAffinityRouter) Invalidate(runID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.affinity, runID)
}

func (r *WorkerAffinityRouter) isWorkerActive(workerID string) bool {
	for _, w := range r.workerNodes {
		if w == workerID {
			return true
		}
	}
	return false
}

func (r *WorkerAffinityRouter) hashToWorker(key string) string {
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	idx := int(h.Sum32()) % len(r.workerNodes)
	if idx < 0 {
		idx = -idx
	}
	return r.workerNodes[idx]
}

// ActiveAffinityCount returns the number of active cached run mappings.
func (r *WorkerAffinityRouter) ActiveAffinityCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.affinity)
}

// PruneExpired evicts mappings older than TTL.
func (r *WorkerAffinityRouter) PruneExpired() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	pruned := 0
	for k, v := range r.affinity {
		if now.Sub(v.LastUsedAt) >= r.ttl {
			delete(r.affinity, k)
			pruned++
		}
	}
	return pruned
}
