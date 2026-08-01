package postgres

import (
	"context"
	"sync"
	"time"
)

// LockContentionEvent records a lock acquisition wait incident.
type LockContentionEvent struct {
	RelationName string        `json:"relation_name"`
	LockMode     string        `json:"lock_mode"`
	WaitDuration time.Duration `json:"wait_duration"`
	BlockedTxID  int64         `json:"blocked_txid"`
	ObservedAt   time.Time     `json:"observed_at"`
}

// LockContentionSampler aggregates lock acquisition latency statistics.
type LockContentionSampler struct {
	mu          sync.RWMutex
	events      []LockContentionEvent
	warnLatency time.Duration
}

// NewLockContentionSampler initializes a lock contention observer.
func NewLockContentionSampler(warnLatencyThreshold time.Duration) *LockContentionSampler {
	if warnLatencyThreshold <= 0 {
		warnLatencyThreshold = 50 * time.Millisecond
	}
	return &LockContentionSampler{
		events:      make([]LockContentionEvent, 0),
		warnLatency: warnLatencyThreshold,
	}
}

// RecordLockWait stores an observed lock acquisition event.
func (s *LockContentionSampler) RecordLockWait(ctx context.Context, relation, mode string, dur time.Duration, txID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.events = append(s.events, LockContentionEvent{
		RelationName: relation,
		LockMode:     mode,
		WaitDuration: dur,
		BlockedTxID:  txID,
		ObservedAt:   time.Now(),
	})
}

// CriticalContentionCount returns incidents exceeding warning latency.
func (s *LockContentionSampler) CriticalContentionCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	for _, e := range s.events {
		if e.WaitDuration >= s.warnLatency {
			count++
		}
	}
	return count
}

// TotalContentionTime sums total time spent waiting on locks.
func (s *LockContentionSampler) TotalContentionTime() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var total time.Duration
	for _, e := range s.events {
		total += e.WaitDuration
	}
	return total
}
