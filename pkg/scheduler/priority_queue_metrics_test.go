package scheduler

import (
	"testing"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/storage"
)

func TestMonitoredPriorityQueue(t *testing.T) {
	fq := NewFairPriorityQueue()
	mq := NewMonitoredPriorityQueue(fq)

	task1 := &storage.QueuedTask{ID: core.NewID("task-1"), Priority: core.PriorityHigh, ScheduledAt: time.Now()}
	task2 := &storage.QueuedTask{ID: core.NewID("task-2"), Priority: core.PriorityNormal, ScheduledAt: time.Now()}

	mq.Push(task1)
	mq.Push(task2)

	metrics := mq.GetMetrics()
	if metrics.TotalEnqueued != 2 || metrics.HighWatermark != 2 {
		t.Errorf("unexpected metrics: %+v", metrics)
	}

	popped := mq.Pop()
	if popped.ID != task1.ID {
		t.Errorf("expected task-1, got %s", popped.ID)
	}

	metrics = mq.GetMetrics()
	if metrics.TotalDequeued != 1 {
		t.Errorf("expected 1 dequeued, got %d", metrics.TotalDequeued)
	}
}
