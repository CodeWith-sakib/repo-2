package metrics

import (
	"fmt"
	"sync"
	"time"
)

// SLIEvent indicates whether an operation met its service level objective.
type SLIEvent struct {
	Timestamp time.Time
	Success   bool
}

// SLATrackerConfig sets availability targets and calculation windows.
type SLATrackerConfig struct {
	TargetAvailability float64       // e.g. 0.999 for 99.9%
	WindowDuration     time.Duration // e.g. 30 days
	ShortWindow        time.Duration // e.g. 1 hour
}

// DefaultSLATrackerConfig returns standard configuration.
func DefaultSLATrackerConfig() SLATrackerConfig {
	return SLATrackerConfig{
		TargetAvailability: 0.999,
		WindowDuration:     30 * 24 * time.Hour,
		ShortWindow:        1 * time.Hour,
	}
}

// SLABudgetStatus reports the current state of an error budget.
type SLABudgetStatus struct {
	TotalRequests      int64
	FailedRequests     int64
	Availability       float64
	AllowedFailures    float64
	BudgetRemainingPct float64
	BurnRateShort      float64
	IsExhausted        bool
}

// SLATracker evaluates service level indicators against objectives.
type SLATracker struct {
	mu     sync.RWMutex
	cfg    SLATrackerConfig
	events []SLIEvent
}

// NewSLATracker creates an SLA tracker instance.
func NewSLATracker(cfg SLATrackerConfig) *SLATracker {
	if cfg.TargetAvailability <= 0 || cfg.TargetAvailability >= 1.0 {
		cfg.TargetAvailability = 0.999
	}
	if cfg.WindowDuration <= 0 {
		cfg.WindowDuration = 24 * time.Hour
	}
	if cfg.ShortWindow <= 0 {
		cfg.ShortWindow = time.Hour
	}
	return &SLATracker{
		cfg: cfg,
	}
}

// Record observes an operation result.
func (t *SLATracker) Record(success bool) {
	t.RecordAt(time.Now(), success)
}

// RecordAt records an observation with an explicit timestamp.
func (t *SLATracker) RecordAt(at time.Time, success bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.events = append(t.events, SLIEvent{
		Timestamp: at,
		Success:   success,
	})

	// Prune events older than WindowDuration
	cutoff := at.Add(-t.cfg.WindowDuration)
	i := 0
	for ; i < len(t.events); i++ {
		if t.events[i].Timestamp.After(cutoff) {
			break
		}
	}
	if i > 0 {
		t.events = t.events[i:]
	}
}

// Status calculates current availability, remaining error budget, and burn rate.
func (t *SLATracker) Status() SLABudgetStatus {
	t.mu.RLock()
	defer t.mu.RUnlock()

	now := time.Now()
	cutoffWindow := now.Add(-t.cfg.WindowDuration)
	cutoffShort := now.Add(-t.cfg.ShortWindow)

	var total, failed int64
	var totalShort, failedShort int64

	for _, e := range t.events {
		if e.Timestamp.After(cutoffWindow) {
			total++
			if !e.Success {
				failed++
			}
		}
		if e.Timestamp.After(cutoffShort) {
			totalShort++
			if !e.Success {
				failedShort++
			}
		}
	}

	if total == 0 {
		return SLABudgetStatus{
			Availability:       1.0,
			BudgetRemainingPct: 100.0,
		}
	}

	avail := float64(total-failed) / float64(total)
	allowedFailures := float64(total) * (1.0 - t.cfg.TargetAvailability)

	budgetRemaining := 100.0
	if allowedFailures > 0 {
		budgetRemaining = ((allowedFailures - float64(failed)) / allowedFailures) * 100.0
	}
	if budgetRemaining < 0 {
		budgetRemaining = 0
	}

	// Burn rate in short window: (observed error rate) / (allowed error rate)
	burnRate := 0.0
	allowedErrorRate := 1.0 - t.cfg.TargetAvailability
	if totalShort > 0 && allowedErrorRate > 0 {
		observedRate := float64(failedShort) / float64(totalShort)
		burnRate = observedRate / allowedErrorRate
	}

	return SLABudgetStatus{
		TotalRequests:      total,
		FailedRequests:     failed,
		Availability:       avail,
		AllowedFailures:    allowedFailures,
		BudgetRemainingPct: budgetRemaining,
		BurnRateShort:      burnRate,
		IsExhausted:        budgetRemaining <= 0,
	}
}

// String provides a human-readable diagnosis of the SLA status.
func (s SLABudgetStatus) String() string {
	return fmt.Sprintf("reqs=%d fail=%d avail=%.4f%% budget=%.1f%% burn=%.2fx",
		s.TotalRequests, s.FailedRequests, s.Availability*100, s.BudgetRemainingPct, s.BurnRateShort)
}
