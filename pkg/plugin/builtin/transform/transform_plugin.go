package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/plugin"
)

type TransformPlugin struct{}

func NewTransformPlugin() *TransformPlugin {
	return &TransformPlugin{}
}

func (p *TransformPlugin) Name() string {
	return "kestrel.transform"
}

func (p *TransformPlugin) Version() string {
	return "1.0.0"
}

func (p *TransformPlugin) Execute(ctx context.Context, req plugin.ExecutionRequest) (*plugin.ExecutionResponse, error) {
	var config struct {
		Expressions map[string]string `json:"expressions"`
	}

	if len(req.Config) > 0 {
		if err := json.Unmarshal(req.Config, &config); err != nil {
			return nil, fmt.Errorf("invalid transform config: %w", err)
		}
	}

	var input map[string]interface{}
	if len(req.Input) > 0 {
		if err := json.Unmarshal(req.Input, &input); err != nil {
			return nil, fmt.Errorf("invalid transform input: %w", err)
		}
	} else {
		input = make(map[string]interface{})
	}

	transformer := core.NewPayloadTransformer()
	output := make(map[string]interface{})

	for outKey, expr := range config.Expressions {
		val, err := transformer.ExtractPath(input, expr)
		if err != nil {
			return nil, fmt.Errorf("failed evaluating path %s: %w", expr, err)
		}
		output[outKey] = val
	}

	outBytes, err := json.Marshal(output)
	if err != nil {
		return nil, fmt.Errorf("failed marshaling output: %w", err)
	}

	return &plugin.ExecutionResponse{
		Output:   outBytes,
		Metadata: map[string]string{"transformer": "jsonpath"},
	}, nil
}
