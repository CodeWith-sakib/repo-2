package fixtures

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

// MLTrainingPipeline builds an end-to-end distributed ML model training & deployment DAG.
func MLTrainingPipeline() *core.WorkflowDefinition {
	return &core.WorkflowDefinition{
		ID:          core.NewID("wf_ml_training"),
		TenantID:    "ai-labs",
		Name:        "Distributed ML Training & Canary Deployment",
		Version:     1,
		Description: "Stratified dataset partitioning, feature normalization, distributed hyperparameter search, GPU training, evaluation, model card generation, and canary rollout",
		Timeout:     180 * time.Second,
		Steps: []core.StepDefinition{
			{
				ID:       "dataset-split",
				TaskType: "dataset_splitter",
			},
			{
				ID:        "feature-engineering",
				TaskType:  "feature_engineering",
				DependsOn: []string{"dataset-split"},
			},
			{
				ID:        "hyperparameter-tuning",
				TaskType:  "tuning",
				DependsOn: []string{"feature-engineering"},
			},
			{
				ID:        "model-training",
				TaskType:  "training",
				DependsOn: []string{"hyperparameter-tuning"},
			},
			{
				ID:        "model-evaluation",
				TaskType:  "evaluation",
				DependsOn: []string{"model-training"},
			},
			{
				ID:        "model-card-generator",
				TaskType:  "model_card",
				DependsOn: []string{"model-evaluation"},
			},
			{
				ID:        "canary-deployment",
				TaskType:  "canary_deploy",
				DependsOn: []string{"model-card-generator"},
			},
		},
	}
}

// DatasetSplitHandler splits inputs into train, validation, and test datasets.
type DatasetSplitHandler struct{}

func (h *DatasetSplitHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"train_samples":800000,"val_samples":100000,"test_samples":100000,"stratified":true,"features_count":128}`),
	}, nil
}

// FeatureEngineeringHandler applies scaling, one-hot encoding, and feature crossing.
type FeatureEngineeringHandler struct{}

func (h *FeatureEngineeringHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"transformed_features":256,"scaling":"standard","embedding_dim":32,"null_imputed":420}`),
	}, nil
}

// HyperparameterTuningHandler runs Bayesian search over model hyperparameters.
type HyperparameterTuningHandler struct{}

func (h *HyperparameterTuningHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"trials_evaluated":50,"best_params":{"learning_rate":0.001,"batch_size":256,"num_layers":4,"dropout":0.2},"best_val_loss":0.182}`),
	}, nil
}

// DistributedTrainingHandler performs multi-GPU model convergence training.
type DistributedTrainingHandler struct{}

func (h *DistributedTrainingHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"epochs_trained":30,"gpu_nodes":8,"final_train_loss":0.142,"checkpoint_uri":"s3://ml-models/nlp/v3/model.pt"}`),
	}, nil
}

// ModelEvaluationHandler benchmarks model accuracy, ROC-AUC, and latency against baseline.
type ModelEvaluationHandler struct{}

func (h *ModelEvaluationHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"test_accuracy":0.962,"roc_auc":0.988,"f1_score":0.954,"p95_inference_ms":12.4,"passes_baseline":true}`),
	}, nil
}

// ModelCardGeneratorHandler produces model provenance, fairness audit, and governance documentation.
type ModelCardGeneratorHandler struct{}

func (h *ModelCardGeneratorHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"model_card_id":"mc-nlp-v3","author":"mlops-team","intended_use":"semantic_search","fairness_bias_checked":true}`),
	}, nil
}

// CanaryDeploymentHandler sets up a 5% canary traffic split on inference clusters.
type CanaryDeploymentHandler struct{}

func (h *CanaryDeploymentHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	deploymentID := fmt.Sprintf("dep-canary-%d", time.Now().UnixNano())
	payload := fmt.Sprintf(`{"deployment_id":"%s","traffic_pct":5,"cluster":"prod-inference-us-east","status":"healthy"}`, deploymentID)
	return &worker.StepResult{Output: json.RawMessage(payload)}, nil
}
