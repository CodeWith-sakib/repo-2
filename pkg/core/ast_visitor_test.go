package core

import (
	"testing"
)

func TestASTWalkerVariableCollector(t *testing.T) {
	node, err := SimpleBinaryExpr("a + b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	collector := NewVariableCollector()
	walker := NewASTWalker(collector)
	if err := walker.Walk(node); err != nil {
		t.Fatalf("walk failed: %v", err)
	}

	vars := collector.Variables()
	if len(vars) != 2 {
		t.Errorf("expected 2 variables, got %d", len(vars))
	}
}
