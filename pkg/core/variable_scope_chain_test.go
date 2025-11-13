package core

import (
	"testing"
)

func TestVariableScopeChain_InheritanceAndShadowing(t *testing.T) {
	root := NewRootScope("root")
	_ = root.Set("globalVar", "global")
	_ = root.Set("shadowVar", "rootValue")

	child := root.NewChildScope("child")
	_ = child.Set("localVar", 42)
	_ = child.Set("shadowVar", "childValue") // shadows root

	grandchild := child.NewChildScope("grandchild")

	// Resolve inherited from root
	val, ok := grandchild.Get("globalVar")
	if !ok || val != "global" {
		t.Errorf("expected globalVar='global', got %v (ok=%v)", val, ok)
	}

	// Resolve local from child
	val, ok = grandchild.Get("localVar")
	if !ok || val != 42 {
		t.Errorf("expected localVar=42, got %v (ok=%v)", val, ok)
	}

	// Resolve shadowed variable: grandchild sees child's shadowed value
	val, ok = grandchild.Get("shadowVar")
	if !ok || val != "childValue" {
		t.Errorf("expected shadowed value 'childValue', got %v", val)
	}

	if grandchild.Depth() != 2 {
		t.Errorf("expected depth 2, got %d", grandchild.Depth())
	}

	// Flattened map
	flat := grandchild.Flatten()
	if flat["globalVar"] != "global" || flat["localVar"] != 42 || flat["shadowVar"] != "childValue" {
		t.Errorf("unexpected flat map: %+v", flat)
	}
}

func TestVariableScopeChain_Freeze(t *testing.T) {
	scope := NewRootScope("ro")
	_ = scope.Set("k", "v1")
	scope.Freeze()

	err := scope.Set("k", "v2")
	if err == nil {
		t.Error("expected error writing to frozen scope")
	}
}
