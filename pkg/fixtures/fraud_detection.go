package fixtures

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

// FraudDetectionPipeline builds a streaming anomaly detection DAG for real-time payment fraud.
func FraudDetectionPipeline() *core.WorkflowDefinition {
	return &core.WorkflowDefinition{
		ID:          core.NewID("wf_fraud_detection"),
		TenantID:    "fintech-corp",
		Name:        "Streaming Fraud Anomaly Detection",
		Version:     1,
		Description: "Real-time multi-signal transaction risk scoring with ML ensemble and rule-based decisioning",
		Timeout:     30 * time.Second,
		Steps: []core.StepDefinition{
			{
				ID:       "ingest-transaction",
				TaskType: "tx_ingest",
			},
			{
				ID:        "feature-extraction",
				TaskType:  "feature_extractor",
				DependsOn: []string{"ingest-transaction"},
			},
			{
				ID:        "velocity-check",
				TaskType:  "velocity_checker",
				DependsOn: []string{"ingest-transaction"},
			},
			{
				ID:        "geo-anomaly-check",
				TaskType:  "geo_anomaly",
				DependsOn: []string{"ingest-transaction"},
			},
			{
				ID:        "ml-risk-score",
				TaskType:  "ml_scorer",
				DependsOn: []string{"feature-extraction"},
			},
			{
				ID:        "rule-engine",
				TaskType:  "rule_evaluator",
				DependsOn: []string{"velocity-check", "geo-anomaly-check"},
			},
			{
				ID:        "risk-aggregator",
				TaskType:  "risk_aggregator",
				DependsOn: []string{"ml-risk-score", "rule-engine"},
			},
			{
				ID:        "decision-engine",
				TaskType:  "decision",
				DependsOn: []string{"risk-aggregator"},
			},
		},
	}
}

// TxIngestHandler ingests raw transaction events.
type TxIngestHandler struct{}

func (h *TxIngestHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"tx_id":"TX-8812","amount":4920.00,"merchant":"ONLINE","country":"US","currency":"USD"}`),
	}, nil
}

// FeatureExtractorHandler computes behavioral features.
type FeatureExtractorHandler struct{}

func (h *FeatureExtractorHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"avg_tx_amount":1200.00,"tx_count_7d":34,"new_device":true,"vpn_detected":false}`),
	}, nil
}

// VelocityCheckerHandler checks transaction velocity.
type VelocityCheckerHandler struct{}

func (h *VelocityCheckerHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"velocity_score":0.72,"txns_last_hour":5,"above_threshold":false}`),
	}, nil
}

// GeoAnomalyHandler detects impossible travel or geo-fenced violations.
type GeoAnomalyHandler struct{}

func (h *GeoAnomalyHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"geo_risk":0.15,"country_match":true,"distance_km":0,"impossible_travel":false}`),
	}, nil
}

// MLScorerHandler returns ML ensemble risk score.
type MLScorerHandler struct{}

func (h *MLScorerHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"ml_risk_score":0.31,"model_version":"xgb-v4.1","confidence":0.87}`),
	}, nil
}

// RuleEvaluatorHandler applies deterministic fraud rules.
type RuleEvaluatorHandler struct{}

func (h *RuleEvaluatorHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"rules_triggered":["HIGH_AMOUNT_NEW_DEVICE"],"rule_score":0.55}`),
	}, nil
}

// RiskAggregatorHandler merges ML and rule signals.
type RiskAggregatorHandler struct{}

func (h *RiskAggregatorHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"composite_risk":0.43,"ml_weight":0.6,"rule_weight":0.4}`),
	}, nil
}

// DecisionHandler renders the final fraud decision.
type DecisionHandler struct{}

func (h *DecisionHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	payload := fmt.Sprintf(`{"decision":"REVIEW","reason":"composite_risk=0.43 exceeds soft threshold","ts":%d}`, time.Now().Unix())
	return &worker.StepResult{
		Output: json.RawMessage(payload),
	}, nil
}
