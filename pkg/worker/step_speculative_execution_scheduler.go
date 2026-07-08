package worker

import (
	"sync"
	"time"
)

// SpeculativeExecutionSpec models threshold for launching backup clone attempts.
type SpeculativeExecutionSpec struct {
	P95Duration     time.Duration `json:"p95_duration"`
	MinRunDuration  time.Duration `json:"min_run_duration"`
	SlowFactorMulti float64       `json:"slow_factor_multi"`
}

// SpeculativeExecutionScheduler detects stragglers and suggests redundant speculative clones.
type SpeculativeExecutionScheduler struct {
	mu   sync.RWMutex
	spec SpeculativeExecutionSpec
}

// NewSpeculativeExecutionScheduler creates a straggler mitigation scheduler.
func NewSpeculativeExecutionScheduler(spec SpeculativeExecutionSpec) *SpeculativeExecutionScheduler {
	if spec.SlowFactorMulti <= 1.0 {
		spec.SlowFactorMulti = 1.5
	}
	if spec.MinRunDuration <= 0 {
		spec.MinRunDuration = 5 * time.Second
	}
	return &SpeculativeExecutionScheduler{
		spec: spec,
	}
}

// ShouldSpeculate evaluates whether an ongoing task has run long enough to warrant a clone.
func (s *SpeculativeExecutionScheduler) ShouldSpeculate(startedAt time.Time, now time.Time) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	elapsed := now.Sub(startedAt)
	if elapsed < s.spec.MinRunDuration {
		return false
	}

	threshold := time.Duration(float64(s.spec.P95Duration) * s.spec.SlowFactorMulti)
	return elapsed >= threshold
}
