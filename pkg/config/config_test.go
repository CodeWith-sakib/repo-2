package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigPrecedence(t *testing.T) {
	// Clean environment
	cleanEnv := func() {
		os.Unsetenv("KESTREL_CONFIG")
		os.Unsetenv("KESTREL_SERVER_PORT")
		os.Unsetenv("KESTREL_SERVER_HOST")
		os.Unsetenv("KESTREL_DB_TYPE")
		os.Unsetenv("KESTREL_WORKER_CONCURRENCY")
	}
	cleanEnv()
	defer cleanEnv()

	// 1. Defaults only
	cfg1, err := Load([]string{})
	if err != nil {
		t.Fatalf("failed loading default config: %v", err)
	}
	if cfg1.Server.Port != 8080 {
		t.Errorf("expected default port 8080, got %d", cfg1.Server.Port)
	}
	if cfg1.Database.Type != "memory" {
		t.Errorf("expected default db type memory, got %s", cfg1.Database.Type)
	}
	if cfg1.Worker.Concurrency != 4 {
		t.Errorf("expected default concurrency 4, got %d", cfg1.Worker.Concurrency)
	}

	// 2. File overrides Default
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "kestrel.json")
	fileContent := `{
		"server": {"port": 9090, "host": "10.0.0.1"},
		"database": {"type": "postgres"},
		"worker": {"concurrency": 8}
	}`
	if err := os.WriteFile(configPath, []byte(fileContent), 0644); err != nil {
		t.Fatalf("failed writing temp config file: %v", err)
	}

	cfg2, err := Load([]string{"--config", configPath})
	if err != nil {
		t.Fatalf("failed loading config with file: %v", err)
	}
	if cfg2.Server.Port != 9090 {
		t.Errorf("expected file port 9090, got %d", cfg2.Server.Port)
	}
	if cfg2.Database.Type != "postgres" {
		t.Errorf("expected file db type postgres, got %s", cfg2.Database.Type)
	}
	if cfg2.Worker.Concurrency != 8 {
		t.Errorf("expected file concurrency 8, got %d", cfg2.Worker.Concurrency)
	}

	// 3. Env overrides File
	os.Setenv("KESTREL_SERVER_PORT", "9191")
	os.Setenv("KESTREL_WORKER_CONCURRENCY", "12")
	cfg3, err := Load([]string{"--config", configPath})
	if err != nil {
		t.Fatalf("failed loading config with env: %v", err)
	}
	if cfg3.Server.Port != 9191 {
		t.Errorf("expected env port 9191, got %d", cfg3.Server.Port)
	}
	if cfg3.Worker.Concurrency != 12 {
		t.Errorf("expected env concurrency 12, got %d", cfg3.Worker.Concurrency)
	}
	// File values not overridden by env should persist
	if cfg3.Database.Type != "postgres" {
		t.Errorf("expected file db type postgres preserved, got %s", cfg3.Database.Type)
	}

	// 4. CLI Flag overrides Env, File, and Default
	cfg4, err := Load([]string{"--config", configPath, "--port", "9292", "--concurrency", "16", "--db-type", "memory"})
	if err != nil {
		t.Fatalf("failed loading config with flags: %v", err)
	}
	if cfg4.Server.Port != 9292 {
		t.Errorf("expected flag port 9292, got %d", cfg4.Server.Port)
	}
	if cfg4.Worker.Concurrency != 16 {
		t.Errorf("expected flag concurrency 16, got %d", cfg4.Worker.Concurrency)
	}
	if cfg4.Database.Type != "memory" {
		t.Errorf("expected flag db type memory, got %s", cfg4.Database.Type)
	}
}
