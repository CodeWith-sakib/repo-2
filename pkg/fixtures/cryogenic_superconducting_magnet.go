package fixtures

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

// CryogenicMagnetQuenchProtectionPipeline builds a particle accelerator / fusion tokamak quench DAG.
func CryogenicMagnetQuenchProtectionPipeline() *core.WorkflowDefinition {
	return &core.WorkflowDefinition{
		ID:          core.NewID("wf_magnet_quench_protection"),
		TenantID:    "high-energy-physics-lab",
		Name:        "13-Tesla Superconducting Magnet Cryogenic Quench Detection & Energy Dump DAG",
		Version:     1,
		Description: "Monitors sub-millivolt bridge voltages across Nb3Sn coils, detects resistive quench transition within 10ms, fires capacitive quench protection strip heaters, closes solid-state DC thyristor breakers, and routes 500MJ stored inductive energy into subsea dump resistor tanks.",
		Timeout:     30 * time.Second,
		Steps: []core.StepDefinition{
			{
				ID:       "bridge-voltage-coil-telemetry-ingest",
				TaskType: "coil_voltage_ingest",
			},
			{
				ID:        "resistive-quench-transient-detection",
				TaskType:  "quench_transient_detect",
				DependsOn: []string{"bridge-voltage-coil-telemetry-ingest"},
			},
			{
				ID:        "strip-heater-capacitive-firing",
				TaskType:  "quench_heater_fire",
				DependsOn: []string{"resistive-quench-transient-detection"},
			},
			{
				ID:        "thyristor-breaker-fast-discharge",
				TaskType:  "thyristor_breaker_trip",
				DependsOn: []string{"resistive-quench-transient-detection"},
			},
			{
				ID:        "energy-dump-resistor-thermal-dissipate",
				TaskType:  "energy_dump_dissipate",
				DependsOn: []string{"thyristor-breaker-fast-discharge"},
			},
		},
	}
}

// CoilVoltageIngestHandler samples balanced bridge voltage sensors at 100 kHz.
type CoilVoltageIngestHandler struct{}

func (h *CoilVoltageIngestHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"magnet_id":"TOROIDAL-COIL-04","operating_current_amps":16500,"central_field_tesla":13.2,"bath_temp_kelvin":4.2}`),
	}, nil
}

// QuenchTransientDetectHandler identifies voltage spikes exceeding 100mV threshold persisting for >5ms.
type QuenchTransientDetectHandler struct{}

func (h *QuenchTransientDetectHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"quench_voltage_mv":185.4,"hotspot_temp_kelvin":24.5,"quench_detected":true,"response_time_us":450}`),
	}, nil
}

// QuenchHeaterFireHandler discharges capacitive energy banks into strip heaters to homogenize normal zone.
type QuenchHeaterFireHandler struct{}

func (h *QuenchHeaterFireHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"heater_bank_discharged":true,"capacitor_bank_v":900,"propagation_velocity_mps":18.4,"status":"NORMAL_ZONE_HOMOGENIZED"}`),
	}, nil
}

// ThyristorBreakerTripHandler commutates DC circuit breaker to interrupt 16.5kA current.
type ThyristorBreakerTripHandler struct{}

func (h *ThyristorBreakerTripHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"breaker_open_time_ms":4.2,"arc_chute_voltage_kv":2.8,"current_decay_di_dt":-3200,"status":"BREAKER_INTERRUPTED"}`),
	}, nil
}

// EnergyDumpDissipateHandler absorbs stored magnetic energy in water-cooled stainless steel resistor banks.
type EnergyDumpDissipateHandler struct{}

func (h *EnergyDumpDissipateHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"energy_extracted_mj":485.2,"dump_resistor_final_temp_c":185.0,"cooling_flow_l_min":2400,"quench_safely_contained":true}`),
	}, nil
}
