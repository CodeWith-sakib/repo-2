package scheduler

import (
	"sync"
)

type DRRQueue struct {
	TenantID string
	Deficit  int
	Quantum  int
	Items    []string
}

type DeficitRoundRobinScheduler struct {
	mu      sync.Mutex
	queues  map[string]*DRRQueue
	order   []string
	currIdx int
	quantum int
}

func NewDeficitRoundRobinScheduler(defaultQuantum int) *DeficitRoundRobinScheduler {
	if defaultQuantum <= 0 {
		defaultQuantum = 100
	}
	return &DeficitRoundRobinScheduler{
		queues:  make(map[string]*DRRQueue),
		order:   make([]string, 0),
		quantum: defaultQuantum,
	}
}

func (s *DeficitRoundRobinScheduler) Enqueue(tenant string, item string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	q, exists := s.queues[tenant]
	if !exists {
		q = &DRRQueue{
			TenantID: tenant,
			Quantum:  s.quantum,
			Items:    make([]string, 0),
		}
		s.queues[tenant] = q
		s.order = append(s.order, tenant)
	}
	q.Items = append(q.Items, item)
}

func (s *DeficitRoundRobinScheduler) Dequeue(itemCost int) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.order) == 0 {
		return "", false
	}

	for rounds := 0; rounds < len(s.order); rounds++ {
		tenant := s.order[s.currIdx]
		q := s.queues[tenant]

		if len(q.Items) > 0 {
			q.Deficit += q.Quantum
			if q.Deficit >= itemCost {
				item := q.Items[0]
				q.Items = q.Items[1:]
				q.Deficit -= itemCost
				s.currIdx = (s.currIdx + 1) % len(s.order)
				return item, true
			}
		} else {
			q.Deficit = 0
		}

		s.currIdx = (s.currIdx + 1) % len(s.order)
	}

	return "", false
}
