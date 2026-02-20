package worker

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestExecutionSandboxJail(t *testing.T) {
	tempBase := filepath.Join(os.TempDir(), "kf_test_jail")
	defer os.RemoveAll(tempBase)

	jailMgr, err := NewExecutionSandboxJail(tempBase)
	if err != nil {
		t.Fatalf("unexpected error creating jail manager: %v", err)
	}

	var observedDir string
	err = jailMgr.RunInJail(context.Background(), "run-99", "step-exec", 500*time.Millisecond, func(jailDir string) error {
		observedDir = jailDir
		testFile := filepath.Join(jailDir, "artifact.txt")
		return os.WriteFile(testFile, []byte("ok"), 0600)
	})

	if err != nil {
		t.Fatalf("unexpected jail execution error: %v", err)
	}

	// Verify cleanup after RunInJail exits
	if _, err := os.Stat(observedDir); !os.IsNotExist(err) {
		t.Errorf("expected jail directory %s to be destroyed after run", observedDir)
	}
}
