package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/scheduler"
	"github.com/kestrelflow/kestrelflow/pkg/storage"
)

type Config struct {
	WorkerID          string
	Concurrency       int
	PollInterval      time.Duration
	LeaseDuration     time.Duration
	HeartbeatInterval time.Duration
}

type Pool struct {
	cfg       Config
	store     storage.EngineStore
	scheduler *scheduler.Scheduler
	registry  *ExecutorRegistry
	activeMu  sync.Mutex
	active    map[core.ID]context.CancelFunc
	taskWG    sync.WaitGroup
	stopCh    chan struct{}
	stopped   bool
}

func NewPool(cfg Config, store storage.EngineStore, sched *scheduler.Scheduler, reg *ExecutorRegistry) *Pool {
	if cfg.WorkerID == "" {
		cfg.WorkerID = string(core.NewID("worker"))
	}
	if cfg.Concurrency <= 0 {
		cfg.Concurrency = 4
	}
	if cfg.PollInterval <= 0 {
		cfg.PollInterval = 20 * time.Millisecond
	}
	if cfg.LeaseDuration <= 0 {
		cfg.LeaseDuration = 10 * time.Second
	}
	if cfg.HeartbeatInterval <= 0 {
		cfg.HeartbeatInterval = cfg.LeaseDuration / 3
	}

	return &Pool{
		cfg:       cfg,
		store:     store,
		scheduler: sched,
		registry:  reg,
		active:    make(map[core.ID]context.CancelFunc),
		stopCh:    make(chan struct{}),
	}
}

func (p *Pool) Start(ctx context.Context) {
	sem := make(chan struct{}, p.cfg.Concurrency)

	for {
		select {
		case <-ctx.Done():
			return
		case <-p.stopCh:
			return
		default:
		}

		// Acquire worker slot
		select {
		case sem <- struct{}{}:
		case <-ctx.Done():
			return
		case <-p.stopCh:
			return
		}

		tasks, err := p.store.DequeueTasks(ctx, p.cfg.WorkerID, 1, p.cfg.LeaseDuration)
		if err != nil || len(tasks) == 0 {
			<-sem
			time.Sleep(p.cfg.PollInterval)
			continue
		}

		task := tasks[0]
		p.taskWG.Add(1)

		go func(t *storage.QueuedTask) {
			defer func() {
				<-sem
				p.taskWG.Done()
			}()

			p.processTask(ctx, t)
		}(task)
	}
}

func (p *Pool) processTask(ctx context.Context, task *storage.QueuedTask) {
	taskCtx, cancel := context.WithCancel(ctx)

	p.activeMu.Lock()
	p.active[task.ID] = cancel
	p.activeMu.Unlock()

	defer func() {
		cancel()
		p.activeMu.Lock()
		delete(p.active, task.ID)
		p.activeMu.Unlock()
	}()

	// Heartbeat loop
	hbStop := make(chan struct{})
	defer close(hbStop)

	go func() {
		ticker := time.NewTicker(p.cfg.HeartbeatInterval)
		defer ticker.Stop()

		for {
			select {
			case <-hbStop:
				return
			case <-taskCtx.Done():
				return
			case <-ticker.C:
				_ = p.store.RenewLease(taskCtx, task.ID, p.cfg.WorkerID, p.cfg.LeaseDuration)
			}
		}
	}()

	// Fetch run and step definition
	run, err := p.store.GetRun(taskCtx, task.RunID)
	if err != nil {
		_ = p.store.NackTask(ctx, task.ID, p.cfg.WorkerID)
		return
	}

	wf, err := p.store.GetWorkflowVersion(taskCtx, run.WorkflowID, run.Version)
	if err != nil {
		_ = p.store.NackTask(ctx, task.ID, p.cfg.WorkerID)
		return
	}

	dag, err := core.BuildDAG(wf.Steps)
	if err != nil {
		_ = p.store.NackTask(ctx, task.ID, p.cfg.WorkerID)
		return
	}

	stepDef, exists := dag.GetNode(task.StepID)
	if !exists {
		_ = p.store.NackTask(ctx, task.ID, p.cfg.WorkerID)
		return
	}

	// Lookup executor
	exec, err := p.registry.Get(stepDef.TaskType)
	if err != nil {
		_ = p.store.AckTask(ctx, task.ID, p.cfg.WorkerID)
		_ = p.scheduler.HandleStepFailed(ctx, task.RunID, task.StepID, err.Error())
		return
	}

	sctx := StepContext{
		RunID:    string(task.RunID),
		StepID:   task.StepID,
		Attempt:  task.Attempt,
		WorkerID: p.cfg.WorkerID,
		Input:    stepDef.Config,
	}

	res, err := exec.Execute(taskCtx, sctx)
	_ = p.store.AckTask(ctx, task.ID, p.cfg.WorkerID)

	if err != nil {
		_ = p.scheduler.HandleStepFailed(ctx, task.RunID, task.StepID, err.Error())
		return
	}

	if res != nil && res.ErrorMessage != "" {
		_ = p.scheduler.HandleStepFailed(ctx, task.RunID, task.StepID, res.ErrorMessage)
		return
	}

	output := []byte("{}")
	if res != nil && len(res.Output) > 0 {
		output = res.Output
	}
	_ = p.scheduler.HandleStepCompleted(ctx, task.RunID, task.StepID, output)
}

func (p *Pool) CancelTask(taskID core.ID) bool {
	p.activeMu.Lock()
	defer p.activeMu.Unlock()

	cancel, ok := p.active[taskID]
	if ok {
		cancel()
		return true
	}
	return false
}

func (p *Pool) Stop() {
	p.activeMu.Lock()
	if p.stopped {
		p.activeMu.Unlock()
		return
	}
	p.stopped = true
	close(p.stopCh)
	p.activeMu.Unlock()

	p.taskWG.Wait()
}
