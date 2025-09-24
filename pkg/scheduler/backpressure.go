package scheduler

import (
	"sync"
	"time"
)

// PressureLevel categorizes the system operational stress state.
type PressureLevel string

const (
	PressureNormal   PressureLevel = "normal"
	PressureModerate PressureLevel = "moderate"
	PressureSevere   PressureLevel = "severe"
	PressureCritical PressureLevel = "critical"
)

// BackpressureThresholds defines the trigger levels for pressure states.
type BackpressureThresholds struct {
	ModerateWatermark float64 // e.g. 0.65
	SevereWatermark   float64 // e.g. 0.85
	CriticalWatermark float64 // e.g. 0.95
	MinPriorityNormal int     // minimum priority accepted at normal
	MinPriorityMod    int     // minimum priority accepted at moderate
	MinPrioritySevere int     // minimum priority accepted at severe
}

// DefaultBackpressureThresholds returns standard defaults.
func DefaultBackpressureThresholds() BackpressureThresholds {
	return BackpressureThresholds{
		ModerateWatermark: 0.65,
		SevereWatermark:   0.85,
		CriticalWatermark: 0.95,
		MinPriorityNormal: 0,
		MinPriorityMod:    20,
		MinPrioritySevere: 50,
	}
}

// SystemPressureSignals encapsulates raw telemetry inputs.
type SystemPressureSignals struct {
	QueueSaturation  float64 // [0.0, 1.0]
	WorkerInFlight   float64 // [0.0, 1.0]
	StorageLatencyMs float64
	MemoryUsageRatio float64 // [0.0, 1.0]
}

// BackpressureController computes smoothed composite pressure and gates task submission.
type BackpressureController struct {
	mu            sync.RWMutex
	thresholds    BackpressureThresholds
	alpha         float64 // EMA smoothing factor, e.g. 0.3
	smoothedScore float64
	currentLevel  PressureLevel
	lastUpdated   time.Time
}

// NewBackpressureController initializes a controller with smoothing and thresholds.
func NewBackpressureController(thresholds BackpressureThresholds, emaAlpha float64) *BackpressureController {
	if emaAlpha <= 0 || emaAlpha > 1.0 {
		emaAlpha = 0.3
	}
	return &BackpressureController{
		thresholds:   thresholds,
		alpha:        emaAlpha,
		currentLevel: PressureNormal,
		lastUpdated:  time.Now(),
	}
}

// UpdateSignals evaluates raw signals and updates the smoothed pressure score and level.
func (c *BackpressureController) UpdateSignals(signals SystemPressureSignals) (score float64, level PressureLevel) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Normalize storage latency (100ms = 1.0)
	normLatency := signals.StorageLatencyMs / 100.0
	if normLatency > 1.0 {
		normLatency = 1.0
	}

	// Weighted composite score
	rawScore := (signals.QueueSaturation * 0.35) +
		(signals.WorkerInFlight * 0.30) +
		(signals.MemoryUsageRatio * 0.20) +
		(normLatency * 0.15)

	if rawScore > 1.0 {
		rawScore = 1.0
	}

	// Exponential moving average smoothing
	c.smoothedScore = (c.alpha * rawScore) + ((1.0 - c.alpha) * c.smoothedScore)
	c.lastUpdated = time.Now()

	switch {
	case c.smoothedScore >= c.thresholds.CriticalWatermark:
		c.currentLevel = PressureCritical
	case c.smoothedScore >= c.thresholds.SevereWatermark:
		c.currentLevel = PressureSevere
	case c.smoothedScore >= c.thresholds.ModerateWatermark:
		c.currentLevel = PressureModerate
	default:
		c.currentLevel = PressureNormal
	}

	return c.smoothedScore, c.currentLevel
}

// ShouldAdmit decides if a task with the given priority should be accepted under current pressure.
func (c *BackpressureController) ShouldAdmit(priority int) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	switch c.currentLevel {
	case PressureNormal:
		return priority >= c.thresholds.MinPriorityNormal
	case PressureModerate:
		return priority >= c.thresholds.MinPriorityMod
	case PressureSevere:
		return priority >= c.thresholds.MinPrioritySevere
	case PressureCritical:
		// Critical sheds everything except highest emergency priority (>= 90)
		return priority >= 90
	default:
		return true
	}
}

// State returns current pressure score and level.
func (c *BackpressureController) State() (float64, PressureLevel) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.smoothedScore, c.currentLevel
}
