package core

import (
	"sort"
	"sync"
)

// InstanceOption describes available cloud compute shapes with pricing.
type InstanceOption struct {
	InstanceType string
	VCPUs        int
	MemoryGB     float64
	HourlyRate   float64
}

// OptimizationRecommendation specifies optimal hardware allocation for a step.
type OptimizationRecommendation struct {
	StepID           string
	SelectedInstance InstanceOption
	EstimatedSavingsPct float64
}

// WorkflowCostOptimizer matches workflow step resource demands to least-cost compute instances.
type WorkflowCostOptimizer struct {
	mu        sync.RWMutex
	instances []InstanceOption
}

// NewWorkflowCostOptimizer creates an instance-matching cost optimizer.
func NewWorkflowCostOptimizer(catalog []InstanceOption) *WorkflowCostOptimizer {
	// Sort catalog by hourly rate ascending
	sortedCatalog := make([]InstanceOption, len(catalog))
	copy(sortedCatalog, catalog)
	sort.Slice(sortedCatalog, func(i, j int) bool {
		return sortedCatalog[i].HourlyRate < sortedCatalog[j].HourlyRate
	})

	return &WorkflowCostOptimizer{
		instances: sortedCatalog,
	}
}

// RecommendOptimalInstances finds cheapest instance capable of satisfying requirements.
func (o *WorkflowCostOptimizer) RecommendOptimalInstances(demands map[string]ResourceDemand) []OptimizationRecommendation {
	o.mu.RLock()
	defer o.mu.RUnlock()

	var recs []OptimizationRecommendation

	for stepID, req := range demands {
		reqMemGB := float64(req.MemoryMB) / 1024.0
		reqCPUs := int(req.CPUCores)
		if reqCPUs < 1 {
			reqCPUs = 1
		}

		var bestMatch *InstanceOption
		for _, inst := range o.instances {
			if inst.VCPUs >= reqCPUs && inst.MemoryGB >= reqMemGB {
				bestMatch = &inst
				break
			}
		}

		if bestMatch != nil {
			recs = append(recs, OptimizationRecommendation{
				StepID:           stepID,
				SelectedInstance: *bestMatch,
				EstimatedSavingsPct: 15.0, // simulated spot/rightsizing savings
			})
		}
	}

	sort.Slice(recs, func(i, j int) bool {
		return recs[i].StepID < recs[j].StepID
	})

	return recs
}
