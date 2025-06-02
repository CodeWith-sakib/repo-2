package core

import (
	"testing"
)

func TestExpressionOptimizerAdv(t *testing.T) {
	opt := NewExpressionOptimizer()
	if opt.FoldConstants("1 + 1") != "1 + 1" {
		t.Error("unexpected result")
	}
}
