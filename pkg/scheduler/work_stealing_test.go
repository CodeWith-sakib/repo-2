package scheduler

import (
	"fmt"
	"testing"
)

func TestWorkStealingPool_LocalAndSteal(t *testing.T) {
	workers := []string{"w1", "w2", "w3"}
	pool, err := NewWorkStealingPool(workers)
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}

	// Push 3 tasks to w1
	for i := 1; i <= 3; i++ {
		task := &StealableTask{ID: fmt.Sprintf("t-%d", i), Priority: i}
		if err := pool.Submit("w1", task); err != nil {
			t.Fatalf("submit failed: %v", err)
		}
	}

	if pool.TotalPending() != 3 {
		t.Fatalf("expected 3 pending, got %d", pool.TotalPending())
	}

	// w1 fetches locally (LIFO -> should get t-3 first)
	t3, ok := pool.FetchWork("w1")
	if !ok || t3.ID != "t-3" {
		t.Errorf("w1 expected t-3 locally, got %+v (ok=%v)", t3, ok)
	}

	// w2 is idle, fetches work -> steals from w1 (FIFO from top -> should get t-1!)
	t1, ok := pool.FetchWork("w2")
	if !ok || t1.ID != "t-1" {
		t.Errorf("w2 expected to steal t-1, got %+v (ok=%v)", t1, ok)
	}

	// w3 is idle -> steals remaining task t-2 from w1
	t2, ok := pool.FetchWork("w3")
	if !ok || t2.ID != "t-2" {
		t.Errorf("w3 expected to steal t-2, got %+v (ok=%v)", t2, ok)
	}

	// Now pool should be empty
	if pool.TotalPending() != 0 {
		t.Errorf("expected 0 pending, got %d", pool.TotalPending())
	}

	pushed, popped, stolen, _ := pool.Stats()
	if pushed != 3 || popped != 1 || stolen != 2 {
		t.Errorf("stats mismatch: pushed=%d popped=%d stolen=%d", pushed, popped, stolen)
	}
}
