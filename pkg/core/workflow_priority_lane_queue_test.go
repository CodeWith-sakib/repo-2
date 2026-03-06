package core

import (
	"testing"
)

func TestPriorityLaneQueue(t *testing.T) {
	q := NewPriorityLaneQueue()

	q.Enqueue("run-low", LaneLow)
	q.Enqueue("run-normal", LaneNormal)
	q.Enqueue("run-crit", LaneCritical)
	q.Enqueue("run-high", LaneHigh)

	if q.Len() != 4 {
		t.Fatalf("expected 4 items in queue, got %d", q.Len())
	}

	// Dequeue should yield strictly in priority order: Critical, High, Normal, Low
	item1, lane1, _ := q.Dequeue()
	if item1 != "run-crit" || lane1 != LaneCritical {
		t.Errorf("expected run-crit first, got %s", item1)
	}

	item2, lane2, _ := q.Dequeue()
	if item2 != "run-high" || lane2 != LaneHigh {
		t.Errorf("expected run-high second, got %s", item2)
	}

	item3, lane3, _ := q.Dequeue()
	if item3 != "run-normal" || lane3 != LaneNormal {
		t.Errorf("expected run-normal third, got %s", item3)
	}

	item4, lane4, _ := q.Dequeue()
	if item4 != "run-low" || lane4 != LaneLow {
		t.Errorf("expected run-low fourth, got %s", item4)
	}

	_, _, err := q.Dequeue()
	if err == nil {
		t.Error("expected error dequeuing from empty queue, got nil")
	}
}
