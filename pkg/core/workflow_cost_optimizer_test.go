package core

import (
	"testing"
)

func TestWorkflowCostOptimizer(t *testing.T) {
	catalog := []InstanceOption{
		{InstanceType: "c6i.large", VCPUs: 2, MemoryGB: 4.0, HourlyRate: 0.085},
		{InstanceType: "c6i.xlarge", VCPUs: 4, MemoryGB: 8.0, HourlyRate: 0.170},
		{InstanceType: "r6i.large", VCPUs: 2, MemoryGB: 16.0, HourlyRate: 0.126},
	}

	opt := NewWorkflowCostOptimizer(catalog)

	demands := map[string]ResourceDemand{
		"step-cpu": {
			CPUCores: 4.0,
			MemoryMB: 4096,
		},
		"step-mem": {
			CPUCores: 2.0,
			MemoryMB: 12288,
		},
	}

	recs := opt.RecommendOptimalInstances(demands)
	if len(recs) != 2 {
		t.Fatalf("expected 2 recommendations, got %d", len(recs))
	}

	if recs[0].StepID == "step-cpu" && recs[0].SelectedInstance.InstanceType != "c6i.xlarge" {
		t.Errorf("expected c6i.xlarge for step-cpu, got %s", recs[0].SelectedInstance.InstanceType)
	}
}
