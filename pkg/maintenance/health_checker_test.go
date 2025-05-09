package maintenance

import (
	"context"
	"testing"
	"time"
)

type mockChecker struct {
	status ComponentStatus
}

func (m *mockChecker) CheckHealth(ctx context.Context) ComponentHealth {
	return ComponentHealth{
		Name:      "database",
		Status:    m.status,
		CheckedAt: time.Now().UTC(),
	}
}

func TestSystemDiagnosticsEngine(t *testing.T) {
	engine := NewSystemDiagnosticsEngine()
	engine.Register("database", &mockChecker{status: StatusHealthy})

	results := engine.RunDiagnostics(context.Background())
	dbHealth, exists := results["database"]
	if !exists {
		t.Fatal("expected database health result")
	}
	if dbHealth.Status != StatusHealthy {
		t.Errorf("expected healthy status, got %v", dbHealth.Status)
	}
}
