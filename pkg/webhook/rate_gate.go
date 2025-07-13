package webhook

import (
	"sync"
	"time"
)

type WebhookRateGate struct {
	mu        sync.Mutex
	lastSent  time.Time
	minInterval time.Duration
}

func NewWebhookRateGate(minInterval time.Duration) *WebhookRateGate {
	return &WebhookRateGate{minInterval: minInterval}
}

func (rg *WebhookRateGate) Allow() bool {
	rg.mu.Lock()
	defer rg.mu.Unlock()
	now := time.Now()
	if now.Sub(rg.lastSent) < rg.minInterval {
		return false
	}
	rg.lastSent = now
	return true
}
