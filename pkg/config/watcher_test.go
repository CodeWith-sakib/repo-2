package config

import (
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func TestConfigWatcherDetectsChanges(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "kestrel.yaml")

	initialYAML := []byte("server:\n  http_port: 8080\n")
	if err := os.WriteFile(cfgPath, initialYAML, 0644); err != nil {
		t.Fatalf("failed writing initial config: %v", err)
	}

	var reloadCount int32
	watcher := NewConfigWatcher(cfgPath, 50*time.Millisecond, func(newCfg *Config) {
		atomic.AddInt32(&reloadCount, 1)
	})

	changed, err := watcher.CheckOnce()
	if err != nil || !changed {
		t.Fatalf("first check expected changed=true, got changed=%v err=%v", changed, err)
	}

	// No change
	changed, _ = watcher.CheckOnce()
	if changed {
		t.Error("expected changed=false when file has not changed")
	}

	// Modify file
	newYAML := []byte("server:\n  http_port: 9090\n")
	_ = os.WriteFile(cfgPath, newYAML, 0644)

	changed, err = watcher.CheckOnce()
	if err != nil || !changed {
		t.Fatalf("expected changed=true after file update: %v", err)
	}

	if atomic.LoadInt32(&reloadCount) != 2 {
		t.Errorf("expected 2 reloads, got %d", reloadCount)
	}
}
