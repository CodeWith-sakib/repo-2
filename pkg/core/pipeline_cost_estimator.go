package core

import (
	"sync"
	"time"
)

// ResourceDemand describes compute, memory, and duration needed for a step.
type ResourceDemand struct {
	CPUCores     float64       `json:"cpu_cores"`
	MemoryMB     int64         `json:"memory_mb"`
	EstDuration  time.Duration `json:"est_duration"`
	NetworkEgressMB int64      `json:"network_egress_mb"`
}

// UnitPricing defines cost per compute/storage unit per hour.
type UnitPricing struct {
	CPUPerHour       float64
	MemoryGBPerHour  float64
	EgressPerGB      float64
}

// PipelineCostEstimate contains total cost projection for a workflow run.
type PipelineCostEstimate struct {
	TotalCostUSD      float64
	ComputeCostUSD    float64
	MemoryCostUSD     float64
	EgressCostUSD     float64
	EstimatedDuration time.Duration
}

// PipelineCostEstimator predicts infrastructure budget demands before DAG launch.
type PipelineCostEstimator struct {
	mu      sync.RWMutex
	pricing UnitPricing
}

// NewPipelineCostEstimator creates a cost estimator with active cloud unit pricing.
func NewPipelineCostEstimator(pricing UnitPricing) *PipelineCostEstimator {
	return &PipelineCostEstimator{
		pricing: pricing,
	}
}

// Estimate evaluates cumulative resource demands across all workflow steps.
func (e *PipelineCostEstimator) Estimate(demands map[string]ResourceDemand) PipelineCostEstimate {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var est PipelineCostEstimate

	for _, d := range demands {
		hours := d.EstDuration.Hours()
		if hours <= 0 {
			hours = 1.0 / 3600.0 // minimum 1 second
		}

		cpuCost := d.CPUCores * e.pricing.CPUPerHour * hours
		memCost := (float64(d.MemoryMB) / 1024.0) * e.pricing.MemoryGBPerHour * hours
		egressCost := (float64(d.NetworkEgressMB) / 1024.0) * e.pricing.EgressPerGB

		est.ComputeCostUSD += cpuCost
		est.MemoryCostUSD += memCost
		est.EgressCostUSD += egressCost
		if d.EstDuration > est.EstimatedDuration {
			est.EstimatedDuration = d.EstDuration
		}
	}

	est.TotalCostUSD = est.ComputeCostUSD + est.MemoryCostUSD + est.EgressCostUSD
	return est
}
