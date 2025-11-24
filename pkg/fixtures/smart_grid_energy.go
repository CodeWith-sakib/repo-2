package fixtures

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

// SmartGridEnergyPipeline builds an automated demand-response energy grid balancing DAG.
func SmartGridEnergyPipeline() *core.WorkflowDefinition {
	return &core.WorkflowDefinition{
		ID:          core.NewID("wf_smart_grid"),
		TenantID:    "grid-operator-iso",
		Name:        "Renewable Grid Balancing & Demand Response Pipeline",
		Version:     1,
		Description: "Sub-second AMI meter ingestion, solar/wind generation forecast, industrial demand response curtailment, BESS battery storage dispatch, wholesale spot market bidding, and grid frequency stabilization",
		Timeout:     45 * time.Second,
		Steps: []core.StepDefinition{
			{
				ID:       "smart-meter-ingest",
				TaskType: "meter_ingest",
			},
			{
				ID:        "renewable-generation-forecast",
				TaskType:  "forecast_gen",
				DependsOn: []string{"smart-meter-ingest"},
			},
			{
				ID:        "load-curtailment-optimizer",
				TaskType:  "curtail_opt",
				DependsOn: []string{"renewable-generation-forecast"},
			},
			{
				ID:        "battery-storage-dispatch",
				TaskType:  "bess_dispatch",
				DependsOn: []string{"load-curtailment-optimizer"},
			},
			{
				ID:        "spot-market-bidder",
				TaskType:  "market_bid",
				DependsOn: []string{"battery-storage-dispatch"},
			},
			{
				ID:        "grid-frequency-stabilizer",
				TaskType:  "freq_stabilize",
				DependsOn: []string{"spot-market-bidder"},
			},
			{
				ID:        "utility-settlement-ledger",
				TaskType:  "settle_ledger",
				DependsOn: []string{"grid-frequency-stabilizer"},
			},
		},
	}
}

// SmartMeterIngestHandler parses high-frequency AMI smart electrical meter telemetry.
type SmartMeterIngestHandler struct{}

func (h *SmartMeterIngestHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"substation_id":"SUB-WEST-04","current_load_mw":1480.5,"voltage_kv":230,"power_factor":0.98}`),
	}, nil
}

// RenewableForecastHandler calculates solar irradiance and wind velocity output predictions.
type RenewableForecastHandler struct{}

func (h *RenewableForecastHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"solar_forecast_mw":420.0,"wind_forecast_mw":610.5,"net_renewable_mw":1030.5,"deficit_mw":450.0}`),
	}, nil
}

// LoadCurtailmentOptimizerHandler executes demand-response contracts on commercial HVAC and smelters.
type LoadCurtailmentOptimizerHandler struct{}

func (h *LoadCurtailmentOptimizerHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"curtailed_facilities":18,"curtailed_load_mw":120.0,"incentive_payable_usd":3600.00}`),
	}, nil
}

// BESSBatteryDispatchHandler controls utility-scale lithium iron phosphate (LFP) battery storage packs.
type BESSBatteryDispatchHandler struct{}

func (h *BESSBatteryDispatchHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"bess_pack_id":"BESS-MOSS-01","discharge_rate_mw":250.0,"duration_hours":4,"state_of_charge_pct":84.0}`),
	}, nil
}

// SpotMarketBidderHandler submits bids to ISO real-time 5-minute Locational Marginal Price (LMP) markets.
type SpotMarketBidderHandler struct{}

func (h *SpotMarketBidderHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"cleared_mw":80.0,"clearing_price_usd_mwh":42.50,"clearing_interval":"14:00-14:05"}`),
	}, nil
}

// GridFrequencyStabilizerHandler verifies AGC (Automated Generation Control) keeps frequency at 60.00 Hz.
type GridFrequencyStabilizerHandler struct{}

func (h *GridFrequencyStabilizerHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"grid_frequency_hz":60.01,"droop_response_mw":15.0,"frequency_stable":true}`),
	}, nil
}

// UtilitySettlementLedgerHandler records ISO energy and ancillary market revenue credits.
type UtilitySettlementLedgerHandler struct{}

func (h *UtilitySettlementLedgerHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	entryID := fmt.Sprintf("SETTLE-%d", time.Now().UnixNano())
	payload := fmt.Sprintf(`{"settlement_id":"%s","iso_payable_usd":14166.67,"bess_cost_usd":6250.00,"net_margin_usd":7916.67}`, entryID)
	return &worker.StepResult{Output: json.RawMessage(payload)}, nil
}
