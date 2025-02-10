package scheduler

import (
	"testing"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

func TestMonitoredPriorityQueue(t *testing.T) {
	pq := NewPriorityQueue()
	mq := NewMonitoredPriorityQueue(pq)

	item1 := &QueueItem{ID: "task-1", Priority: core.PriorityHigh, ScheduledAt: time.Now()}
	item2 := &QueueItem{ID: "task-2", Priority: core.PriorityNormal, ScheduledAt: time.Now()}

	mq.Push(item1)
	mq.Push(item2)

	metrics := mq.GetMetrics()
	if metrics.TotalEnqueued != 2 || metrics.HighWatermark != 2 {
		t.Errorf("unexpected metrics: %+v", metrics)
	}

	popped := mq.Pop()
	if popped.ID != "task-1" {
		t.Errorf("expected task-1, got %s", popped.ID)
	}

	metrics = mq.GetMetrics()
	if metrics.TotalDequeued != 1 {
		t.Errorf("expected 1 dequeued, got %d", metrics.TotalDequeued)
	}
}
