package fixtures

import (
	"context"
	"strings"
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

func TestGeothermalReservoirHydrodynamicsPipeline(t *testing.T) {
	dag := GeothermalReservoirHydrodynamicsPipeline()
	if dag == nil {
		t.Fatal("expected non-nil DAG")
	}
	if dag.TenantID != "geothermal-energy-consortium" {
		t.Errorf("expected tenant geothermal-energy-consortium, got %s", dag.TenantID)
	}
	if len(dag.Steps) != 3 {
		t.Errorf("expected 3 steps, got %d", len(dag.Steps))
	}

	televiewer := &TeleviewerFractureScanHandler{}
	out1, err := televiewer.Execute(context.Background(), worker.StepContext{})
	if err != nil {
		t.Fatalf("unexpected error executing televiewer: %v", err)
	}
	if !strings.Contains(string(out1.Output), "fracture_density_per_m") {
		t.Errorf("output missing fracture_density_per_m")
	}

	stim := &HydraulicShearStimulationHandler{}
	out2, err := stim.Execute(context.Background(), worker.StepContext{})
	if err != nil {
		t.Fatalf("unexpected error executing stimulation: %v", err)
	}
	if !strings.Contains(string(out2.Output), "injection_pressure_mpa") {
		t.Errorf("output missing injection_pressure_mpa")
	}

	flow := &ThermoHydraulicSolverHandler{}
	out3, err := flow.Execute(context.Background(), worker.StepContext{})
	if err != nil {
		t.Fatalf("unexpected error executing flow solver: %v", err)
	}
	if !strings.Contains(string(out3.Output), "production_temp_celsius") {
		t.Errorf("output missing production_temp_celsius")
	}
}
