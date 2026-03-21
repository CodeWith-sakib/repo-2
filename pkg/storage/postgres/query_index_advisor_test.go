package postgres

import (
	"context"
	"strings"
	"testing"
)

func TestQueryIndexAdvisor(t *testing.T) {
	advisor := NewQueryIndexAdvisor(nil)

	recs, err := advisor.RecommendMissingIndexes(context.Background(), 500)
	if err != nil {
		t.Fatalf("unexpected advisor error: %v", err)
	}

	if len(recs) != 1 {
		t.Fatalf("expected 1 recommendation, got %d", len(recs))
	}

	if recs[0].TableName != "workflow_runs" {
		t.Errorf("expected workflow_runs recommendation, got %s", recs[0].TableName)
	}

	if !strings.Contains(recs[0].SuggestedIndex, "CREATE INDEX CONCURRENTLY") {
		t.Errorf("expected concurrent index suggestion, got %s", recs[0].SuggestedIndex)
	}
}
