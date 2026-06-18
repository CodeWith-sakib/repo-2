package worker

import (
	"errors"
	"testing"
)

func TestStepBackpressureRegulator(t *testing.T) {
	reg := NewStepBackpressureRegulator(10, 20, 0.3)

	if !reg.ShouldAccept() {
		t.Error("expected ShouldAccept true at start")
	}
	if reg.CurrentState() != BackpressureNormal {
		t.Errorf("expected NORMAL state, got %v", reg.CurrentState())
	}

	// Exceed high watermark
	reg.UpdateQueueDepth(25)
	if reg.CurrentState() != BackpressureSaturated {
		t.Errorf("expected SATURATED state, got %v", reg.CurrentState())
	}
	if reg.ShouldAccept() {
		t.Error("expected ShouldAccept false when saturated")
	}

	// Lower queue depth, trigger error throttle
	reg.UpdateQueueDepth(5)
	for i := 0; i < 4; i++ {
		reg.RecordExecutionResult(errors.New("fail"))
	}
	for i := 0; i < 6; i++ {
		reg.RecordExecutionResult(nil)
	}
	// 40% error rate > 30% limit
	if reg.CurrentState() != BackpressureThrottled {
		t.Errorf("expected THROTTLED state, got %v", reg.CurrentState())
	}
	if !reg.ShouldAccept() {
		t.Error("expected ShouldAccept true when throttled (still allows work)")
	}
}
