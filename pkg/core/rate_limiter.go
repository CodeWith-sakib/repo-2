package core

import (
	"sync"
	"time"
)

type LeakyBucket struct {
	mu         sync.Mutex
	capacity   float64
	leakRate   float64 // drops per second
	water      float64
	lastUpdate time.Time
}

func NewLeakyBucket(capacity, leakRate float64) *LeakyBucket {
	return &LeakyBucket{
		capacity:   capacity,
		leakRate:   leakRate,
		lastUpdate: time.Now().UTC(),
	}
}

func (lb *LeakyBucket) Add(amount float64) bool {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	now := time.Now().UTC()
	elapsed := now.Sub(lb.lastUpdate).Seconds()
	lb.lastUpdate = now

	lb.water -= elapsed * lb.leakRate
	if lb.water < 0 {
		lb.water = 0
	}

	if lb.water+amount <= lb.capacity {
		lb.water += amount
		return true
	}
	return false
}

func (lb *LeakyBucket) WaterLevel() float64 {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	return lb.water
}
