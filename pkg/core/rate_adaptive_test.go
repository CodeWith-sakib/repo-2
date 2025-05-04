package core

import (
	"testing"
)

func TestAdaptiveRateController(t *testing.T) {
	ctrl := NewAdaptiveRateController(10.0, 2.0, 20.0)

	ctrl.OnSuccess()
	if ctrl.CurrentRate() != 11.0 {
		t.Errorf("expected rate 11.0, got %v", ctrl.CurrentRate())
	}

	ctrl.OnThrottle()
	expected := 11.0 * 0.8
	if ctrl.CurrentRate() != expected {
		t.Errorf("expected rate %v, got %v", expected, ctrl.CurrentRate())
	}
}
