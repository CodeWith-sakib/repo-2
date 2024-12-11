package scheduler

import (
	"container/heap"
	"sync"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/storage"
)

type QueueItem struct {
	Task     *storage.QueuedTask
	Weight   int
	Index    int
	QueuedAt time.Time
}

type itemHeap []*QueueItem

func (h itemHeap) Len() int { return len(h) }
func (h itemHeap) Less(i, j int) bool {
	// Highest priority first, tie-break by oldest scheduled
	if h[i].Task.Priority != h[j].Task.Priority {
		return h[i].Task.Priority > h[j].Task.Priority
	}
	return h[i].Task.ScheduledAt.Before(h[j].Task.ScheduledAt)
}
func (h itemHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].Index = i
	h[j].Index = j
}
func (h *itemHeap) Push(x interface{}) {
	n := len(*h)
	item := x.(*QueueItem)
	item.Index = n
	*h = append(*h, item)
}
func (h *itemHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.Index = -1
	*h = old[0 : n-1]
	return item
}

type FairPriorityQueue struct {
	mu           sync.Mutex
	pq           itemHeap
	tenantCounts map[string]int
}

func NewFairPriorityQueue() *FairPriorityQueue {
	fq := &FairPriorityQueue{
		pq:           make(itemHeap, 0),
		tenantCounts: make(map[string]int),
	}
	heap.Init(&fq.pq)
	return fq
}

func (fq *FairPriorityQueue) Push(task *storage.QueuedTask) {
	fq.mu.Lock()
	defer fq.mu.Unlock()

	item := &QueueItem{
		Task:     task,
		QueuedAt: time.Now().UTC(),
	}
	heap.Push(&fq.pq, item)
	fq.tenantCounts[task.TenantID]++
}

func (fq *FairPriorityQueue) Pop() *storage.QueuedTask {
	fq.mu.Lock()
	defer fq.mu.Unlock()

	if len(fq.pq) == 0 {
		return nil
	}

	item := heap.Pop(&fq.pq).(*QueueItem)
	fq.tenantCounts[item.Task.TenantID]--
	if fq.tenantCounts[item.Task.TenantID] <= 0 {
		delete(fq.tenantCounts, item.Task.TenantID)
	}
	return item.Task
}

func (fq *FairPriorityQueue) Len() int {
	fq.mu.Lock()
	defer fq.mu.Unlock()
	return len(fq.pq)
}

func (fq *FairPriorityQueue) TenantLoad(tenant string) int {
	fq.mu.Lock()
	defer fq.mu.Unlock()
	return fq.tenantCounts[tenant]
}
