package events

import (
	"testing"
)

func TestEventFilterExpressionEvaluator(t *testing.T) {
	eval := NewEventFilterExpressionEvaluator()

	rules := []HeaderFilterRule{
		{Attribute: "source", Operator: HeaderOpPrefix, Value: "payment-"},
		{Attribute: "env", Operator: HeaderOpEquals, Value: "production"},
	}

	metaPass := map[string]string{
		"source": "payment-gateway-us",
		"env":    "production",
	}
	if !eval.Matches(rules, metaPass) {
		t.Error("expected match for valid metadata")
	}

	metaFail := map[string]string{
		"source": "payment-gateway-us",
		"env":    "staging",
	}
	if eval.Matches(rules, metaFail) {
		t.Error("expected non-match when env is staging")
	}
}
