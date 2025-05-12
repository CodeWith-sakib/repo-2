package plugin

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type PluginTaskJob struct {
	TaskType string
	Payload  []byte
	ResultCh chan PluginTaskResult
}

type PluginTaskResult struct {
	Output []byte
	Error  error
}

type PluginExecutorPool struct {
	registry *Registry
	workers  int
	jobQueue chan PluginTaskJob
	stopCh   chan struct{}
	wg       sync.WaitGroup
}

func NewPluginExecutorPool(registry *Registry, workers int, queueSize int) *PluginExecutorPool {
	if workers <= 0 {
		workers = 4
	}
	if queueSize <= 0 {
		queueSize = 100
	}
	p := &PluginExecutorPool{
		registry: registry,
		workers:  workers,
		jobQueue: make(chan PluginTaskJob, queueSize),
		stopCh:   make(chan struct{}),
	}
	p.start()
	return p
}

func (p *PluginExecutorPool) start() {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.workerLoop()
	}
}

func (p *PluginExecutorPool) workerLoop() {
	defer p.wg.Done()

	for {
		select {
		case <-p.stopCh:
			return
		case job, ok := <-p.jobQueue:
			if !ok {
				return
			}
			p.executeJob(job)
		}
	}
}

func (p *PluginExecutorPool) executeJob(job PluginTaskJob) {
	handler, exists := p.registry.Get(job.TaskType)
	if !exists {
		job.ResultCh <- PluginTaskResult{
			Error: fmt.Errorf("unregistered task type: %s", job.TaskType),
		}
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	output, err := handler.Execute(ctx, job.Payload)
	job.ResultCh <- PluginTaskResult{
		Output: output,
		Error:  err,
	}
}

func (p *PluginExecutorPool) Submit(job PluginTaskJob) bool {
	select {
	case p.jobQueue <- job:
		return true
	default:
		return false
	}
}

func (p *PluginExecutorPool) Shutdown() {
	close(p.stopCh)
	p.wg.Wait()
}
