package plugin

import (
	"testing"
	"time"
)

func TestSandboxEnvFilter(t *testing.T) {
	sb := NewSandbox(ExecutionPolicy{
		AllowedEnvVars:  []string{"APP_ENV", "USER_ID"},
		ForbiddenPrefix: []string{"AWS_", "SECRET_"},
		MaxDuration:     5 * time.Second,
	})

	input := map[string]string{
		"APP_ENV":      "production",
		"UNAUTHORIZED": "skip_me",
	}

	env, err := sb.FilterEnvironment(input)
	if err != nil {
		t.Fatalf("unexpected filter error: %v", err)
	}

	foundAppEnv := false
	foundUnauthorized := false
	for _, e := range env {
		if e == "APP_ENV=production" {
			foundAppEnv = true
		}
		if e == "UNAUTHORIZED=skip_me" {
			foundUnauthorized = true
		}
	}

	if !foundAppEnv {
		t.Error("expected APP_ENV in result")
	}
	if foundUnauthorized {
		t.Error("unauthorized env var was not stripped")
	}

	// Test forbidden prefix
	forbiddenInput := map[string]string{
		"SECRET_KEY": "compromised",
	}
	_, err = sb.FilterEnvironment(forbiddenInput)
	if err == nil {
		t.Error("expected forbidden prefix error, got nil")
	}
}
