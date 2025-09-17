package fixtures

import (
	"context"
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

func TestETLWarehousePipeline_Validate(t *testing.T) {
	pipeline := ETLWarehousePipeline()
	if err := pipeline.Validate(); err != nil {
		t.Fatalf("pipeline validation failed: %v", err)
	}
}

func TestETLWarehousePipeline_AllHandlers(t *testing.T) {
	ctx := context.Background()
	sctx := worker.StepContext{RunID: "run-etl-001", StepID: "source-extract"}

	handlers := []struct {
		name    string
		handler worker.TaskExecutor
	}{
		{"extract", &SourceExtractHandler{}},
		{"schema", &SchemaValidationHandler{}},
		{"cleanse", &DataCleanseHandler{}},
		{"dimension", &DimensionLookupHandler{}},
		{"aggregate", &AggregateHandler{}},
		{"partitions", &PartitionWriterHandler{}},
		{"metadata", &MetadataHandler{}},
		{"audit", &ETLAuditHandler{}},
	}

	for _, h := range handlers {
		res, err := h.handler.Execute(ctx, sctx)
		if err != nil {
			t.Fatalf("handler %s failed: %v", h.name, err)
		}
		if len(res.Output) == 0 {
			t.Errorf("handler %s returned empty output", h.name)
		}
	}
}
