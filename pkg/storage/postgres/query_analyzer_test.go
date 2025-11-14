package postgres

import (
	"testing"
)

func TestQueryPlanAnalyzer_SeqScanDetection(t *testing.T) {
	explainJSON := `[
		{
			"Plan": {
				"Node Type": "Seq Scan",
				"Relation Name": "workflow_runs",
				"Startup Cost": 0.00,
				"Total Cost": 15420.50,
				"Plan Rows": 50000,
				"Plan Width": 64,
				"Filter": "(status = 'FAILED'::text)"
			}
		}
	]`

	analyzer := NewQueryPlanAnalyzer(1000.0, 5000)
	findings, err := analyzer.AnalyzeJSON([]byte(explainJSON))
	if err != nil {
		t.Fatalf("analyzer error: %v", err)
	}

	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}

	f := findings[0]
	if f.Severity != "HIGH" {
		t.Errorf("expected HIGH severity, got %s", f.Severity)
	}
	if f.Table != "workflow_runs" {
		t.Errorf("expected table workflow_runs, got %s", f.Table)
	}
}

func TestQueryPlanAnalyzer_HealthyIndexScan(t *testing.T) {
	explainJSON := `[
		{
			"Plan": {
				"Node Type": "Index Scan",
				"Relation Name": "workflow_runs",
				"Index Name": "idx_runs_status",
				"Startup Cost": 0.42,
				"Total Cost": 8.44,
				"Plan Rows": 1,
				"Plan Width": 64
			}
		}
	]`

	analyzer := NewQueryPlanAnalyzer(1000.0, 5000)
	findings, err := analyzer.AnalyzeJSON([]byte(explainJSON))
	if err != nil {
		t.Fatalf("analyzer error: %v", err)
	}

	if len(findings) != 0 {
		t.Errorf("expected 0 findings for index scan, got %d", len(findings))
	}
}
