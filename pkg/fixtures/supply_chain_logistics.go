package fixtures

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

// SupplyChainLogisticsPipeline builds a maritime freight and intermodal logistics DAG.
func SupplyChainLogisticsPipeline() *core.WorkflowDefinition {
	return &core.WorkflowDefinition{
		ID:          core.NewID("wf_supply_chain"),
		TenantID:    "freight-forwarder-intl",
		Name:        "Global Freight Tracking & Intermodal Drayage Pipeline",
		Version:     1,
		Description: "End-to-end AIS container telematics ingestion, customs entry review, cold chain temperature compliance audit, machine-learning ETA prediction, demurrage calculation, and automated drayage dispatch",
		Timeout:     90 * time.Second,
		Steps: []core.StepDefinition{
			{
				ID:       "container-telemetry-ingest",
				TaskType: "telemetry_ingest",
			},
			{
				ID:        "customs-clearance-validator",
				TaskType:  "customs_validate",
				DependsOn: []string{"container-telemetry-ingest"},
			},
			{
				ID:        "cold-chain-temperature-auditor",
				TaskType:  "cold_chain_audit",
				DependsOn: []string{"customs-clearance-validator"},
			},
			{
				ID:        "route-eta-predictor",
				TaskType:  "eta_predict",
				DependsOn: []string{"cold-chain-temperature-auditor"},
			},
			{
				ID:        "demurrage-fee-calculator",
				TaskType:  "demurrage_calc",
				DependsOn: []string{"route-eta-predictor"},
			},
			{
				ID:        "bill-of-lading-generator",
				TaskType:  "bol_generate",
				DependsOn: []string{"demurrage-fee-calculator"},
			},
			{
				ID:        "drayage-dispatch-notifier",
				TaskType:  "drayage_notify",
				DependsOn: []string{"bill-of-lading-generator"},
			},
		},
	}
}

// ContainerTelemetryIngestHandler parses AIS transponder and smart reefer sensor readings.
type ContainerTelemetryIngestHandler struct{}

func (h *ContainerTelemetryIngestHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"container_no":"MSKU-782910-4","vessel_imo":9811002,"vessel_name":"Maersk Mc-Kinney Moller","port_of_entry":"USLAX"}`),
	}, nil
}

// CustomsClearanceValidatorHandler inspects bill of entry against US CBP / ACE automated manifest systems.
type CustomsClearanceValidatorHandler struct{}

func (h *CustomsClearanceValidatorHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"customs_entry_no":"CBP-2026-LAX-4491","hs_code":"8504.40","duty_paid_usd":1240.50,"customs_cleared":true}`),
	}, nil
}

// ColdChainAuditorHandler inspects refrigeration telemetry logs for temperature excursions.
type ColdChainAuditorHandler struct{}

func (h *ColdChainAuditorHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"reefer_setpoint_c":4.0,"avg_temp_c":4.2,"max_temp_c":4.8,"excursion_minutes":0,"cold_chain_certified":true}`),
	}, nil
}

// RouteETAPredictorHandler uses vessel AIS speed, terminal congestion, and berth schedules to estimate berth time.
type RouteETAPredictorHandler struct{}

func (h *RouteETAPredictorHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"estimated_berth_time":"2026-03-12T14:30:00Z","terminal_congestion_hours":8.5,"pilot_board_confirmed":true}`),
	}, nil
}

// DemurrageCalculatorHandler calculates free time and container detention/demurrage fees.
type DemurrageCalculatorHandler struct{}

func (h *DemurrageCalculatorHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"free_days_allowed":4,"days_in_terminal":2,"demurrage_incurred_usd":0.00,"risk_of_fee":false}`),
	}, nil
}

// BillOfLadingGeneratorHandler produces legally binding multimodal ocean Bill of Lading (BOL).
type BillOfLadingGeneratorHandler struct{}

func (h *BillOfLadingGeneratorHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"bol_number":"BOL-MSK-2026-9921","consignee":"Western Retail Distribution LLC","gross_weight_kg":21450,"seal_number":"SEAL-88992"}`),
	}, nil
}

// DrayageDispatchNotifierHandler books a drayage motor carrier for inland port pickup.
type DrayageDispatchNotifierHandler struct{}

func (h *DrayageDispatchNotifierHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	dispatchID := fmt.Sprintf("DRAY-%d", time.Now().UnixNano())
	payload := fmt.Sprintf(`{"dispatch_id":"%s","carrier":"Pacific Drayage Corp","terminal_gate_appointment":"2026-03-13T08:00:00Z","status":"dispatched"}`, dispatchID)
	return &worker.StepResult{Output: json.RawMessage(payload)}, nil
}
