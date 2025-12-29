package fixtures

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

// DigitalTwinManufacturingPipeline builds an industrial IoT predictive maintenance & digital twin DAG.
func DigitalTwinManufacturingPipeline() *core.WorkflowDefinition {
	return &core.WorkflowDefinition{
		ID:          core.NewID("wf_iot_digital_twin_cnc"),
		TenantID:    "industrial-smart-factory",
		Name:        "CNC High-Speed Spindle Digital Twin & Predictive Maintenance DAG",
		Version:     1,
		Description: "Ingests OPC-UA telemetry from 5-axis CNC machining center, computes vibration FFT spectrum, calculates bearing thermal expansion model, estimates Remaining Useful Life (RUL), and issues maintenance work orders.",
		Timeout:     120 * time.Second,
		Steps: []core.StepDefinition{
			{
				ID:       "opc-ua-telemetry-ingest",
				TaskType: "opcua_telemetry_ingest",
			},
			{
				ID:        "vibration-fft-spectral-analysis",
				TaskType:  "vibration_fft",
				DependsOn: []string{"opc-ua-telemetry-ingest"},
			},
			{
				ID:        "bearing-thermal-expansion-sim",
				TaskType:  "thermal_expansion_model",
				DependsOn: []string{"opc-ua-telemetry-ingest"},
			},
			{
				ID:        "remaining-useful-life-rul-est",
				TaskType:  "rul_estimation",
				DependsOn: []string{"vibration-fft-spectral-analysis", "bearing-thermal-expansion-sim"},
			},
			{
				ID:        "erp-maintenance-work-order-gen",
				TaskType:  "erp_work_order_gen",
				DependsOn: []string{"remaining-useful-life-rul-est"},
			},
		},
	}
}

// OPCUATelemetryHandler reads sensor streams from industrial edge gateways.
type OPCUATelemetryHandler struct{}

func (h *OPCUATelemetryHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"machine_id":"CNC-5AXIS-901","spindle_rpm":18000,"vibration_rms_g":0.42,"spindle_temp_c":48.3,"samples_collected":4096}`),
	}, nil
}

// VibrationFFTHandler calculates Fast Fourier Transform spectral frequencies to detect inner/outer bearing race defects.
type VibrationFFTHandler struct{}

func (h *VibrationFFTHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"peak_frequency_hz":245.5,"harmonics_detected":2,"bpfo_amplitude_mm_s":1.8,"defect_signature":"BEARING_OUTER_RACE_WARNING"}`),
	}, nil
}

// ThermalExpansionHandler computes finite element thermal expansion of ceramic spindle bearings.
type ThermalExpansionHandler struct{}

func (h *ThermalExpansionHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"axial_growth_microns":12.4,"radial_clearance_microns":4.1,"pre_load_loss_pct":6.8,"thermal_state":"EQUILIBRIUM"}`),
	}, nil
}

// RULEstimationHandler evaluates Remaining Useful Life using degradation Weibull distribution models.
type RULEstimationHandler struct{}

func (h *RULEstimationHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"estimated_rul_operating_hours":184.0,"confidence_pct":94.5,"recommended_action":"SCHEDULE_MAINTENANCE_NEXT_SHIFT"}`),
	}, nil
}

// ERPWorkOrderGenHandler creates automated SAP/MES maintenance tickets.
type ERPWorkOrderGenHandler struct{}

func (h *ERPWorkOrderGenHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"work_order_id":"WO-2026-CNC-0042","priority":"HIGH","assigned_crew":"FACILITIES-MECH-A","sap_status":"CREATED"}`),
	}, nil
}
