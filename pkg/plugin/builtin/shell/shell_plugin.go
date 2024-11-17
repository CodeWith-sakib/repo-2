package shell

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

type ShellTaskConfig struct {
	Command string            `json:"command"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	Timeout string            `json:"timeout,omitempty"`
}

type Plugin struct{}

func NewPlugin() *Plugin {
	return &Plugin{}
}

func (p *Plugin) Type() string {
	return "shell"
}

func (p *Plugin) ValidateConfig(cfgJSON json.RawMessage) error {
	if len(cfgJSON) == 0 {
		return fmt.Errorf("%w: shell task config is required", core.ErrValidationFailed)
	}
	var cfg ShellTaskConfig
	if err := json.Unmarshal(cfgJSON, &cfg); err != nil {
		return fmt.Errorf("%w: invalid shell task config: %v", core.ErrValidationFailed, err)
	}
	if cfg.Command == "" {
		return fmt.Errorf("%w: shell command is required", core.ErrValidationFailed)
	}
	return nil
}

func (p *Plugin) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	if err := p.ValidateConfig(sctx.Input); err != nil {
		return nil, err
	}

	var cfg ShellTaskConfig
	_ = json.Unmarshal(sctx.Input, &cfg)

	timeout := 60 * time.Second
	if cfg.Timeout != "" {
		if d, err := time.ParseDuration(cfg.Timeout); err == nil && d > 0 {
			timeout = d
		}
	}

	cmdCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, cfg.Command, cfg.Args...)

	// Inject safe environment variables
	for k, v := range cfg.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}
	cmd.Env = append(cmd.Env, fmt.Sprintf("KESTREL_RUN_ID=%s", sctx.RunID))
	cmd.Env = append(cmd.Env, fmt.Sprintf("KESTREL_STEP_ID=%s", sctx.StepID))

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return nil, fmt.Errorf("shell execution failure: %w", err)
		}
	}

	resultMap := map[string]interface{}{
		"exit_code": exitCode,
		"stdout":    stdout.String(),
		"stderr":    stderr.String(),
	}
	outputJSON, _ := json.Marshal(resultMap)

	if exitCode != 0 {
		return &worker.StepResult{
			Output:       outputJSON,
			ErrorMessage: fmt.Sprintf("command exited with code %d: %s", exitCode, stderr.String()),
			Retryable:    false,
		}, nil
	}

	return &worker.StepResult{
		Output: outputJSON,
	}, nil
}
