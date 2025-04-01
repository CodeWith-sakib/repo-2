package worker

import (
	"context"
	"sync"
)

type WorkerStealStats struct {
	TasksExecuted int64
	TasksStolen   int64
}

type PriorityWorker struct {
	id         string
	highQueue  chan string
	lowQueue   chan string
	peerQueues []chan string
	mu         sync.Mutex
	stats      WorkerStealStats
}

func NewPriorityWorker(id string, queueSize int) *PriorityWorker {
	if queueSize <= 0 {
		queueSize = 100
	}
	return &PriorityWorker{
		id:         id,
		highQueue:  make(chan string, queueSize),
		lowQueue:   make(chan string, queueSize),
		peerQueues: make([]chan string, 0),
	}
}

func (pw *PriorityWorker) AddPeerQueue(ch chan string) {
	pw.mu.Lock()
	defer pw.mu.Unlock()
	pw.peerQueues = append(pw.peerQueues, ch)
}

func (pw *PriorityWorker) Submit(taskID string, highPriority bool) bool {
	if highPriority {
		select {
		case pw.highQueue <- taskID:
			return true
		default:
			return false
		}
	}
	select {
	case pw.lowQueue <- taskID:
		return true
	default:
		return false
	}
}

func (pw *PriorityWorker) Poll(ctx context.Context) (string, bool) {
	// 1. High priority queue
	select {
	case task := <-pw.highQueue:
		pw.mu.Lock()
		pw.stats.TasksExecuted++
		pw.mu.Unlock()
		return task, true
	default:
	}

	// 2. Low priority queue
	select {
	case task := <-pw.lowQueue:
		pw.mu.Lock()
		pw.stats.TasksExecuted++
		pw.mu.Unlock()
		return task, true
	default:
	}

	// 3. Work-stealing from peers
	pw.mu.Lock()
	peers := append([]chan string(nil), pw.peerQueues...)
	pw.mu.Unlock()

	for _, peer := range peers {
		select {
		case stolen := <-peer:
			pw.mu.Lock()
			pw.stats.TasksExecuted++
			pw.stats.TasksStolen++
			pw.mu.Unlock()
			return stolen, true
		default:
		}
	}

	return "", false
}

func (pw *PriorityWorker) GetStats() WorkerStealStats {
	pw.mu.Lock()
	defer pw.mu.Unlock()
	return pw.stats
}
