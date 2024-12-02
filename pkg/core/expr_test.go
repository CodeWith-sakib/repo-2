package core

import (
	"testing"
)

func TestExpressionEvaluation(t *testing.T) {
	ctx := map[string]interface{}{
		"status": "success",
		"code":   200,
		"meta": map[string]interface{}{
			"retry": false,
			"score": 95.5,
		},
	}

	cases := []struct {
		expr     string
		expected bool
	}{
		{"status == 'success'", true},
		{"status == 'failed'", false},
		{"code == 200", true},
		{"code >= 200 && code < 300", true},
		{"meta.score > 90", true},
		{"meta.retry == false", true},
		{"status == 'success' && (code == 200 || code == 201)", true},
	}

	for _, tc := range cases {
		t.Run(tc.expr, func(t *testing.T) {
			res, err := EvaluateExpression(tc.expr, ctx)
			if err != nil {
				t.Fatalf("evaluation error: %v", err)
			}
			if res != tc.expected {
				t.Errorf("expr %q: got %v, expected %v", tc.expr, res, tc.expected)
			}
		})
	}
}
