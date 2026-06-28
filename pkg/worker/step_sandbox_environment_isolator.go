package worker

import (
	"fmt"
	"strings"
	"sync"
)

// SandboxEnvConfig defines permitted environment variables and paths for step executions.
type SandboxEnvConfig struct {
	AllowedEnvVars []string `json:"allowed_env_vars"`
	RestrictedKeys []string `json:"restricted_keys"`
	WorkingDirBase string   `json:"working_dir_base"`
}

// SandboxEnvironmentIsolator sanitizes host process environment variables before child execution.
type SandboxEnvironmentIsolator struct {
	mu     sync.RWMutex
	config SandboxEnvConfig
}

// NewSandboxEnvironmentIsolator creates a sandbox sanitizer.
func NewSandboxEnvironmentIsolator(cfg SandboxEnvConfig) *SandboxEnvironmentIsolator {
	if len(cfg.RestrictedKeys) == 0 {
		cfg.RestrictedKeys = []string{"AWS_SECRET_ACCESS_KEY", "DATABASE_PASSWORD", "PRIVATE_KEY", "SSH_AUTH_SOCK"}
	}
	return &SandboxEnvironmentIsolator{
		config: cfg,
	}
}

// SanitizeEnv filters raw environment variables against whitelist and blacklist rules.
func (i *SandboxEnvironmentIsolator) SanitizeEnv(rawEnv []string) []string {
	i.mu.RLock()
	defer i.mu.RUnlock()

	sanitized := make([]string, 0, len(rawEnv))

	for _, item := range rawEnv {
		parts := strings.SplitN(item, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := parts[0]

		isRestricted := false
		for _, restricted := range i.config.RestrictedKeys {
			if strings.EqualFold(key, restricted) || strings.Contains(strings.ToUpper(key), "PASSWORD") || strings.Contains(strings.ToUpper(key), "SECRET") {
				isRestricted = true
				break
			}
		}

		if !isRestricted {
			sanitized = append(sanitized, item)
		}
	}

	return sanitized
}

// PrepareWorkingDir returns an isolated execution path for a step run.
func (i *SandboxEnvironmentIsolator) PrepareWorkingDir(workflowID, stepID string) string {
	i.mu.RLock()
	defer i.mu.RUnlock()

	base := i.config.WorkingDirBase
	if base == "" {
		base = "/tmp/kestrelflow_sandbox"
	}
	return fmt.Sprintf("%s/%s/%s", base, workflowID, stepID)
}
