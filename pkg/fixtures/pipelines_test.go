package fixtures

import (
	"testing"
)

func TestEnterprisePipelineDefinitions(t *testing.T) {
	etl := BuildETLPipelineDefinition()
	if err := etl.Validate(); err != nil {
		t.Fatalf("etl pipeline validation failed: %v", err)
	}
	if len(etl.Steps) != 3 {
		t.Errorf("expected 3 steps in ETL pipeline, got %d", len(etl.Steps))
	}

	ml := BuildMLInferencePipelineDefinition()
	if err := ml.Validate(); err != nil {
		t.Fatalf("ml pipeline validation failed: %v", err)
	}
	if len(ml.Steps) != 3 {
		t.Errorf("expected 3 steps in ML pipeline, got %d", len(ml.Steps))
	}
}
