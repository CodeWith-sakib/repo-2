package postgres

import (
	"context"
	"testing"
)

func TestTablePartitionRebalancer(t *testing.T) {
	rebalancer := NewTablePartitionRebalancer(nil)

	summaries, maxSkew, err := rebalancer.AnalyzePartitionSkew(context.Background(), "workflow_audit_logs")
	if err != nil {
		t.Fatalf("unexpected error analyzing skew: %v", err)
	}

	if len(summaries) != 3 {
		t.Fatalf("expected 3 mock partitions, got %d", len(summaries))
	}

	if maxSkew <= 0 {
		t.Errorf("expected positive skew percentage, got %f", maxSkew)
	}

	if summaries[0].PartitionName != "workflow_audit_logs_y2026m01" {
		t.Errorf("unexpected partition name: %s", summaries[0].PartitionName)
	}
}
