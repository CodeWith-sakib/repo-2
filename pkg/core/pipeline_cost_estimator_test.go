package core

import (
	"testing"
	"time"
)

func TestPipelineCostEstimator(t *testing.T) {
	pricing := UnitPricing{
		CPUPerHour:      0.04,
		MemoryGBPerHour: 0.005,
		EgressPerGB:     0.08,
	}

	estimator := NewPipelineCostEstimator(pricing)

	demands := map[string]ResourceDemand{
		"step-1": {
			CPUCores:        4.0,
			MemoryMB:        8192,
			EstDuration:     30 * time.Minute,
			NetworkEgressMB: 2048,
		},
		"step-2": {
			CPUCores:        2.0,
			MemoryMB:        4096,
			EstDuration:     1 * time.Hour,
			NetworkEgressMB: 512,
		},
	}

	est := estimator.Estimate(demands)

	if est.TotalCostUSD <= 0 {
		t.Errorf("expected positive total cost estimate, got %f", est.TotalCostUSD)
	}

	if est.ComputeCostUSD <= 0 {
		t.Errorf("expected positive compute cost, got %f", est.ComputeCostUSD)
	}

	if est.EstimatedDuration != 1*time.Hour {
		t.Errorf("expected max duration 1h, got %v", est.EstimatedDuration)
	}
}
