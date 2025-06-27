package scheduler

import (
	"math"
	"sync"
)

type FairShareDecay struct {
	mu     sync.Mutex
	scores map[string]float64
	factor float64
}

func NewFairShareDecay(decayFactor float64) *FairShareDecay {
	if decayFactor <= 0 || decayFactor >= 1 {
		decayFactor = 0.95
	}
	return &FairShareDecay{
		scores: make(map[string]float64),
		factor: decayFactor,
	}
}

func (d *FairShareDecay) Decay() {
	d.mu.Lock()
	defer d.mu.Unlock()
	for k, v := range d.scores {
		d.scores[k] = math.Round(v*d.factor*100) / 100
	}
}

func (d *FairShareDecay) Record(tenant string, score float64) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.scores[tenant] += score
}

func (d *FairShareDecay) Get(tenant string) float64 {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.scores[tenant]
}
