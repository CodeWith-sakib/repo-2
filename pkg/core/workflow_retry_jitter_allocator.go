package core

import (
	"math"
	"math/rand"
	"sync"
	"time"
)

// JitterStrategy specifies how backoff randomization is applied.
type JitterStrategy string

const (
	FullJitter  JitterStrategy = "FULL_JITTER"
	EqualJitter JitterStrategy = "EQUAL_JITTER"
	Decorrelated JitterStrategy = "DECORRELATED"
)

// RetryJitterAllocator computes randomized backoff delays to avoid retry storms across clusters.
type RetryJitterAllocator struct {
	mu       sync.Mutex
	rng      *rand.Rand
	base     time.Duration
	max      time.Duration
	strategy JitterStrategy
}

// NewRetryJitterAllocator constructs a jitter calculator.
func NewRetryJitterAllocator(base, max time.Duration, strategy JitterStrategy) *RetryJitterAllocator {
	if base <= 0 {
		base = 100 * time.Millisecond
	}
	if max <= base {
		max = 30 * time.Second
	}
	return &RetryJitterAllocator{
		rng:      rand.New(rand.NewSource(time.Now().UnixNano())),
		base:     base,
		max:      max,
		strategy: strategy,
	}
}

// ComputeDelay calculates the backoff interval for the n-th retry attempt (0-indexed).
func (a *RetryJitterAllocator) ComputeDelay(attempt int) time.Duration {
	a.mu.Lock()
	defer a.mu.Unlock()

	// exponential cap
	exp := math.Min(float64(a.max), float64(a.base)*math.Pow(2, float64(attempt)))
	capDuration := time.Duration(exp)

	switch a.strategy {
	case EqualJitter:
		half := capDuration / 2
		rnd := time.Duration(a.rng.Int63n(int64(half + 1)))
		return half + rnd
	case Decorrelated:
		rnd := time.Duration(a.rng.Int63n(int64(capDuration + 1)))
		if rnd < a.base {
			rnd = a.base
		}
		return rnd
	case FullJitter:
		fallthrough
	default:
		return time.Duration(a.rng.Int63n(int64(capDuration + 1)))
	}
}
