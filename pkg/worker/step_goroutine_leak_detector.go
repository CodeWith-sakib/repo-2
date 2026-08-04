package worker

import (
	"runtime"
	"sync"
	"time"
)

// GoroutineLeakReport captures runtime goroutine count differentials before and after task execution.
type GoroutineLeakReport struct {
	StepID            string    `json:"step_id"`
	InitialCount      int       `json:"initial_count"`
	FinalCount        int       `json:"final_count"`
	SuspectedLeakDiff int       `json:"suspected_leak_diff"`
	MeasuredAt        time.Time `json:"measured_at"`
}

// GoroutineLeakDetector monitors worker threads to identify unclosed channels and dangling routines.
type GoroutineLeakDetector struct {
	mu           sync.RWMutex
	warnDiff     int
	initialCounts map[string]int
}

// NewGoroutineLeakDetector initializes a leak detector.
func NewGoroutineLeakDetector(warnDiffThreshold int) *GoroutineLeakDetector {
	if warnDiffThreshold <= 0 {
		warnDiffThreshold = 5
	}
	return &GoroutineLeakDetector{
		warnDiff:      warnDiffThreshold,
		initialCounts: make(map[string]int),
	}
}

// BeginTrack captures baseline goroutine count before starting step.
func (d *GoroutineLeakDetector) BeginTrack(stepID string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.initialCounts[stepID] = runtime.NumGoroutine()
}

// EndTrack compares current goroutine count against baseline.
func (d *GoroutineLeakDetector) EndTrack(stepID string) GoroutineLeakReport {
	d.mu.Lock()
	defer d.mu.Unlock()

	initCount := d.initialCounts[stepID]
	delete(d.initialCounts, stepID)

	finalCount := runtime.NumGoroutine()
	diff := finalCount - initCount
	if diff < 0 {
		diff = 0
	}

	return GoroutineLeakReport{
		StepID:            stepID,
		InitialCount:      initCount,
		FinalCount:        finalCount,
		SuspectedLeakDiff: diff,
		MeasuredAt:        time.Now(),
	}
}
