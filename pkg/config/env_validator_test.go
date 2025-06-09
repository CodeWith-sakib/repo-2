package config

import (
	"os"
	"testing"
)

func TestEnvValidator(t *testing.T) {
	v := NewEnvValidator()
	os.Setenv("KESTREL_TEST_VAR", "present")
	defer os.Unsetenv("KESTREL_TEST_VAR")

	if err := v.Require("KESTREL_TEST_VAR"); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if err := v.Require("KESTREL_NONEXISTENT_VAR"); err == nil {
		t.Error("expected error for nonexistent var")
	}
}
