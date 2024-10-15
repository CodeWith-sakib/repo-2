package retry

import (
	"math"
	"math/rand"
	"sync"
	"time"
)

type BackoffStrategy interface {
	NextInterval(attempt int) time.Duration
}

type ExponentialBackoff struct {
	InitialInterval time.Duration
	MaxInterval     time.Duration
	Multiplier      float64
	Jitter          bool
	mu              sync.Mutex
	rng             *rand.Rand
}

func NewExponentialBackoff(initial, max time.Duration, multiplier float64, jitter bool) *ExponentialBackoff {
	if initial <= 0 {
		initial = 100 * time.Millisecond
	}
	if max <= 0 || max < initial {
		max = 30 * time.Second
	}
	if multiplier <= 1.0 {
		multiplier = 2.0
	}

	return &ExponentialBackoff{
		InitialInterval: initial,
		MaxInterval:     max,
		Multiplier:      multiplier,
		Jitter:          jitter,
		rng:             rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (eb *ExponentialBackoff) NextInterval(attempt int) time.Duration {
	if attempt <= 0 {
		attempt = 1
	}

	// baseInterval = initial * multiplier^(attempt-1)
	factor := math.Pow(eb.Multiplier, float64(attempt-1))
	baseInterval := float64(eb.InitialInterval) * factor

	if baseInterval > float64(eb.MaxInterval) {
		baseInterval = float64(eb.MaxInterval)
	}

	interval := time.Duration(baseInterval)
	if !eb.Jitter {
		return interval
	}

	eb.mu.Lock()
	defer eb.mu.Unlock()

	// Full jitter: random duration between 0 and baseInterval
	jittered := eb.rng.Float64() * float64(interval)
	return time.Duration(jittered)
}

type LinearBackoff struct {
	Step        time.Duration
	MaxInterval time.Duration
}

func NewLinearBackoff(step, max time.Duration) *LinearBackoff {
	return &LinearBackoff{
		Step:        step,
		MaxInterval: max,
	}
}

func (lb *LinearBackoff) NextInterval(attempt int) time.Duration {
	interval := time.Duration(attempt) * lb.Step
	if interval > lb.MaxInterval {
		return lb.MaxInterval
	}
	return interval
}
