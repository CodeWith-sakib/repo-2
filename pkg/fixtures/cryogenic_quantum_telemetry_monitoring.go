package fixtures

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

// CryogenicQuantumTelemetryPipeline builds a dilution refrigerator dilution stages telemetry monitor DAG.
func CryogenicQuantumTelemetryPipeline() *core.WorkflowDefinition {
	return &core.WorkflowDefinition{
		ID:          core.NewID("wf_dilution_fridge_telemetry"),
		TenantID:    "quantum-processor-foundry",
		Name:        "Sub-10mK Dilution Refrigerator Continuous Telemetry & Superconducting Qubit Readout DAG",
		Version:     1,
		Description: "Monitors 3He/4He circulation condensing pressures, pulse tube stage temperatures (50K, 4K, Still, Cold Plate, Mixing Chamber at 8.5mK), calculates thermal load margins, and schedules qubit XY microwave pulse drive calibration.",
		Timeout:     60 * time.Second,
		Steps: []core.StepDefinition{
			{
				ID:       "fridge-thermal-stage-sensing",
				TaskType: "fridge_thermal_sensing",
			},
			{
				ID:        "he3-circulation-pressure-check",
				TaskType:  "he3_pressure_check",
				DependsOn: []string{"fridge-thermal-stage-sensing"},
			},
			{
				ID:        "qubit-coherence-time-t1-t2-sweep",
				TaskType:  "qubit_coherence_sweep",
				DependsOn: []string{"he3-circulation-pressure-check"},
			},
		},
	}
}

// FridgeThermalSensingHandler samples RuO2 resistance thermometry across all cryogenic plates.
type FridgeThermalSensingHandler struct{}

func (h *FridgeThermalSensingHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"mixing_chamber_temp_mk":8.42,"still_temp_k":0.78,"plate_4k_temp":3.85,"thermal_margin_uw":14.2}`),
	}, nil
}

// He3PressureCheckHandler verifies turbo-molecular pumping speeds and condenser backing pressure.
type He3PressureCheckHandler struct{}

func (h *He3PressureCheckHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"condensing_pressure_bar":1.45,"flow_rate_mmol_per_s":0.35,"compressor_discharge_psi":85.2}`),
	}, nil
}

// QubitCoherenceSweepHandler measures energy relaxation T1 and Hahn echo dephasing T2 for transmons.
type QubitCoherenceSweepHandler struct{}

func (h *QubitCoherenceSweepHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"qubit_id":"Q0","t1_relaxation_us":124.5,"t2_hahn_echo_us":158.2,"single_qubit_fidelity":0.9994}`),
	}, nil
}
