package core

type ASTOptimizer struct{}

func NewASTOptimizer() *ASTOptimizer {
	return &ASTOptimizer{}
}

func (o *ASTOptimizer) Optimize(node ExprNode) ExprNode {
	if node == nil {
		return nil
	}

	switch n := node.(type) {
	case *BinaryOpNode:
		left := o.Optimize(n.Left)
		right := o.Optimize(n.Right)

		// Constant folding for literals
		litLeft, leftIsLit := left.(*LiteralNode)
		litRight, rightIsLit := right.(*LiteralNode)

		if leftIsLit && rightIsLit {
			folded := o.foldBinary(n.Op, litLeft.Value, litRight.Value)
			if folded != nil {
				return folded
			}
		}

		return &BinaryOpNode{Op: n.Op, Left: left, Right: right}

	case *UnaryOpNode:
		expr := o.Optimize(n.Expr)
		if lit, ok := expr.(*LiteralNode); ok {
			if b, ok := lit.Value.(bool); ok && n.Op == TokenNot {
				return &LiteralNode{Value: !b}
			}
		}
		return &UnaryOpNode{Op: n.Op, Expr: expr}

	default:
		return node
	}
}

func (o *ASTOptimizer) foldBinary(op TokenType, left, right interface{}) ExprNode {
	switch op {
	case TokenEqual:
		return &LiteralNode{Value: left == right}
	case TokenNotEqual:
		return &LiteralNode{Value: left != right}
	case TokenAnd:
		lb, lok := left.(bool)
		rb, rok := right.(bool)
		if lok && rok {
			return &LiteralNode{Value: lb && rb}
		}
	case TokenOr:
		lb, lok := left.(bool)
		rb, rok := right.(bool)
		if lok && rok {
			return &LiteralNode{Value: lb || rb}
		}
	}
	return nil
}
