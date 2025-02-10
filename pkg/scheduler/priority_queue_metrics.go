package scheduler

import (
	"sync"
	"time"
)

type QueueMetrics struct {
	TotalEnqueued   int64
	TotalDequeued   int64
	TotalExpired    int64
	HighWatermark   int
	AvgWaitDuration time.Duration
}

type MonitoredPriorityQueue struct {
	pq              *PriorityQueue
	mu              sync.RWMutex
	metrics         QueueMetrics
	enqueueTimes    map[string]time.Time
	totalWaitNanos  int64
	waitedTaskCount int64
}

func NewMonitoredPriorityQueue(pq *PriorityQueue) *MonitoredPriorityQueue {
	return &MonitoredPriorityQueue{
		pq:           pq,
		enqueueTimes: make(map[string]time.Time),
	}
}

func (mq *MonitoredPriorityQueue) Push(item *QueueItem) {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	mq.metrics.TotalEnqueued++
	mq.enqueueTimes[item.ID] = time.Now()
	mq.pq.Push(item)

	if mq.pq.Len() > mq.metrics.HighWatermark {
		mq.metrics.HighWatermark = mq.pq.Len()
	}
}

func (mq *MonitoredPriorityQueue) Pop() *QueueItem {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	item := mq.pq.Pop()
	if item == nil {
		return nil
	}

	mq.metrics.TotalDequeued++
	if enqTime, exists := mq.enqueueTimes[item.ID]; exists {
		delete(mq.enqueueTimes, item.ID)
		wait := time.Since(enqTime)
		mq.totalWaitNanos += wait.Nanoseconds()
		mq.waitedTaskCount++
		mq.metrics.AvgWaitDuration = time.Duration(mq.totalWaitNanos / mq.waitedTaskCount)
	}

	return item
}

func (mq *MonitoredPriorityQueue) GetMetrics() QueueMetrics {
	mq.mu.RLock()
	defer mq.mu.RUnlock()
	return mq.metrics
}
