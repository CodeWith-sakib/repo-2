package postgres

import (
	"math"
	"math/rand"
	"time"
)

// FullJitterBackoff computes exponential backoff with full jitter to avoid thundering herds.
type FullJitterBackoff struct {
	baseInterval time.Duration
	maxInterval  time.Duration
	multiplier   float64
}

// NewFullJitterBackoff creates an exponential backoff jitter calculator.
func NewFullJitterBackoff(base, max time.Duration) *FullJitterBackoff {
	if base <= 0 {
		base = 20 * time.Millisecond
	}
	if max <= 0 {
		max = 5 * time.Second
	}
	return &FullJitterBackoff{
		baseInterval: base,
		maxInterval:  max,
		multiplier:   2.0,
	}
}

// ComputeBackoff returns a jittered sleep duration between 0 and min(maxInterval, base * multiplier^attempt).
func (b *FullJitterBackoff) ComputeBackoff(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}

	temp := float64(b.baseInterval) * math.Pow(b.multiplier, float64(attempt))
	capVal := math.Min(float64(b.maxInterval), temp)

	if capVal <= 0 {
		return b.baseInterval
	}

	// Full jitter: uniformly random duration between 0 and cap
	jitter := rand.Float64() * capVal
	return time.Duration(jitter)
}
