package plugin

import (
	"context"
	"testing"
)

func TestProcessSandbox_FilteredEnv(t *testing.T) {
	cfg := DefaultSandboxConfig()
	s := NewProcessSandbox(cfg)

	env := s.filteredEnv()
	for _, e := range env {
		// Each entry should contain only allowed keys
		allowed := false
		for _, k := range cfg.AllowedEnvKeys {
			if len(e) >= len(k) && e[:len(k)] == k {
				allowed = true
				break
			}
		}
		if !allowed {
			t.Errorf("unexpected env entry leaked through filter: %s", e)
		}
	}
}

func TestProcessSandbox_ViolationRecording(t *testing.T) {
	s := NewProcessSandbox(DefaultSandboxConfig())

	s.RecordViolation("filesystem", "attempted write to /etc/passwd")
	s.RecordViolation("network", "attempted outbound connection to 8.8.8.8")

	violations := s.Violations()
	if len(violations) != 2 {
		t.Errorf("expected 2 violations, got %d", len(violations))
	}
	if violations[0].Category != "filesystem" {
		t.Errorf("unexpected category: %s", violations[0].Category)
	}
}

func TestProcessSandbox_ActiveCount(t *testing.T) {
	s := NewProcessSandbox(DefaultSandboxConfig())

	_, id1, err := s.BuildCommand(context.Background(), "echo", "hello")
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}

	if s.ActiveProcessCount() != 1 {
		t.Errorf("expected 1 active process, got %d", s.ActiveProcessCount())
	}

	s.ReleaseProcess(id1)

	if s.ActiveProcessCount() != 0 {
		t.Errorf("expected 0 after release, got %d", s.ActiveProcessCount())
	}
}
