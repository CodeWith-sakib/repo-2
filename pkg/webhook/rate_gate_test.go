package webhook

import (
	"testing"
	"time"
)

func TestWebhookRateGate(t *testing.T) {
	rg := NewWebhookRateGate(20 * time.Millisecond)
	if !rg.Allow() {
		t.Error("expected first allow")
	}
	if rg.Allow() {
		t.Error("expected immediate second allow to be throttled")
	}
	time.Sleep(25 * time.Millisecond)
	if !rg.Allow() {
		t.Error("expected allow after interval")
	}
}
