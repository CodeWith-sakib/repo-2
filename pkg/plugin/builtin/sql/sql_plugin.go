package sql

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

type SQLTaskConfig struct {
	Driver   string `json:"driver"`
	DSN      string `json:"dsn"`
	Query    string `json:"query"`
	MaxRows  int    `json:"max_rows"`
	Timeout  string `json:"timeout"`
}

type SQLPlugin struct{}

func NewSQLPlugin() *SQLPlugin {
	return &SQLPlugin{}
}

func (p *SQLPlugin) Type() string {
	return "sql"
}

func (p *SQLPlugin) ValidateConfig(config json.RawMessage) error {
	if len(config) == 0 {
		return fmt.Errorf("%w: sql config is required", core.ErrValidationFailed)
	}
	var cfg SQLTaskConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return fmt.Errorf("%w: invalid sql config: %v", core.ErrValidationFailed, err)
	}
	if cfg.Query == "" {
		return fmt.Errorf("%w: sql query is required", core.ErrValidationFailed)
	}
	return nil
}

func (p *SQLPlugin) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	if err := p.ValidateConfig(sctx.Input); err != nil {
		return nil, err
	}

	var cfg SQLTaskConfig
	_ = json.Unmarshal(sctx.Input, &cfg)

	// Simulated query execution returning structured JSON
	results := []map[string]interface{}{
		{"id": 101, "status": "processed", "updated_at": time.Now().UTC().Format(time.RFC3339)},
		{"id": 102, "status": "processed", "updated_at": time.Now().UTC().Format(time.RFC3339)},
	}

	outBytes, err := json.Marshal(map[string]interface{}{
		"rows_affected": 2,
		"records":       results,
		"driver":        cfg.Driver,
	})
	if err != nil {
		return nil, err
	}

	return &worker.StepResult{
		Output: outBytes,
	}, nil
}
