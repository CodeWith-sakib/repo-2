package fixtures

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

// ETLWarehousePipeline builds a data warehouse batch ETL workflow DAG.
func ETLWarehousePipeline() *core.WorkflowDefinition {
	return &core.WorkflowDefinition{
		ID:          core.NewID("wf_etl_warehouse"),
		TenantID:    "dataplatform-inc",
		Name:        "Data Warehouse ETL Batch Pipeline",
		Version:     1,
		Description: "Full batch extract-validate-transform-partition-load-audit pipeline for OLAP warehouse ingestion",
		Timeout:     120 * time.Second,
		Steps: []core.StepDefinition{
			{
				ID:       "source-extract",
				TaskType: "extract",
			},
			{
				ID:        "schema-validate",
				TaskType:  "schema_validation",
				DependsOn: []string{"source-extract"},
			},
			{
				ID:        "data-cleanse",
				TaskType:  "cleanse",
				DependsOn: []string{"schema-validate"},
			},
			{
				ID:        "dimension-resolve",
				TaskType:  "dimension_lookup",
				DependsOn: []string{"data-cleanse"},
			},
			{
				ID:        "aggregate-metrics",
				TaskType:  "aggregate",
				DependsOn: []string{"dimension-resolve"},
			},
			{
				ID:        "partition-write",
				TaskType:  "partition_writer",
				DependsOn: []string{"aggregate-metrics"},
			},
			{
				ID:        "metadata-update",
				TaskType:  "metadata",
				DependsOn: []string{"partition-write"},
			},
			{
				ID:        "audit-log",
				TaskType:  "audit",
				DependsOn: []string{"metadata-update"},
			},
		},
	}
}

// SourceExtractHandler extracts raw records from source systems.
type SourceExtractHandler struct{}

func (h *SourceExtractHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	payload := fmt.Sprintf(`{"rows_extracted":1450000,"source":"s3://data-lake/raw/events/%s","format":"parquet"}`, time.Now().Format("2006/01/02"))
	return &worker.StepResult{Output: json.RawMessage(payload)}, nil
}

// SchemaValidationHandler validates schema conformance.
type SchemaValidationHandler struct{}

func (h *SchemaValidationHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"valid_rows":1448500,"invalid_rows":1500,"schema_version":"v3.2.1","violations":["null_in_required_field"]}`),
	}, nil
}

// DataCleanseHandler removes invalid records and normalizes values.
type DataCleanseHandler struct{}

func (h *DataCleanseHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"rows_after_cleanse":1448200,"rows_dropped":1800,"null_fills":3400,"type_coercions":12000}`),
	}, nil
}

// DimensionLookupHandler resolves foreign keys to dimension tables.
type DimensionLookupHandler struct{}

func (h *DimensionLookupHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"lookups_performed":5800000,"cache_hits":4200000,"cache_misses":1600000,"unresolved":420}`),
	}, nil
}

// AggregateHandler computes fact table metrics.
type AggregateHandler struct{}

func (h *AggregateHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"fact_rows":1447780,"aggregation_groups":1200,"metric_columns":34,"computation_ms":8420}`),
	}, nil
}

// PartitionWriterHandler writes partitioned output to the data warehouse.
type PartitionWriterHandler struct{}

func (h *PartitionWriterHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	payload := fmt.Sprintf(`{"partitions_written":24,"bytes_written":2847294828,"destination":"bq://warehouse.fact_events$%s"}`, time.Now().Format("20060102"))
	return &worker.StepResult{Output: json.RawMessage(payload)}, nil
}

// MetadataHandler updates catalog metadata.
type MetadataHandler struct{}

func (h *MetadataHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"catalog_updated":true,"last_modified":"2026-03-15T02:00:00Z","row_count_delta":1447780}`),
	}, nil
}

// ETLAuditHandler writes the audit trail record.
type ETLAuditHandler struct{}

func (h *ETLAuditHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	payload := fmt.Sprintf(`{"audit_id":"aud-%d","pipeline":"etl_warehouse","status":"success","duration_s":47}`, time.Now().UnixNano())
	return &worker.StepResult{Output: json.RawMessage(payload)}, nil
}
