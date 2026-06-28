package worker

import (
	"strings"
	"testing"
)

func TestSandboxEnvironmentIsolator(t *testing.T) {
	isolator := NewSandboxEnvironmentIsolator(SandboxEnvConfig{
		WorkingDirBase: "/var/run/kestrel",
	})

	raw := []string{
		"PATH=/usr/bin:/bin",
		"APP_ENV=production",
		"DATABASE_PASSWORD=supersecret",
		"SECRET_TOKEN=xyz123",
		"LANG=en_US.UTF-8",
	}

	clean := isolator.SanitizeEnv(raw)
	for _, item := range clean {
		if strings.Contains(item, "PASSWORD") || strings.Contains(item, "SECRET") {
			t.Errorf("found leaked sensitive variable: %s", item)
		}
	}

	if len(clean) != 3 {
		t.Errorf("expected 3 safe variables, got %d", len(clean))
	}

	dir := isolator.PrepareWorkingDir("wf-1", "step-2")
	if dir != "/var/run/kestrel/wf-1/step-2" {
		t.Errorf("unexpected path: %s", dir)
	}
}
