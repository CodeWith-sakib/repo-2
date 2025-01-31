package fixtures

import (
	"encoding/json"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

func BuildETLPipelineDefinition() *core.WorkflowDefinition {
	extractConfig, _ := json.Marshal(map[string]interface{}{
		"driver": "postgres",
		"query":  "SELECT id, user_id, amount, created_at FROM transactions WHERE processed = false",
	})

	transformConfig, _ := json.Marshal(map[string]interface{}{
		"expressions": map[string]string{
			"total": "sum(amount)",
			"batch": "batch_id",
		},
	})

	loadConfig, _ := json.Marshal(map[string]interface{}{
		"url":    "https://warehouse.internal/api/v1/ingest",
		"method": "POST",
	})

	return &core.WorkflowDefinition{
		ID:          core.NewID("wf-etl-daily"),
		TenantID:    "analytics-platform",
		Name:        "Daily Transaction ETL & Aggregation",
		Version:     1,
		Description: "Extracts pending transactions from PostgreSQL, aggregates hourly totals, and loads to warehouse",
		Steps: []core.StepDefinition{
			{
				ID:        "extract-tx",
				Name:      "Extract Unprocessed Transactions",
				TaskType:  "sql",
				Config:    extractConfig,
				Timeout:   5 * time.Minute,
				DependsOn: []string{},
			},
			{
				ID:        "transform-aggregate",
				Name:      "Compute Aggregates and Rollups",
				TaskType:  "transform",
				Config:    transformConfig,
				Timeout:   3 * time.Minute,
				DependsOn: []string{"extract-tx"},
			},
			{
				ID:        "load-warehouse",
				Name:      "Ingest into Analytics Warehouse",
				TaskType:  "http",
				Config:    loadConfig,
				Timeout:   2 * time.Minute,
				DependsOn: []string{"transform-aggregate"},
			},
		},
		Timeout: 30 * time.Minute,
	}
}

func BuildMLInferencePipelineDefinition() *core.WorkflowDefinition {
	fetchData, _ := json.Marshal(map[string]interface{}{
		"url":    "https://feature-store.internal/features/user_4982",
		"method": "GET",
	})

	evalModel, _ := json.Marshal(map[string]interface{}{
		"url":    "https://model-serving.internal/v2/predict",
		"method": "POST",
	})

	publishEvent, _ := json.Marshal(map[string]interface{}{
		"driver": "postgres",
		"query":  "INSERT INTO prediction_logs (user_id, score) VALUES ('user_4982', 0.94)",
	})

	return &core.WorkflowDefinition{
		ID:          core.NewID("wf-ml-scoring"),
		TenantID:    "risk-engine",
		Name:        "Realtime Risk Assessment Scoring",
		Version:     1,
		Description: "Multi-stage machine learning inference scoring for real-time risk assessment",
		Steps: []core.StepDefinition{
			{
				ID:        "fetch-features",
				Name:      "Retrieve Real-Time Features",
				TaskType:  "http",
				Config:    fetchData,
				Timeout:   10 * time.Second,
				DependsOn: []string{},
			},
			{
				ID:        "predict-score",
				Name:      "Compute Model Inference",
				TaskType:  "http",
				Config:    evalModel,
				Timeout:   5 * time.Second,
				DependsOn: []string{"fetch-features"},
			},
			{
				ID:        "log-prediction",
				Name:      "Persist Scoring Audit Log",
				TaskType:  "sql",
				Config:    publishEvent,
				Timeout:   10 * time.Second,
				DependsOn: []string{"predict-score"},
			},
		},
		Timeout: 5 * time.Minute,
	}
}
