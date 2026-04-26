package core

import (
	"testing"
	"time"
)

func TestSLATracker(t *testing.T) {
	tracker := NewSLATracker()

	tracker.SetPolicy("order_fulfillment", WorkflowSLAPolicy{
		MaxDuration:         100 * time.Millisecond,
		WarningThresholdPct: 50.0,
	})

	start := time.Now().UTC()

	// 20ms elapsed (20%) -> NONE
	if r := tracker.EvaluateRisk("order_fulfillment", start, start.Add(20*time.Millisecond)); r != SLARiskNone {
		t.Errorf("expected SLARiskNone, got %s", r)
	}

	// 60ms elapsed (60%) -> WARNING
	if r := tracker.EvaluateRisk("order_fulfillment", start, start.Add(60*time.Millisecond)); r != SLARiskWarning {
		t.Errorf("expected SLARiskWarning, got %s", r)
	}

	// 95ms elapsed (95%) -> CRITICAL
	if r := tracker.EvaluateRisk("order_fulfillment", start, start.Add(95*time.Millisecond)); r != SLARiskCritical {
		t.Errorf("expected SLARiskCritical, got %s", r)
	}

	// 110ms elapsed (110%) -> BREACHED
	if r := tracker.EvaluateRisk("order_fulfillment", start, start.Add(110*time.Millisecond)); r != SLARiskBreached {
		t.Errorf("expected SLARiskBreached, got %s", r)
	}
}
