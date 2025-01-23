package core

import (
	"testing"
)

func TestASTOptimizerConstantFolding(t *testing.T) {
	opt := NewASTOptimizer()

	// 10 == 10 -> true
	tree := &BinaryOpNode{
		Op:    TokenEqual,
		Left:  &LiteralNode{Value: 10.0},
		Right: &LiteralNode{Value: 10.0},
	}

	optimized := opt.Optimize(tree)
	lit, ok := optimized.(*LiteralNode)
	if !ok || lit.Value != true {
		t.Fatalf("expected folded literal true, got: %v", optimized)
	}

	// true && false -> false
	treeAnd := &BinaryOpNode{
		Op:    TokenAnd,
		Left:  &LiteralNode{Value: true},
		Right: &LiteralNode{Value: false},
	}
	optimizedAnd := opt.Optimize(treeAnd)
	litAnd, ok := optimizedAnd.(*LiteralNode)
	if !ok || litAnd.Value != false {
		t.Fatalf("expected folded literal false, got: %v", optimizedAnd)
	}
}
