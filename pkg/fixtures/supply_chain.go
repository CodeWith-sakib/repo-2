package fixtures

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

// SupplyChainTrackingPipeline builds an end-to-end multi-echelon global logistics fulfillment DAG.
func SupplyChainTrackingPipeline() *core.WorkflowDefinition {
	return &core.WorkflowDefinition{
		ID:          core.NewID("wf_supply_chain"),
		TenantID:    "logistics-corp",
		Name:        "Global Supply Chain Orchestrator",
		Version:     1,
		Description: "Multi-modal freight tracking, customs clearance, temperature-controlled telematics, and dock dispatching",
		Timeout:     3 * time.Hour,
		Steps: []core.StepDefinition{
			{
				ID:       "ingest-manifest",
				TaskType: "manifest_ingest",
			},
			{
				ID:        "customs-compliance-check",
				TaskType:  "customs_verifier",
				DependsOn: []string{"ingest-manifest"},
			},
			{
				ID:        "cold-chain-telematics",
				TaskType:  "telematics_eval",
				DependsOn: []string{"ingest-manifest"},
			},
			{
				ID:        "route-optimization",
				TaskType:  "route_planner",
				DependsOn: []string{"customs-compliance-check"},
			},
			{
				ID:        "dock-appointment-schedule",
				TaskType:  "dock_scheduler",
				DependsOn: []string{"route-optimization", "cold-chain-telematics"},
			},
			{
				ID:        "bill-of-lading-signoff",
				TaskType:  "ebol_signer",
				DependsOn: []string{"dock-appointment-schedule"},
			},
		},
	}
}

// ManifestIngestHandler handles EDI ASN ingestion.
type ManifestIngestHandler struct{}

func (h *ManifestIngestHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"manifest_id":"EDI-856-9920194","containers":48,"status":"manifest_validated"}`),
	}, nil
}

// CustomsVerifierHandler verifies trade regulations.
type CustomsVerifierHandler struct{}

func (h *CustomsVerifierHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"clearance_status":"duty_paid_precleared","hs_code_count":12}`),
	}, nil
}

// TelematicsAuditHandler evaluates temperature integrity.
type TelematicsAuditHandler struct{}

func (h *TelematicsAuditHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"excursions_detected":0,"avg_temperature_c":-19.4,"integrity":"pass"}`),
	}, nil
}

// RoutePlannerHandler computes logistics paths.
type RoutePlannerHandler struct{}

func (h *RoutePlannerHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"mode":"intermodal_rail","eta_hours":34,"carbon_savings_pct":28.5}`),
	}, nil
}

// DockSchedulerHandler allocates logistics facilities.
type DockSchedulerHandler struct{}

func (h *DockSchedulerHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"facility":"OAK-DC-04","bay":14,"slot":"2026-05-12T08:00:00Z"}`),
	}, nil
}

// EBOLSignerHandler manages digital freight transfer.
type EBOLSignerHandler struct{}

func (h *EBOLSignerHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	payload := fmt.Sprintf(`{"title_transferred":true,"signature_digest":"sha256:logistics:%d"}`, time.Now().UnixNano())
	return &worker.StepResult{
		Output: json.RawMessage(payload),
	}, nil
}
