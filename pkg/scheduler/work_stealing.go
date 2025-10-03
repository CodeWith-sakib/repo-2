package scheduler

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

// StealableTask represents a unit of schedulable work.
type StealableTask struct {
	ID       string
	TenantID string
	Priority int
	Payload  interface{}
}

// WorkDeque is a double-ended queue for a single worker thread/routine.
// The owner pushes/pops at the bottom (LIFO).
// Thieves steal from the top (FIFO).
type WorkDeque struct {
	mu     sync.Mutex
	tasks  []*StealableTask
	worker string
}

func newWorkDeque(worker string) *WorkDeque {
	return &WorkDeque{
		worker: worker,
	}
}

// PushBottom pushes a task onto the bottom of the deque (owner only).
func (d *WorkDeque) PushBottom(task *StealableTask) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.tasks = append(d.tasks, task)
}

// PopBottom pops a task from the bottom of the deque (owner only, LIFO).
func (d *WorkDeque) PopBottom() (*StealableTask, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if len(d.tasks) == 0 {
		return nil, false
	}

	idx := len(d.tasks) - 1
	task := d.tasks[idx]
	d.tasks = d.tasks[:idx]
	return task, true
}

// StealTop steals a task from the top of the deque (thief worker, FIFO).
func (d *WorkDeque) StealTop() (*StealableTask, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if len(d.tasks) == 0 {
		return nil, false
	}

	task := d.tasks[0]
	d.tasks = d.tasks[1:]
	return task, true
}

// Len returns the current number of tasks in the deque.
func (d *WorkDeque) Len() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.tasks)
}

// WorkStealingPool manages a pool of workers with work-stealing deques.
type WorkStealingPool struct {
	mu            sync.RWMutex
	deques        map[string]*WorkDeque
	workerIDs     []string
	rng           *rand.Rand
	rngMu         sync.Mutex
	tasksPushed   atomic.Int64
	tasksPopped   atomic.Int64
	tasksStolen   atomic.Int64
	stealAttempts atomic.Int64
}

// NewWorkStealingPool creates a pool with the given worker IDs.
func NewWorkStealingPool(workerIDs []string) (*WorkStealingPool, error) {
	if len(workerIDs) == 0 {
		return nil, fmt.Errorf("must provide at least one worker ID")
	}

	pool := &WorkStealingPool{
		deques:    make(map[string]*WorkDeque, len(workerIDs)),
		workerIDs: workerIDs,
		rng:       rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	for _, id := range workerIDs {
		pool.deques[id] = newWorkDeque(id)
	}

	return pool, nil
}

// Submit assigns a task to a worker's local deque.
func (p *WorkStealingPool) Submit(workerID string, task *StealableTask) error {
	p.mu.RLock()
	deque, ok := p.deques[workerID]
	p.mu.RUnlock()

	if !ok {
		return fmt.Errorf("unknown worker %q", workerID)
	}

	deque.PushBottom(task)
	p.tasksPushed.Add(1)
	return nil
}

// FetchWork gets work for workerID: first from local deque (bottom), otherwise attempts to steal from another worker.
func (p *WorkStealingPool) FetchWork(workerID string) (*StealableTask, bool) {
	p.mu.RLock()
	localDeque, ok := p.deques[workerID]
	p.mu.RUnlock()

	if !ok {
		return nil, false
	}

	// 1. Try local pop (LIFO)
	if task, found := localDeque.PopBottom(); found {
		p.tasksPopped.Add(1)
		return task, true
	}

	// 2. Try stealing from other workers (FIFO)
	p.stealAttempts.Add(1)
	p.mu.RLock()
	numWorkers := len(p.workerIDs)
	if numWorkers <= 1 {
		p.mu.RUnlock()
		return nil, false
	}

	// Random offset to avoid thundering herd on a single victim
	p.rngMu.Lock()
	offset := p.rng.Intn(numWorkers)
	p.rngMu.Unlock()

	for i := 0; i < numWorkers; i++ {
		targetID := p.workerIDs[(offset+i)%numWorkers]
		if targetID == workerID {
			continue
		}

		victim := p.deques[targetID]
		if task, stolen := victim.StealTop(); stolen {
			p.mu.RUnlock()
			p.tasksStolen.Add(1)
			return task, true
		}
	}
	p.mu.RUnlock()

	return nil, false
}

// TotalPending returns total pending tasks across all workers.
func (p *WorkStealingPool) TotalPending() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	total := 0
	for _, d := range p.deques {
		total += d.Len()
	}
	return total
}

// Stats returns counters for pushed, popped, and stolen tasks.
func (p *WorkStealingPool) Stats() (pushed, popped, stolen, stealAttempts int64) {
	return p.tasksPushed.Load(),
		p.tasksPopped.Load(),
		p.tasksStolen.Load(),
		p.stealAttempts.Load()
}
