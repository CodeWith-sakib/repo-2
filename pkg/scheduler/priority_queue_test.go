package scheduler

import (
	"testing"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/storage"
)

func TestFairPriorityQueueOrder(t *testing.T) {
	fq := NewFairPriorityQueue()

	now := time.Now()
	tLow := &storage.QueuedTask{ID: "t-low", Priority: core.PriorityLow, ScheduledAt: now, TenantID: "t1"}
	tHigh := &storage.QueuedTask{ID: "t-high", Priority: core.PriorityHigh, ScheduledAt: now.Add(time.Second), TenantID: "t1"}
	tCritical := &storage.QueuedTask{ID: "t-crit", Priority: core.PriorityCritical, ScheduledAt: now.Add(2 * time.Second), TenantID: "t2"}

	fq.Push(tLow)
	fq.Push(tCritical)
	fq.Push(tHigh)

	if fq.Len() != 3 {
		t.Fatalf("expected 3 items, got %d", fq.Len())
	}
	if fq.TenantLoad("t1") != 2 {
		t.Errorf("expected tenant load 2 for t1, got %d", fq.TenantLoad("t1"))
	}

	// Pop order: Critical -> High -> Low
	first := fq.Pop()
	if first.ID != "t-crit" {
		t.Errorf("expected t-crit first, got %s", first.ID)
	}

	second := fq.Pop()
	if second.ID != "t-high" {
		t.Errorf("expected t-high second, got %s", second.ID)
	}

	third := fq.Pop()
	if third.ID != "t-low" {
		t.Errorf("expected t-low third, got %s", third.ID)
	}
}
