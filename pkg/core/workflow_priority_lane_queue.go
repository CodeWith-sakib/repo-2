package core

import (
	"errors"
	"sync"
)

// PriorityLane defines priority tier classification.
type PriorityLane int

const (
	LaneCritical PriorityLane = iota
	LaneHigh
	LaneNormal
	LaneLow
)

// QueuedWorkflowItem holds a workflow run identifier and assigned priority lane.
type QueuedWorkflowItem struct {
	RunID    string
	Lane     PriorityLane
	Weight   int
}

// PriorityLaneQueue implements strict weighted multi-lane priority scheduling.
type PriorityLaneQueue struct {
	mu     sync.Mutex
	lanes  map[PriorityLane][]string
	counts map[PriorityLane]int
}

// NewPriorityLaneQueue creates a multi-tier priority queue.
func NewPriorityLaneQueue() *PriorityLaneQueue {
	return &PriorityLaneQueue{
		lanes: map[PriorityLane][]string{
			LaneCritical: nil,
			LaneHigh:     nil,
			LaneNormal:   nil,
			LaneLow:      nil,
		},
		counts: make(map[PriorityLane]int),
	}
}

// Enqueue inserts a workflow run into designated priority lane.
func (q *PriorityLaneQueue) Enqueue(runID string, lane PriorityLane) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.lanes[lane] = append(q.lanes[lane], runID)
	q.counts[lane]++
}

// Dequeue pops the next highest-priority workflow run.
func (q *PriorityLaneQueue) Dequeue() (string, PriorityLane, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	order := []PriorityLane{LaneCritical, LaneHigh, LaneNormal, LaneLow}

	for _, lane := range order {
		items := q.lanes[lane]
		if len(items) > 0 {
			item := items[0]
			q.lanes[lane] = items[1:]
			q.counts[lane]--
			return item, lane, nil
		}
	}

	return "", 0, errors.New("priority queue is empty")
}

// Len returns total pending items across all lanes.
func (q *PriorityLaneQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	total := 0
	for _, count := range q.counts {
		total += count
	}
	return total
}
