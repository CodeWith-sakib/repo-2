package metrics

import (
	"testing"
)

func TestCardinalityLimiter_OverflowProtection(t *testing.T) {
	cfg := CardinalityConfig{
		MaxDistinctPerTag: 3,
		OverflowValue:     "__other__",
	}
	limiter := NewCardinalityLimiter(cfg)

	key := "user_id"

	// First 3 distinct users should pass through
	if v := limiter.SanitizeLabel(key, "u1"); v != "u1" {
		t.Errorf("expected u1, got %s", v)
	}
	if v := limiter.SanitizeLabel(key, "u2"); v != "u2" {
		t.Errorf("expected u2, got %s", v)
	}
	if v := limiter.SanitizeLabel(key, "u3"); v != "u3" {
		t.Errorf("expected u3, got %s", v)
	}

	// Repeated seen value should pass through
	if v := limiter.SanitizeLabel(key, "u1"); v != "u1" {
		t.Errorf("expected u1 on repeated call, got %s", v)
	}

	// 4th distinct user should be mapped to overflow
	if v := limiter.SanitizeLabel(key, "u4"); v != "__other__" {
		t.Errorf("expected __other__, got %s", v)
	}

	if limiter.DistinctCount(key) != 3 {
		t.Errorf("expected distinct count 3, got %d", limiter.DistinctCount(key))
	}
	if limiter.OverflowHits() != 1 {
		t.Errorf("expected 1 overflow hit, got %d", limiter.OverflowHits())
	}
}

func TestCardinalityLimiter_SanitizeMap(t *testing.T) {
	limiter := NewCardinalityLimiter(CardinalityConfig{
		MaxDistinctPerTag: 1,
		OverflowValue:     "__overflow__",
	})

	limiter.SanitizeLabel("endpoint", "/api/v1") // fills capacity for "endpoint"

	m := map[string]string{
		"endpoint": "/api/v2", // should overflow
		"status":   "200",     // new tag, should be allowed
	}

	res := limiter.SanitizeMap(m)
	if res["endpoint"] != "__overflow__" {
		t.Errorf("expected __overflow__ for endpoint, got %s", res["endpoint"])
	}
	if res["status"] != "200" {
		t.Errorf("expected 200 for status, got %s", res["status"])
	}
}
