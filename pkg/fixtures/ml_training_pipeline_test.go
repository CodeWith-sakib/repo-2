package fixtures

import (
	"context"
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

func TestMLTrainingPipeline_Validate(t *testing.T) {
	pipeline := MLTrainingPipeline()
	if err := pipeline.Validate(); err != nil {
		t.Fatalf("pipeline validation failed: %v", err)
	}
}

func TestMLTrainingPipeline_AllHandlers(t *testing.T) {
	ctx := context.Background()
	sctx := worker.StepContext{RunID: "run-ml-01", StepID: "dataset-split"}

	handlers := []struct {
		name    string
		handler worker.TaskExecutor
	}{
		{"split", &DatasetSplitHandler{}},
		{"features", &FeatureEngineeringHandler{}},
		{"tuning", &HyperparameterTuningHandler{}},
		{"training", &DistributedTrainingHandler{}},
		{"evaluation", &ModelEvaluationHandler{}},
		{"model_card", &ModelCardGeneratorHandler{}},
		{"canary", &CanaryDeploymentHandler{}},
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
