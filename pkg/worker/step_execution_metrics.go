package worker

import (
	"sync/atomic"
	"time"
)

// StepExecutionTelemetry aggregates cumulative task execution durations and success tallies.
type StepExecutionTelemetry struct {
	TotalRuns     atomic.Int64
	SuccessRuns   atomic.Int64
	FailureRuns   atomic.Int64
	TotalDuration atomic.Int64 // nanoseconds
	MaxDuration   atomic.Int64 // nanoseconds
}

// NewStepExecutionTelemetry creates a new metrics recorder.
func NewStepExecutionTelemetry() *StepExecutionTelemetry {
	return &StepExecutionTelemetry{}
}

// RecordExecution stores an execution outcome with duration.
func (m *StepExecutionTelemetry) RecordExecution(duration time.Duration, success bool) {
	m.TotalRuns.Add(1)
	if success {
		m.SuccessRuns.Add(1)
	} else {
		m.FailureRuns.Add(1)
	}

	nanos := duration.Nanoseconds()
	m.TotalDuration.Add(nanos)

	// Update MaxDuration CAS loop
	for {
		currMax := m.MaxDuration.Load()
		if nanos <= currMax {
			break
		}
		if m.MaxDuration.CompareAndSwap(currMax, nanos) {
			break
		}
	}
}

// MeanDuration calculates average task latency.
func (m *StepExecutionTelemetry) MeanDuration() time.Duration {
	total := m.TotalRuns.Load()
	if total == 0 {
		return 0
	}
	return time.Duration(m.TotalDuration.Load() / total)
}

// SuccessRatePct calculates completion success percentage.
func (m *StepExecutionTelemetry) SuccessRatePct() float64 {
	total := m.TotalRuns.Load()
	if total == 0 {
		return 100.0
	}
	return (float64(m.SuccessRuns.Load()) / float64(total)) * 100.0
}
