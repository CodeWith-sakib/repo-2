package metrics

import (
	"testing"
)

func TestStreamingAnomalyDetector_Detection(t *testing.T) {
	cfg := AnomalyDetectorConfig{
		MinSamples:         10,
		WarningZThreshold:  2.0,
		CriticalZThreshold: 3.0,
	}
	det := NewStreamingAnomalyDetector(cfg)

	// Feed 20 normal values around 100 with low variance
	for i := 0; i < 20; i++ {
		val := 100.0 + float64(i%3) // 100, 101, 102
		res := det.Observe(val)
		if i < 9 && res.IsAnomaly {
			t.Errorf("sample %d should not be anomaly during warmup", i)
		}
	}

	// Normal value should have no anomaly
	normalRes := det.Observe(101.0)
	if normalRes.IsAnomaly {
		t.Errorf("101 should not be anomaly, got %+v", normalRes)
	}

	// Huge spike to 500 should trigger Critical anomaly!
	spikeRes := det.Observe(500.0)
	if !spikeRes.IsAnomaly {
		t.Fatal("expected 500.0 to be flagged as anomaly")
	}
	if spikeRes.Severity != SeverityCritical {
		t.Errorf("expected Critical severity, got %s", spikeRes.Severity)
	}
	if spikeRes.ZScore < 3.0 {
		t.Errorf("expected z-score >= 3.0, got %v", spikeRes.ZScore)
	}

	_, _, _, anomalies := det.Stats()
	if anomalies != 1 {
		t.Errorf("expected 1 anomaly recorded, got %d", anomalies)
	}
}
