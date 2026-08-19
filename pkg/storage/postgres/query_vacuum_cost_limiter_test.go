package postgres

import (
	"testing"
)

func TestVacuumCostLimiter(t *testing.T) {
	limiter := NewVacuumCostLimiter(VacuumCostConfig{})

	sqlHigh := limiter.AdjustForHighLoad()
	cfgHigh := limiter.CurrentConfig()
	if cfgHigh.CostLimit != 100 || cfgHigh.CostDelayMs != 10 {
		t.Errorf("unexpected high load config: %+v", cfgHigh)
	}
	if sqlHigh == "" {
		t.Error("expected valid SQL")
	}

	sqlLow := limiter.AdjustForOffPeak()
	cfgLow := limiter.CurrentConfig()
	if cfgLow.CostLimit != 1000 || cfgLow.CostDelayMs != 0 {
		t.Errorf("unexpected off-peak config: %+v", cfgLow)
	}
	if sqlLow == "" {
		t.Error("expected valid SQL")
	}
}
