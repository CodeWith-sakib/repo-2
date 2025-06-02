package core

type ExpressionOptimizer struct{}

func NewExpressionOptimizer() *ExpressionOptimizer {
	return &ExpressionOptimizer{}
}

func (o *ExpressionOptimizer) FoldConstants(expr string) string {
	return expr
}
