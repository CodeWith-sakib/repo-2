package scheduler

import (
	"testing"
)

func TestBackpressureController_NormalAdmission(t *testing.T) {
	ctrl := NewBackpressureController(DefaultBackpressureThresholds(), 0.5)

	signals := SystemPressureSignals{
		QueueSaturation:  0.1,
		WorkerInFlight:   0.2,
		StorageLatencyMs: 5.0,
		MemoryUsageRatio: 0.3,
	}

	score, level := ctrl.UpdateSignals(signals)
	if level != PressureNormal {
		t.Errorf("expected PressureNormal, got %s (score=%.2f)", level, score)
	}

	if !ctrl.ShouldAdmit(0) {
		t.Error("should admit priority 0 at normal pressure")
	}
}

func TestBackpressureController_ElevatedPressure(t *testing.T) {
	ctrl := NewBackpressureController(DefaultBackpressureThresholds(), 1.0) // alpha=1.0 for instant test

	highLoad := SystemPressureSignals{
		QueueSaturation:  0.95,
		WorkerInFlight:   0.90,
		StorageLatencyMs: 80.0,
		MemoryUsageRatio: 0.85,
	}

	score, level := ctrl.UpdateSignals(highLoad)
	if level != PressureSevere && level != PressureCritical {
		t.Errorf("expected severe or critical, got %s (score=%.2f)", level, score)
	}

	// Low priority task should be shed
	if ctrl.ShouldAdmit(10) {
		t.Error("expected priority 10 to be rejected under high pressure")
	}

	// High priority task should still be admitted
	if !ctrl.ShouldAdmit(95) {
		t.Error("emergency priority 95 should be admitted even under high pressure")
	}
}

func TestBackpressureController_EMASmoothing(t *testing.T) {
	ctrl := NewBackpressureController(DefaultBackpressureThresholds(), 0.2) // slow smoothing

	spike := SystemPressureSignals{
		QueueSaturation:  1.0,
		WorkerInFlight:   1.0,
		MemoryUsageRatio: 1.0,
		StorageLatencyMs: 100.0,
	}

	score, level := ctrl.UpdateSignals(spike)
	// After single update with alpha 0.2, score is ~0.2, still normal
	if level != PressureNormal {
		t.Errorf("expected normal on first spike with slow EMA, got %s (score=%.2f)", level, score)
	}
}
