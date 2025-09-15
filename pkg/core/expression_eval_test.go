package core

import (
	"encoding/json"
	"testing"
)

func TestExprEvaluator_Arithmetic(t *testing.T) {
	ctx := NewEvalContext().
		Set("x", ExprNum(10)).
		Set("y", ExprNum(4))

	eval := NewExprEvaluator(ctx)

	// x + y = 14
	expr := json.RawMessage(`{"op":"add","left":{"op":"var","name":"x"},"right":{"op":"var","name":"y"}}`)
	result, err := eval.EvalJSON(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.NumVal != 14 {
		t.Errorf("expected 14, got %.2f", result.NumVal)
	}

	// x / y = 2.5
	div := json.RawMessage(`{"op":"div","left":{"op":"var","name":"x"},"right":{"op":"var","name":"y"}}`)
	d, err := eval.EvalJSON(div)
	if err != nil {
		t.Fatalf("div error: %v", err)
	}
	if d.NumVal != 2.5 {
		t.Errorf("expected 2.5, got %.2f", d.NumVal)
	}
}

func TestExprEvaluator_BooleanLogic(t *testing.T) {
	ctx := NewEvalContext().
		Set("active", ExprBool(true)).
		Set("locked", ExprBool(false))

	eval := NewExprEvaluator(ctx)

	// active && !locked → true
	expr := json.RawMessage(`{"op":"and","left":{"op":"var","name":"active"},"right":{"op":"not","expr":{"op":"var","name":"locked"}}}`)
	result, err := eval.EvalJSON(expr)
	if err != nil {
		t.Fatalf("bool eval error: %v", err)
	}
	if !result.Truthy() {
		t.Errorf("expected true for active && !locked, got %v", result)
	}
}

func TestExprEvaluator_Conditional(t *testing.T) {
	ctx := NewEvalContext().Set("score", ExprNum(85))
	eval := NewExprEvaluator(ctx)

	// if score > 80 then "pass" else "fail"
	expr := json.RawMessage(`{"op":"if","cond":{"op":"gt","left":{"op":"var","name":"score"},"right":{"op":"lit","value":80}},"then":{"op":"lit","value":"pass"},"else":{"op":"lit","value":"fail"}}`)
	result, err := eval.EvalJSON(expr)
	if err != nil {
		t.Fatalf("conditional eval error: %v", err)
	}
	if result.StrVal != "pass" {
		t.Errorf("expected 'pass', got %q", result.StrVal)
	}
}

func TestExprEvaluator_StringContains(t *testing.T) {
	ctx := NewEvalContext().Set("msg", ExprStr("hello world"))
	eval := NewExprEvaluator(ctx)

	expr := json.RawMessage(`{"op":"contains","left":{"op":"var","name":"msg"},"right":{"op":"lit","value":"world"}}`)
	result, err := eval.EvalJSON(expr)
	if err != nil {
		t.Fatalf("contains error: %v", err)
	}
	if !result.Truthy() {
		t.Error("expected contains('hello world', 'world') = true")
	}
}

func TestExprEvaluator_DivisionByZero(t *testing.T) {
	ctx := NewEvalContext().Set("n", ExprNum(5))
	eval := NewExprEvaluator(ctx)

	expr := json.RawMessage(`{"op":"div","left":{"op":"var","name":"n"},"right":{"op":"lit","value":0}}`)
	if _, err := eval.EvalJSON(expr); err == nil {
		t.Error("expected division by zero error")
	}
}
