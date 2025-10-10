package worker

import (
	"fmt"
	"sync"
)

// QueuedTask represents a unit of work with an associated resource cost.
type QueuedTask struct {
	ID       string
	TenantID string
	Cost     int
	Payload  interface{}
}

// tenantQueue holds buffered tasks and DRR credit deficit for one tenant.
type tenantQueue struct {
	tenantID string
	tasks    []*QueuedTask
	deficit  int
	quantum  int
	inTurn   bool
}

// DeficitRoundRobinQueue arbitrates multi-tenant worker task dispatches with DRR fairness.
type DeficitRoundRobinQueue struct {
	mu             sync.Mutex
	queues         map[string]*tenantQueue
	activeTenants  []string // list of tenant IDs with non-empty queues
	defaultQuantum int
	totalPending   int
}

// NewDeficitRoundRobinQueue creates a new DRR queue.
func NewDeficitRoundRobinQueue(defaultQuantum int) *DeficitRoundRobinQueue {
	if defaultQuantum <= 0 {
		defaultQuantum = 100
	}
	return &DeficitRoundRobinQueue{
		queues:         make(map[string]*tenantQueue),
		defaultQuantum: defaultQuantum,
	}
}

// SetTenantQuantum customizes the quantum (service allowance per round) for a tenant.
func (q *DeficitRoundRobinQueue) SetTenantQuantum(tenantID string, quantum int) {
	q.mu.Lock()
	defer q.mu.Unlock()

	tq, exists := q.queues[tenantID]
	if !exists {
		tq = &tenantQueue{
			tenantID: tenantID,
			quantum:  quantum,
		}
		q.queues[tenantID] = tq
	} else {
		tq.quantum = quantum
	}
}

// Enqueue adds a task into the corresponding tenant's queue.
func (q *DeficitRoundRobinQueue) Enqueue(task *QueuedTask) error {
	if task == nil || task.TenantID == "" {
		return fmt.Errorf("invalid task: nil or empty tenant ID")
	}
	if task.Cost <= 0 {
		task.Cost = 1
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	tq, exists := q.queues[task.TenantID]
	if !exists {
		tq = &tenantQueue{
			tenantID: task.TenantID,
			quantum:  q.defaultQuantum,
		}
		q.queues[task.TenantID] = tq
	}

	wasEmpty := len(tq.tasks) == 0
	tq.tasks = append(tq.tasks, task)
	q.totalPending++

	if wasEmpty {
		q.activeTenants = append(q.activeTenants, task.TenantID)
	}

	return nil
}

// Dequeue extracts the next fair task using DRR round-robin arbitration.
func (q *DeficitRoundRobinQueue) Dequeue() (*QueuedTask, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for len(q.activeTenants) > 0 {
		currentTenantID := q.activeTenants[0]
		tq := q.queues[currentTenantID]

		// Only add quantum if beginning a new turn for this tenant
		if !tq.inTurn {
			tq.deficit += tq.quantum
			tq.inTurn = true
		}

		if len(tq.tasks) == 0 {
			// No tasks left for this tenant
			tq.deficit = 0
			tq.inTurn = false
			q.activeTenants = q.activeTenants[1:]
			continue
		}

		head := tq.tasks[0]
		if head.Cost <= tq.deficit {
			tq.deficit -= head.Cost
			tq.tasks = tq.tasks[1:]
			q.totalPending--

			if len(tq.tasks) == 0 {
				tq.deficit = 0
				tq.inTurn = false
				q.activeTenants = q.activeTenants[1:]
			} else if tq.tasks[0].Cost > tq.deficit {
				// Turn finished: rotate to back of queue
				tq.inTurn = false
				q.activeTenants = append(q.activeTenants[1:], currentTenantID)
			}

			return head, true
		}

		// Head cost exceeds deficit even after adding quantum -> rotate
		tq.inTurn = false
		q.activeTenants = append(q.activeTenants[1:], currentTenantID)
	}

	return nil, false
}

// PendingCount returns total tasks queued across all tenants.
func (q *DeficitRoundRobinQueue) PendingCount() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.totalPending
}

// ActiveTenantCount returns the number of tenants currently holding backlogged tasks.
func (q *DeficitRoundRobinQueue) ActiveTenantCount() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.activeTenants)
}
