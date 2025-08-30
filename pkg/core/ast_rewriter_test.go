package core

import (
	"testing"
)

func TestExprRewriter_AdditiveIdentity(t *testing.T) {
	// x + 0 → x
	node := &ExprNode{
		Kind:  ExprBinOp,
		Value: "+",
		Children: []*ExprNode{
			{Kind: ExprIdent, Value: "revenue"},
			{Kind: ExprLiteral, Value: "0"},
		},
	}

	r := NewExprRewriter()
	result := r.Rewrite(node)

	if result.Kind != ExprIdent || result.Value != "revenue" {
		t.Errorf("expected 'revenue', got %s", result.Format())
	}
}

func TestExprRewriter_MultiplicativeZero(t *testing.T) {
	// x * 0 → 0
	node := &ExprNode{
		Kind:  ExprBinOp,
		Value: "*",
		Children: []*ExprNode{
			{Kind: ExprIdent, Value: "cost"},
			{Kind: ExprLiteral, Value: "0"},
		},
	}

	r := NewExprRewriter()
	result := r.Rewrite(node)

	if result.Kind != ExprLiteral || result.Value != "0" {
		t.Errorf("expected literal '0', got %s", result.Format())
	}
}

func TestExprRewriter_DoubleNegation(t *testing.T) {
	// !!x → x
	node := &ExprNode{
		Kind:  ExprUnary,
		Value: "!",
		Children: []*ExprNode{
			{
				Kind:  ExprUnary,
				Value: "!",
				Children: []*ExprNode{
					{Kind: ExprIdent, Value: "active"},
				},
			},
		},
	}

	r := NewExprRewriter()
	result := r.Rewrite(node)

	if result.Kind != ExprIdent || result.Value != "active" {
		t.Errorf("expected 'active' after double negation elimination, got %s", result.Format())
	}
}

func TestExprRewriter_ParenUnwrap(t *testing.T) {
	// (x) → x
	node := &ExprNode{
		Kind: ExprParen,
		Children: []*ExprNode{
			{Kind: ExprIdent, Value: "status"},
		},
	}

	r := NewExprRewriter()
	result := r.Rewrite(node)

	if result.Kind != ExprIdent || result.Value != "status" {
		t.Errorf("expected 'status' after paren unwrap, got %s", result.Format())
	}
}

func TestExprRewriter_BoolShortCircuit(t *testing.T) {
	// true && x → x
	node := &ExprNode{
		Kind:  ExprBinOp,
		Value: "&&",
		Children: []*ExprNode{
			{Kind: ExprLiteral, Value: "true"},
			{Kind: ExprIdent, Value: "enabled"},
		},
	}

	r := NewExprRewriter()
	result := r.Rewrite(node)

	if result.Kind != ExprIdent || result.Value != "enabled" {
		t.Errorf("expected 'enabled' after bool short-circuit, got %s", result.Format())
	}
}
