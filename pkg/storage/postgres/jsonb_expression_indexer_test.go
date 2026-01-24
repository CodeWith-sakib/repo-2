package postgres

import (
	"strings"
	"testing"
)

func TestJSONBExpressionIndexer(t *testing.T) {
	indexer := NewJSONBExpressionIndexer()

	def := JSONBIndexDef{
		TableName:  "workflow_runs",
		ColumnName: "metadata",
		JSONPath:   "->'tenant'->>'org_id'",
		DataType:   "text",
		IndexType:  JSONBIndexBTREE,
	}

	idxName, ddl, err := indexer.BuildIndexDDL(def)
	if err != nil {
		t.Fatalf("unexpected error building DDL: %v", err)
	}

	if !strings.HasPrefix(idxName, "idx_workflow_runs_metadata_") {
		t.Errorf("unexpected index name: %s", idxName)
	}

	if !strings.Contains(ddl, "CREATE INDEX CONCURRENTLY IF NOT EXISTS") {
		t.Errorf("expected concurrent index DDL, got %s", ddl)
	}

	if !strings.Contains(ddl, "::text") {
		t.Errorf("expected ::text typecast, got %s", ddl)
	}
}
