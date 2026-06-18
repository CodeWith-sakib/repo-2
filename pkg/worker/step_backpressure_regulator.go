package worker

import (
	"sync"
	"time"
)

// BackpressureState represents the current throttling state of the worker.
type BackpressureState string

const (
	BackpressureNormal    BackpressureState = "NORMAL"
	BackpressureThrottled BackpressureState = "THROTTLED"
	BackpressureSaturated BackpressureState = "SATURATED"
)

// StepBackpressureRegulator monitors queue depth and task error rates to adaptively regulate intake.
type StepBackpressureRegulator struct {
	mu               sync.RWMutex
	highWatermark    int
	lowWatermark     int
	errorRatioLimit  float64
	currentQueueSize int
	recentErrors     int
	recentSuccesses  int
	lastEvaluated    time.Time
}

// NewStepBackpressureRegulator creates a backpressure regulator.
func NewStepBackpressureRegulator(lowWatermark, highWatermark int, errorRatioLimit float64) *StepBackpressureRegulator {
	if lowWatermark <= 0 {
		lowWatermark = 50
	}
	if highWatermark <= lowWatermark {
		highWatermark = lowWatermark * 2
	}
	if errorRatioLimit <= 0.0 || errorRatioLimit >= 1.0 {
		errorRatioLimit = 0.25
	}
	return &StepBackpressureRegulator{
		lowWatermark:    lowWatermark,
		highWatermark:   highWatermark,
		errorRatioLimit: errorRatioLimit,
		lastEvaluated:   time.Now(),
	}
}

// UpdateQueueDepth records the latest observation of pending step queue size.
func (r *StepBackpressureRegulator) UpdateQueueDepth(depth int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.currentQueueSize = depth
}

// RecordExecutionResult records execution outcome for error rate tracking.
func (r *StepBackpressureRegulator) RecordExecutionResult(err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err != nil {
		r.recentErrors++
	} else {
		r.recentSuccesses++
	}
}

// CurrentState evaluates the backpressure status.
func (r *StepBackpressureRegulator) CurrentState() BackpressureState {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.currentQueueSize >= r.highWatermark {
		return BackpressureSaturated
	}

	total := r.recentErrors + r.recentSuccesses
	if total >= 10 {
		errRatio := float64(r.recentErrors) / float64(total)
		if errRatio >= r.errorRatioLimit {
			return BackpressureThrottled
		}
	}

	if r.currentQueueSize > r.lowWatermark {
		return BackpressureThrottled
	}

	return BackpressureNormal
}

// ShouldAccept returns true if the worker is able to ingest new steps.
func (r *StepBackpressureRegulator) ShouldAccept() bool {
	return r.CurrentState() != BackpressureSaturated
}
