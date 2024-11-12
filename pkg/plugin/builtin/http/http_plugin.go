package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

type HTTPTaskConfig struct {
	URL            string            `json:"url"`
	Method         string            `json:"method"`
	Headers        map[string]string `json:"headers,omitempty"`
	Body           string            `json:"body,omitempty"`
	Timeout        string            `json:"timeout,omitempty"`
	ExpectedStatus int               `json:"expected_status,omitempty"`
}

type Plugin struct {
	client *http.Client
}

func NewPlugin(client *http.Client) *Plugin {
	if client == nil {
		client = &http.Client{
			Timeout: 30 * time.Second,
		}
	}
	return &Plugin{client: client}
}

func (p *Plugin) Type() string {
	return "http"
}

func (p *Plugin) ValidateConfig(cfgJSON json.RawMessage) error {
	if len(cfgJSON) == 0 {
		return fmt.Errorf("%w: http task config is required", core.ErrValidationFailed)
	}
	var cfg HTTPTaskConfig
	if err := json.Unmarshal(cfgJSON, &cfg); err != nil {
		return fmt.Errorf("%w: invalid http task config: %v", core.ErrValidationFailed, err)
	}
	if cfg.URL == "" {
		return fmt.Errorf("%w: http url is required", core.ErrValidationFailed)
	}
	if cfg.Method == "" {
		return fmt.Errorf("%w: http method is required", core.ErrValidationFailed)
	}
	return nil
}

func (p *Plugin) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	if err := p.ValidateConfig(sctx.Input); err != nil {
		return nil, err
	}

	var cfg HTTPTaskConfig
	_ = json.Unmarshal(sctx.Input, &cfg)

	timeout := 30 * time.Second
	if cfg.Timeout != "" {
		if d, err := time.ParseDuration(cfg.Timeout); err == nil && d > 0 {
			timeout = d
		}
	}

	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var bodyReader io.Reader
	if cfg.Body != "" {
		bodyReader = bytes.NewBufferString(cfg.Body)
	}

	req, err := http.NewRequestWithContext(reqCtx, cfg.Method, cfg.URL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed creating http request: %w", err)
	}

	for k, v := range cfg.Headers {
		req.Header.Set(k, v)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed reading http response body: %w", err)
	}

	expectedStatus := cfg.ExpectedStatus
	if expectedStatus == 0 {
		expectedStatus = http.StatusOK
	}

	if resp.StatusCode != expectedStatus {
		return &worker.StepResult{
			ErrorMessage: fmt.Sprintf("unexpected status %d, expected %d: %s", resp.StatusCode, expectedStatus, string(respBody)),
			Retryable:    resp.StatusCode >= 500,
		}, nil
	}

	outputMap := map[string]interface{}{
		"status": resp.StatusCode,
		"body":   string(respBody),
	}
	outputJSON, _ := json.Marshal(outputMap)

	return &worker.StepResult{
		Output: outputJSON,
	}, nil
}
