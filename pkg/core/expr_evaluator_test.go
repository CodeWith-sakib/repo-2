package core

import (
	"math"
	"testing"
)

func TestMathExprEvaluator(t *testing.T) {
	env := map[string]float64{
		"base":  10.0,
		"scale": 2.5,
		"delta": -1.5,
	}

	tests := []struct {
		expr     string
		expected float64
		wantErr  bool
	}{
		{"10 + 20 * 2", 50.0, false},
		{"(10 + 20) * 2", 60.0, false},
		{"base * scale + 5", 30.0, false},
		{"100 / (10 + 10)", 5.0, false},
		{"100 / 0", 0, true}, // division by zero
		{"unknown_var + 1", 0, true},
		{"-base + 20", 10.0, false},
	}

	for _, tc := range tests {
		eval, err := NewMathExprEvaluator(tc.expr, env)
		if err != nil {
			if !tc.wantErr {
				t.Fatalf("unexpected tokenizer error for %s: %v", tc.expr, err)
			}
			continue
		}

		res, err := eval.Evaluate()
		if tc.wantErr {
			if err == nil {
				t.Errorf("expected error for %s, got nil", tc.expr)
			}
		} else {
			if err != nil {
				t.Errorf("unexpected evaluation error for %s: %v", tc.expr, err)
			}
			if math.Abs(res-tc.expected) > 1e-6 {
				t.Errorf("expected %f for %s, got %f", tc.expected, tc.expr, res)
			}
		}
	}
}
