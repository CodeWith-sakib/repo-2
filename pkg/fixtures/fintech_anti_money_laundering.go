package fixtures

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

// AMLSanctionsPipeline builds an enterprise AML & sanctions compliance screening DAG.
func AMLSanctionsPipeline() *core.WorkflowDefinition {
	return &core.WorkflowDefinition{
		ID:          core.NewID("wf_aml_compliance"),
		TenantID:    "neobank-corp",
		Name:        "Anti-Money Laundering (AML) & OFAC Sanctions Pipeline",
		Version:     1,
		Description: "Automated high-value payment screening against OFAC SDN lists, PEP databases, structuring velocity detection, risk scoring, and FinCEN SAR case generation",
		Timeout:     60 * time.Second,
		Steps: []core.StepDefinition{
			{
				ID:       "transaction-ingest",
				TaskType: "aml_ingest",
			},
			{
				ID:        "ofac-sanctions-lookup",
				TaskType:  "ofac_lookup",
				DependsOn: []string{"transaction-ingest"},
			},
			{
				ID:        "pep-screener",
				TaskType:  "pep_screen",
				DependsOn: []string{"ofac-sanctions-lookup"},
			},
			{
				ID:        "structuring-detector",
				TaskType:  "velocity_check",
				DependsOn: []string{"pep-screener"},
			},
			{
				ID:        "risk-score-aggregator",
				TaskType:  "risk_score",
				DependsOn: []string{"structuring-detector"},
			},
			{
				ID:        "sar-generator",
				TaskType:  "sar_generate",
				DependsOn: []string{"risk-score-aggregator"},
			},
			{
				ID:        "compliance-case-creation",
				TaskType:  "case_create",
				DependsOn: []string{"sar-generator"},
			},
		},
	}
}

// AMLTransactionIngestHandler parses high-value cross-border wire transfers.
type AMLTransactionIngestHandler struct{}

func (h *AMLTransactionIngestHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"tx_id":"wire-883902","sender":"Acme Holdings Ltd","beneficiary":"Apex Global Trading","amount_usd":49500.00,"currency":"USD","corridor":"US->CH"}`),
	}, nil
}

// OFACSanctionsLookupHandler matches entity names against the OFAC Specially Designated Nationals (SDN) registry.
type OFACSanctionsLookupHandler struct{}

func (h *OFACSanctionsLookupHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"sdn_match":false,"fuzzy_score":0.12,"watchlist_version":"2026-03-01","cleared":true}`),
	}, nil
}

// PEPScreenerHandler screens beneficiary against Politically Exposed Persons databases.
type PEPScreenerHandler struct{}

func (h *PEPScreenerHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"pep_flag":false,"adverse_media_hits":0,"relationship_tier":"direct"}`),
	}, nil
}

// StructuringDetectorHandler detects smurfing or transactions just under the $10,000 CTR reporting threshold.
type StructuringDetectorHandler struct{}

func (h *StructuringDetectorHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"window_hours":24,"sub_10k_transactions":5,"cumulative_usd":48500.00,"structuring_suspected":true}`),
	}, nil
}

// RiskScoreAggregatorHandler calculates composite AML risk probability score [0-100].
type RiskScoreAggregatorHandler struct{}

func (h *RiskScoreAggregatorHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"composite_risk_score":82.5,"risk_level":"HIGH","primary_factor":"structuring_velocity"}`),
	}, nil
}

// SARGeneratorHandler prepares FinCEN Suspicious Activity Report regulatory filing data.
type SARGeneratorHandler struct{}

func (h *SARGeneratorHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"fincen_narrative_draft":"Subject engaged in multiple high-velocity transfers within 24 hours just below reporting thresholds.","narrative_tokens":148}`),
	}, nil
}

// ComplianceCaseCreationHandler opens an urgent review ticket in the compliance officer queue.
type ComplianceCaseCreationHandler struct{}

func (h *ComplianceCaseCreationHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	caseID := fmt.Sprintf("CASE-AML-%d", time.Now().UnixNano())
	payload := fmt.Sprintf(`{"case_id":"%s","assigned_team":"financial_crimes","priority":"P1","hold_placed_on_funds":true}`, caseID)
	return &worker.StepResult{Output: json.RawMessage(payload)}, nil
}
