package scheduler

import (
	"sync"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/storage"
)

type QueueMetrics struct {
	TotalEnqueued   int64
	TotalDequeued   int64
	TotalExpired    int64
	HighWatermark   int
	AvgWaitDuration time.Duration
}

type MonitoredPriorityQueue struct {
	fq              *FairPriorityQueue
	mu              sync.RWMutex
	metrics         QueueMetrics
	enqueueTimes    map[core.ID]time.Time
	totalWaitNanos  int64
	waitedTaskCount int64
}

func NewMonitoredPriorityQueue(fq *FairPriorityQueue) *MonitoredPriorityQueue {
	return &MonitoredPriorityQueue{
		fq:           fq,
		enqueueTimes: make(map[core.ID]time.Time),
	}
}

func (mq *MonitoredPriorityQueue) Push(task *storage.QueuedTask) {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	mq.metrics.TotalEnqueued++
	mq.enqueueTimes[task.ID] = time.Now()
	mq.fq.Push(task)

	if mq.fq.Len() > mq.metrics.HighWatermark {
		mq.metrics.HighWatermark = mq.fq.Len()
	}
}

func (mq *MonitoredPriorityQueue) Pop() *storage.QueuedTask {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	task := mq.fq.Pop()
	if task == nil {
		return nil
	}

	mq.metrics.TotalDequeued++
	if enqTime, exists := mq.enqueueTimes[task.ID]; exists {
		delete(mq.enqueueTimes, task.ID)
		wait := time.Since(enqTime)
		mq.totalWaitNanos += wait.Nanoseconds()
		mq.waitedTaskCount++
		mq.metrics.AvgWaitDuration = time.Duration(mq.totalWaitNanos / mq.waitedTaskCount)
	}

	return task
}

func (mq *MonitoredPriorityQueue) GetMetrics() QueueMetrics {
	mq.mu.RLock()
	defer mq.mu.RUnlock()
	return mq.metrics
}
