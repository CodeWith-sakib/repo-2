package worker

import (
	"sync"
	"time"
)

// VegasConfig sets thresholds for the TCP Vegas adaptive concurrency algorithm.
type VegasConfig struct {
	MinLimit  int
	MaxLimit  int
	Alpha     float64 // queue threshold to increase limit (e.g. 3.0)
	Beta      float64 // queue threshold to decrease limit (e.g. 6.0)
	Smoothing float64 // EMA smoothing factor for observed latency (e.g. 0.2)
}

// DefaultVegasConfig returns standard Vegas parameters.
func DefaultVegasConfig() VegasConfig {
	return VegasConfig{
		MinLimit:  2,
		MaxLimit:  100,
		Alpha:     3.0,
		Beta:      6.0,
		Smoothing: 0.2,
	}
}

// VegasConcurrencyController dynamically tunes in-flight task limits based on latency gradients.
type VegasConcurrencyController struct {
	mu           sync.RWMutex
	cfg          VegasConfig
	currentLimit float64
	minRTT       time.Duration // lowest observed round-trip latency
	smoothedRTT  time.Duration // smoothed current round-trip latency
	inFlight     int
}

// NewVegasConcurrencyController creates a new adaptive controller.
func NewVegasConcurrencyController(cfg VegasConfig) *VegasConcurrencyController {
	if cfg.MinLimit <= 0 {
		cfg.MinLimit = 1
	}
	if cfg.MaxLimit < cfg.MinLimit {
		cfg.MaxLimit = cfg.MinLimit * 10
	}

	initialLimit := float64(cfg.MinLimit * 2)
	if initialLimit > float64(cfg.MaxLimit) {
		initialLimit = float64(cfg.MaxLimit)
	}
	if initialLimit < float64(cfg.MinLimit) {
		initialLimit = float64(cfg.MinLimit)
	}

	return &VegasConcurrencyController{
		cfg:          cfg,
		currentLimit: initialLimit,
	}
}

// TryAcquire attempts to acquire an execution slot according to current limit.
func (c *VegasConcurrencyController) TryAcquire() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if float64(c.inFlight) < c.currentLimit {
		c.inFlight++
		return true
	}
	return false
}

// Release records the observed execution duration and adjusts the concurrency limit.
func (c *VegasConcurrencyController) Release(latency time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.inFlight > 0 {
		c.inFlight--
	}

	if latency <= 0 {
		return
	}

	// Update min RTT baseline
	if c.minRTT == 0 || latency < c.minRTT {
		c.minRTT = latency
	}

	// Update smoothed RTT
	if c.smoothedRTT == 0 {
		c.smoothedRTT = latency
	} else {
		c.smoothedRTT = time.Duration((c.cfg.Smoothing * float64(latency)) + ((1.0 - c.cfg.Smoothing) * float64(c.smoothedRTT)))
	}

	// Vegas queue calculation: queue = limit * (1 - minRTT / smoothedRTT)
	if c.smoothedRTT > 0 && c.minRTT > 0 {
		queue := c.currentLimit * (1.0 - (float64(c.minRTT) / float64(c.smoothedRTT)))

		if queue > c.cfg.Beta {
			// Congestion! Reduce limit
			c.currentLimit -= 1.0
		} else if queue < c.cfg.Alpha {
			// Spare capacity: increase limit
			c.currentLimit += 1.0
		}

		// Clamp to boundaries
		if c.currentLimit < float64(c.cfg.MinLimit) {
			c.currentLimit = float64(c.cfg.MinLimit)
		}
		if c.currentLimit > float64(c.cfg.MaxLimit) {
			c.currentLimit = float64(c.cfg.MaxLimit)
		}
	}
}

// Limit returns the current integer concurrency ceiling.
func (c *VegasConcurrencyController) Limit() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return int(c.currentLimit)
}

// InFlight returns currently acquired slots.
func (c *VegasConcurrencyController) InFlight() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.inFlight
}
