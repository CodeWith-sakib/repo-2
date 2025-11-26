package core

import (
	"testing"
)

func TestInfixExprParser_ArithmeticAndPrecedence(t *testing.T) {
	// (2 + 3) * 4 - 6 / 2 = 5 * 4 - 3 = 17
	p, err := NewInfixExprParser("(2 + 3) * 4 - 6 / 2")
	if err != nil {
		t.Fatalf("parser init failed: %v", err)
	}

	val, err := p.Evaluate(nil)
	if err != nil {
		t.Fatalf("evaluate failed: %v", err)
	}

	if val != 17.0 {
		t.Errorf("expected 17.0, got %v", val)
	}
}

func TestInfixExprParser_VariablesAndComparisons(t *testing.T) {
	p, err := NewInfixExprParser("count >= 10 && status_code == 200")
	if err != nil {
		t.Fatalf("parser init failed: %v", err)
	}

	vars := map[string]float64{
		"count":       15,
		"status_code": 200,
	}

	val, err := p.Evaluate(vars)
	if err != nil {
		t.Fatalf("evaluate failed: %v", err)
	}

	if val != 1.0 {
		t.Errorf("expected true (1.0), got %v", val)
	}
}

func TestInfixExprParser_DivisionByZero(t *testing.T) {
	p, _ := NewInfixExprParser("10 / 0")
	_, err := p.Evaluate(nil)
	if err == nil {
		t.Error("expected error on division by zero")
	}
}
