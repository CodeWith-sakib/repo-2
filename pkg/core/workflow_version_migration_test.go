package core

import (
	"testing"
)

func TestWorkflowVersionMigrator(t *testing.T) {
	migrator := NewWorkflowVersionMigrator()

	rules := []StepMigrationRule{
		{
			SourceStepID: "etl_extract",
			FieldRenames: map[string]string{
				"source_url": "endpoint_uri",
			},
			DefaultValues: map[string]interface{}{
				"timeout_seconds": 60,
			},
		},
	}

	err := migrator.RegisterMigration(1, 2, rules)
	if err != nil {
		t.Fatalf("unexpected error registering migration: %v", err)
	}

	legacyConfig := map[string]interface{}{
		"source_url": "s3://lake/data.csv",
		"format":     "csv",
	}

	migrated, err := migrator.MigrateStepConfig(1, 2, "etl_extract", legacyConfig)
	if err != nil {
		t.Fatalf("unexpected migration error: %v", err)
	}

	if _, exists := migrated["source_url"]; exists {
		t.Error("expected source_url to be renamed")
	}
	if migrated["endpoint_uri"] != "s3://lake/data.csv" {
		t.Errorf("expected endpoint_uri set, got %v", migrated["endpoint_uri"])
	}
	if migrated["timeout_seconds"] != 60 {
		t.Errorf("expected default timeout 60, got %v", migrated["timeout_seconds"])
	}
}
