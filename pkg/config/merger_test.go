package config

import (
	"testing"
)

func TestConfigMergerDeepMerge(t *testing.T) {
	merger := NewConfigMerger()

	base := map[string]interface{}{
		"server": map[string]interface{}{
			"host": "localhost",
			"port": 8080,
		},
		"database": "postgres://localhost",
		"debug":    false,
	}

	overrides := map[string]interface{}{
		"server": map[string]interface{}{
			"port": 9090,
		},
		"debug": true,
	}

	merged := merger.DeepMerge(base, overrides)

	serverMap := merged["server"].(map[string]interface{})
	if serverMap["host"] != "localhost" {
		t.Errorf("expected host localhost, got %v", serverMap["host"])
	}
	if serverMap["port"] != 9090 {
		t.Errorf("expected port 9090, got %v", serverMap["port"])
	}
	if merged["debug"] != true {
		t.Errorf("expected debug true, got %v", merged["debug"])
	}
}
