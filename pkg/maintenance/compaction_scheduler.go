package maintenance

import (
	"fmt"
	"sync"
	"time"
)

// CompactionJobStatus describes the current state of a compaction job.
type CompactionJobStatus string

const (
	CompactionPending   CompactionJobStatus = "pending"
	CompactionRunning   CompactionJobStatus = "running"
	CompactionCompleted CompactionJobStatus = "completed"
	CompactionFailed    CompactionJobStatus = "failed"
)

// CompactionJob describes a scheduled data compaction task.
type CompactionJob struct {
	ID          string
	TableName   string
	ShardID     int
	Priority    int
	ScheduledAt time.Time
	StartedAt   time.Time
	CompletedAt time.Time
	Status      CompactionJobStatus
	BytesBefore uint64
	BytesAfter  uint64
	ErrMsg      string
}

// CompactionResult summarizes the outcome of a compaction run.
type CompactionResult struct {
	JobID       string
	BytesBefore uint64
	BytesAfter  uint64
	Duration    time.Duration
	Err         error
}

// CompactionRunnerFn is the function called to perform the actual compaction.
type CompactionRunnerFn func(job *CompactionJob) CompactionResult

// CompactionScheduler manages a priority queue of pending compaction jobs
// and dispatches them to registered runner functions.
type CompactionScheduler struct {
	mu         sync.Mutex
	jobs       []*CompactionJob
	runners    map[string]CompactionRunnerFn
	maxWorkers int
	sem        chan struct{}
	nextID     int
	history    []*CompactionJob
}

// NewCompactionScheduler creates a scheduler with the given worker concurrency.
func NewCompactionScheduler(maxWorkers int) *CompactionScheduler {
	return &CompactionScheduler{
		runners:    make(map[string]CompactionRunnerFn),
		maxWorkers: maxWorkers,
		sem:        make(chan struct{}, maxWorkers),
	}
}

// RegisterRunner associates a runner function with a table name.
func (s *CompactionScheduler) RegisterRunner(tableName string, fn CompactionRunnerFn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.runners[tableName] = fn
}

// Enqueue adds a compaction job to the priority queue.
func (s *CompactionScheduler) Enqueue(tableName string, shardID, priority int) (*CompactionJob, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.runners[tableName]; !ok {
		return nil, fmt.Errorf("no runner registered for table %q", tableName)
	}

	s.nextID++
	job := &CompactionJob{
		ID:          fmt.Sprintf("cjob-%04d", s.nextID),
		TableName:   tableName,
		ShardID:     shardID,
		Priority:    priority,
		ScheduledAt: time.Now(),
		Status:      CompactionPending,
	}

	// Priority insert (descending)
	inserted := false
	for i, existing := range s.jobs {
		if priority > existing.Priority {
			tail := make([]*CompactionJob, len(s.jobs[i:]))
			copy(tail, s.jobs[i:])
			s.jobs = append(s.jobs[:i], job)
			s.jobs = append(s.jobs, tail...)
			inserted = true
			break
		}
	}
	if !inserted {
		s.jobs = append(s.jobs, job)
	}

	return job, nil
}

// Drain executes all pending jobs synchronously (for testing or forced flush).
func (s *CompactionScheduler) Drain() []*CompactionJob {
	s.mu.Lock()
	pending := make([]*CompactionJob, len(s.jobs))
	copy(pending, s.jobs)
	s.jobs = s.jobs[:0]
	runners := make(map[string]CompactionRunnerFn, len(s.runners))
	for k, v := range s.runners {
		runners[k] = v
	}
	s.mu.Unlock()

	var wg sync.WaitGroup
	for _, job := range pending {
		job := job
		runner := runners[job.TableName]
		s.sem <- struct{}{}
		wg.Add(1)
		go func() {
			defer func() {
				<-s.sem
				wg.Done()
			}()
			job.Status = CompactionRunning
			job.StartedAt = time.Now()
			result := runner(job)
			job.BytesBefore = result.BytesBefore
			job.BytesAfter = result.BytesAfter
			job.CompletedAt = time.Now()
			if result.Err != nil {
				job.Status = CompactionFailed
				job.ErrMsg = result.Err.Error()
			} else {
				job.Status = CompactionCompleted
			}
		}()
	}
	wg.Wait()

	s.mu.Lock()
	s.history = append(s.history, pending...)
	s.mu.Unlock()

	return pending
}

// PendingCount returns the number of queued jobs not yet executed.
func (s *CompactionScheduler) PendingCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.jobs)
}

// History returns the execution history of completed jobs.
func (s *CompactionScheduler) History() []*CompactionJob {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]*CompactionJob, len(s.history))
	copy(out, s.history)
	return out
}

// Stats returns a diagnostic summary of the scheduler state.
func (s *CompactionScheduler) Stats() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return fmt.Sprintf("compaction_scheduler: pending=%d history=%d maxWorkers=%d",
		len(s.jobs), len(s.history), s.maxWorkers)
}
