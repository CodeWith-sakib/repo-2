package maintenance

import (
	"context"
	"sync"
	"time"
)

type ComponentStatus string

const (
	StatusHealthy   ComponentStatus = "HEALTHY"
	StatusDegraded  ComponentStatus = "DEGRADED"
	StatusUnhealthy ComponentStatus = "UNHEALTHY"
)

type ComponentHealth struct {
	Name      string          `json:"name"`
	Status    ComponentStatus `json:"status"`
	Message   string          `json:"message,omitempty"`
	Latency   time.Duration   `json:"latency_ms"`
	CheckedAt time.Time       `json:"checked_at"`
}

type HealthChecker interface {
	CheckHealth(ctx context.Context) ComponentHealth
}

type SystemDiagnosticsEngine struct {
	mu       sync.RWMutex
	checkers map[string]HealthChecker
}

func NewSystemDiagnosticsEngine() *SystemDiagnosticsEngine {
	return &SystemDiagnosticsEngine{
		checkers: make(map[string]HealthChecker),
	}
}

func (e *SystemDiagnosticsEngine) Register(name string, checker HealthChecker) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.checkers[name] = checker
}

func (e *SystemDiagnosticsEngine) RunDiagnostics(ctx context.Context) map[string]ComponentHealth {
	e.mu.RLock()
	defer e.mu.RUnlock()

	results := make(map[string]ComponentHealth)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for name, checker := range e.checkers {
		wg.Add(1)
		go func(n string, chk HealthChecker) {
			defer wg.Done()
			h := chk.CheckHealth(ctx)
			mu.Lock()
			results[n] = h
			mu.Unlock()
		}(name, checker)
	}

	wg.Wait()
	return results
}
