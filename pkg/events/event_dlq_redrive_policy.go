package events

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrMaxRedriveExceeded = errors.New("event redrive count exceeded max allowed attempts")
)

// RedrivePolicy defines rules for replaying dead-lettered events back onto target topics.
type RedrivePolicy struct {
	MaxRedrives   int           `json:"max_redrives"`
	BackoffDelay  time.Duration `json:"backoff_delay"`
	FallbackTopic string        `json:"fallback_topic"`
}

// DLQRedriveManager orchestrates safe event re-ingestion.
type DLQRedriveManager struct {
	mu           sync.RWMutex
	policy       RedrivePolicy
	redriveCounts map[string]int // eventID -> count
}

// NewDLQRedriveManager initializes a redrive controller.
func NewDLQRedriveManager(policy RedrivePolicy) *DLQRedriveManager {
	if policy.MaxRedrives <= 0 {
		policy.MaxRedrives = 3
	}
	return &DLQRedriveManager{
		policy:        policy,
		redriveCounts: make(map[string]int),
	}
}

// CanRedrive checks whether an event is eligible for replay.
func (m *DLQRedriveManager) CanRedrive(eventID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.redriveCounts[eventID] < m.policy.MaxRedrives
}

// RecordRedriveAttempt increments attempt counter or returns error if exhausted.
func (m *DLQRedriveManager) RecordRedriveAttempt(ctx context.Context, eventID string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	current := m.redriveCounts[eventID]
	if current >= m.policy.MaxRedrives {
		return current, ErrMaxRedriveExceeded
	}

	m.redriveCounts[eventID] = current + 1
	return m.redriveCounts[eventID], nil
}

// TargetTopic returns original topic or fallback dead-letter sink topic.
func (m *DLQRedriveManager) TargetTopic(eventID string, originalTopic string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.redriveCounts[eventID] >= m.policy.MaxRedrives {
		return m.policy.FallbackTopic
	}
	return originalTopic
}
