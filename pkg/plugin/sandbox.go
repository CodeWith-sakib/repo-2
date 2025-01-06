package plugin

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

type ExecutionPolicy struct {
	AllowedEnvVars  []string
	ForbiddenPrefix []string
	MaxDuration     time.Duration
	MaxOutputBytes  int
}

type Sandbox struct {
	policy ExecutionPolicy
}

func NewSandbox(policy ExecutionPolicy) *Sandbox {
	if policy.MaxDuration <= 0 {
		policy.MaxDuration = 60 * time.Second
	}
	if policy.MaxOutputBytes <= 0 {
		policy.MaxOutputBytes = 10 * 1024 * 1024 // 10MB
	}
	return &Sandbox{policy: policy}
}

func (s *Sandbox) FilterEnvironment(customEnv map[string]string) ([]string, error) {
	var clean []string
	allowedSet := make(map[string]bool)
	for _, env := range s.policy.AllowedEnvVars {
		allowedSet[env] = true
	}

	// Always allow system path basics if not explicitly forbidden
	baseAllowed := []string{"PATH", "LANG", "LC_ALL", "TMPDIR", "TZ"}
	for _, k := range baseAllowed {
		if val, exists := os.LookupEnv(k); exists {
			clean = append(clean, fmt.Sprintf("%s=%s", k, val))
		}
	}

	for k, v := range customEnv {
		for _, prefix := range s.policy.ForbiddenPrefix {
			if strings.HasPrefix(strings.ToUpper(k), strings.ToUpper(prefix)) {
				return nil, fmt.Errorf("env var %s matches forbidden prefix %s", k, prefix)
			}
		}
		if len(allowedSet) > 0 && !allowedSet[k] {
			continue
		}
		clean = append(clean, fmt.Sprintf("%s=%s", k, v))
	}

	return clean, nil
}

func (s *Sandbox) GuardContext(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, s.policy.MaxDuration)
}

func (s *Sandbox) ValidateOutputSize(size int) error {
	if size > s.policy.MaxOutputBytes {
		return errors.New("plugin output exceeded maximum allowed bytes")
	}
	return nil
}
