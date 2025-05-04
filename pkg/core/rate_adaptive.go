package core

import (
	"sync"
	"time"
)

type AdaptiveRateController struct {
	mu           sync.RWMutex
	currentRate  float64
	minRate      float64
	maxRate      float64
	additiveStep float64
	backoffMul   float64
	lastUpdate   time.Time
}

func NewAdaptiveRateController(initialRate, minRate, maxRate float64) *AdaptiveRateController {
	if minRate <= 0 {
		minRate = 1.0
	}
	if maxRate < minRate {
		maxRate = 1000.0
	}
	if initialRate < minRate {
		initialRate = minRate
	}
	return &AdaptiveRateController{
		currentRate:  initialRate,
		minRate:      minRate,
		maxRate:      maxRate,
		additiveStep: 1.0,
		backoffMul:   0.8,
		lastUpdate:   time.Now().UTC(),
	}
}

// OnSuccess applies additive increase (AIMD).
func (c *AdaptiveRateController) OnSuccess() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.currentRate += c.additiveStep
	if c.currentRate > c.maxRate {
		c.currentRate = c.maxRate
	}
	c.lastUpdate = time.Now().UTC()
}

// OnThrottle applies multiplicative decrease (AIMD).
func (c *AdaptiveRateController) OnThrottle() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.currentRate *= c.backoffMul
	if c.currentRate < c.minRate {
		c.currentRate = c.minRate
	}
	c.lastUpdate = time.Now().UTC()
}

func (c *AdaptiveRateController) CurrentRate() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.currentRate
}
