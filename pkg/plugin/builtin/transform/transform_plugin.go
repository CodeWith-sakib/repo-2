package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

type TransformConfig struct {
	Expressions map[string]string `json:"expressions"`
}

type TransformPlugin struct{}

func NewTransformPlugin() *TransformPlugin {
	return &TransformPlugin{}
}

func (p *TransformPlugin) Type() string {
	return "transform"
}

func (p *TransformPlugin) ValidateConfig(config json.RawMessage) error {
	if len(config) == 0 {
		return fmt.Errorf("%w: transform config is required", core.ErrValidationFailed)
	}
	var cfg TransformConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return fmt.Errorf("%w: invalid transform config: %v", core.ErrValidationFailed, err)
	}
	return nil
}

func (p *TransformPlugin) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	var config TransformConfig
	if len(sctx.Input) > 0 {
		if err := json.Unmarshal(sctx.Input, &config); err != nil {
			return nil, fmt.Errorf("invalid transform input: %w", err)
		}
	}

	var data map[string]interface{}
	if len(sctx.Input) > 0 {
		_ = json.Unmarshal(sctx.Input, &data)
	} else {
		data = make(map[string]interface{})
	}

	output := make(map[string]interface{})
	for outKey, expr := range config.Expressions {
		val, err := core.ExtractJSONPath(data, expr)
		if err != nil {
			return nil, fmt.Errorf("failed evaluating path %s: %w", expr, err)
		}
		output[outKey] = val
	}

	outBytes, err := json.Marshal(output)
	if err != nil {
		return nil, fmt.Errorf("failed marshaling output: %w", err)
	}

	return &worker.StepResult{
		Output: outBytes,
	}, nil
}
