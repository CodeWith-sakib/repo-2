package worker

import (
	"fmt"
	"testing"
)

func TestDeficitRoundRobinQueue_FairInterleaving(t *testing.T) {
	// Quantum = 10
	q := NewDeficitRoundRobinQueue(10)

	// Tenant A enqueues 3 tasks of cost 5
	for i := 1; i <= 3; i++ {
		_ = q.Enqueue(&QueuedTask{
			ID:       fmt.Sprintf("A-%d", i),
			TenantID: "tenant-A",
			Cost:     5,
		})
	}

	// Tenant B enqueues 3 tasks of cost 10
	for i := 1; i <= 3; i++ {
		_ = q.Enqueue(&QueuedTask{
			ID:       fmt.Sprintf("B-%d", i),
			TenantID: "tenant-B",
			Cost:     10,
		})
	}

	if q.PendingCount() != 6 {
		t.Fatalf("expected 6 pending, got %d", q.PendingCount())
	}

	var dequeued []string
	for {
		task, ok := q.Dequeue()
		if !ok {
			break
		}
		dequeued = append(dequeued, task.ID)
	}

	if len(dequeued) != 6 {
		t.Fatalf("expected 6 dequeued, got %d: %v", len(dequeued), dequeued)
	}

	// Tenant A has quantum 10 and cost 5 -> can do A-1 (cost 5, rem 5) and A-2 (cost 5, rem 0)
	// Tenant B has quantum 10 and cost 10 -> can do B-1 (cost 10, rem 0)
	// Then round 2: A-3, then B-2, etc.
	expectedFirstThree := []string{"A-1", "A-2", "B-1"}
	for i := 0; i < 3; i++ {
		if dequeued[i] != expectedFirstThree[i] {
			t.Errorf("pos %d: got %s, want %s", i, dequeued[i], expectedFirstThree[i])
		}
	}

	if q.PendingCount() != 0 {
		t.Errorf("expected 0 pending, got %d", q.PendingCount())
	}
}

func TestDeficitRoundRobinQueue_Empty(t *testing.T) {
	q := NewDeficitRoundRobinQueue(50)
	task, ok := q.Dequeue()
	if ok || task != nil {
		t.Error("expected false on empty dequeue")
	}
}
