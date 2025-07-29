package scheduler

import (
	"testing"
)

func TestDeficitRoundRobinScheduler(t *testing.T) {
	sched := NewDeficitRoundRobinScheduler(50)

	sched.Enqueue("t1", "item-1")
	sched.Enqueue("t2", "item-2")

	// Cost 40 is less than quantum 50, should round robin alternate
	i1, ok1 := sched.Dequeue(40)
	if !ok1 || i1 != "item-1" {
		t.Errorf("expected item-1, got %v", i1)
	}

	i2, ok2 := sched.Dequeue(40)
	if !ok2 || i2 != "item-2" {
		t.Errorf("expected item-2, got %v", i2)
	}
}
