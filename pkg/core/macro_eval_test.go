package core

import (
	"testing"
)

func TestMacroEvaluator(t *testing.T) {
	vars := map[string]float64{
		"retries": 3,
		"factor":  2,
	}
	evaluator := NewMacroEvaluator(vars)

	expr, err := SimpleBinaryExpr("retries * 10")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	val, err := evaluator.Eval(expr)
	if err != nil {
		t.Fatalf("unexpected eval error: %v", err)
	}
	if val != 30 {
		t.Errorf("expected 30, got %v", val)
	}

	// Division by zero error
	divZero, _ := SimpleBinaryExpr("factor / 0")
	_, err = evaluator.Eval(divZero)
	if err == nil {
		t.Error("expected error for division by zero")
	}
}
