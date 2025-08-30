package core

import (
	"fmt"
	"strings"
)

// ExprKind identifies an AST node kind.
type ExprKind string

const (
	ExprLiteral ExprKind = "literal"
	ExprIdent   ExprKind = "ident"
	ExprBinOp   ExprKind = "binop"
	ExprUnary   ExprKind = "unaryop"
	ExprCall    ExprKind = "call"
	ExprParen   ExprKind = "paren"
)

// ExprNode is a node in a workflow expression abstract syntax tree.
type ExprNode struct {
	Kind     ExprKind
	Value    string // literal value or operator
	Children []*ExprNode
}

// ExprRewriter transforms an expression AST by applying rule-based simplifications.
type ExprRewriter struct {
	rules []rewriteRule
}

type rewriteRule struct {
	name    string
	matches func(*ExprNode) bool
	apply   func(*ExprNode) *ExprNode
}

// NewExprRewriter creates a rewriter with the standard algebraic and boolean simplification rules.
func NewExprRewriter() *ExprRewriter {
	r := &ExprRewriter{}

	// Identity: x + 0 → x, x * 1 → x
	r.rules = append(r.rules, rewriteRule{
		name: "additive_identity",
		matches: func(n *ExprNode) bool {
			return n.Kind == ExprBinOp && n.Value == "+" &&
				len(n.Children) == 2 && isLiteralZero(n.Children[1])
		},
		apply: func(n *ExprNode) *ExprNode { return n.Children[0] },
	})

	r.rules = append(r.rules, rewriteRule{
		name: "multiplicative_identity",
		matches: func(n *ExprNode) bool {
			return n.Kind == ExprBinOp && n.Value == "*" &&
				len(n.Children) == 2 && isLiteralOne(n.Children[1])
		},
		apply: func(n *ExprNode) *ExprNode { return n.Children[0] },
	})

	// Zero annihilator: x * 0 → 0
	r.rules = append(r.rules, rewriteRule{
		name: "multiplicative_zero",
		matches: func(n *ExprNode) bool {
			return n.Kind == ExprBinOp && n.Value == "*" &&
				len(n.Children) == 2 && isLiteralZero(n.Children[1])
		},
		apply: func(n *ExprNode) *ExprNode {
			return &ExprNode{Kind: ExprLiteral, Value: "0"}
		},
	})

	// Boolean: true && x → x, false || x → x
	r.rules = append(r.rules, rewriteRule{
		name: "bool_and_true",
		matches: func(n *ExprNode) bool {
			return n.Kind == ExprBinOp && n.Value == "&&" &&
				len(n.Children) == 2 && isLiteralBool(n.Children[0], "true")
		},
		apply: func(n *ExprNode) *ExprNode { return n.Children[1] },
	})

	r.rules = append(r.rules, rewriteRule{
		name: "bool_or_false",
		matches: func(n *ExprNode) bool {
			return n.Kind == ExprBinOp && n.Value == "||" &&
				len(n.Children) == 2 && isLiteralBool(n.Children[0], "false")
		},
		apply: func(n *ExprNode) *ExprNode { return n.Children[1] },
	})

	// Double negation: !!x → x
	r.rules = append(r.rules, rewriteRule{
		name: "double_negation",
		matches: func(n *ExprNode) bool {
			return n.Kind == ExprUnary && n.Value == "!" &&
				len(n.Children) == 1 &&
				n.Children[0].Kind == ExprUnary && n.Children[0].Value == "!" &&
				len(n.Children[0].Children) == 1
		},
		apply: func(n *ExprNode) *ExprNode { return n.Children[0].Children[0] },
	})

	// Paren unwrap: (x) → x
	r.rules = append(r.rules, rewriteRule{
		name: "paren_unwrap",
		matches: func(n *ExprNode) bool {
			return n.Kind == ExprParen && len(n.Children) == 1
		},
		apply: func(n *ExprNode) *ExprNode { return n.Children[0] },
	})

	return r
}

// Rewrite applies all rules bottom-up to the AST, returning the simplified tree.
func (r *ExprRewriter) Rewrite(node *ExprNode) *ExprNode {
	if node == nil {
		return nil
	}

	// Bottom-up: rewrite children first
	simplified := &ExprNode{
		Kind:  node.Kind,
		Value: node.Value,
	}
	for _, child := range node.Children {
		simplified.Children = append(simplified.Children, r.Rewrite(child))
	}

	// Apply first matching rule
	for _, rule := range r.rules {
		if rule.matches(simplified) {
			return r.Rewrite(rule.apply(simplified))
		}
	}

	return simplified
}

// Format renders an ExprNode back to a human-readable expression string.
func (node *ExprNode) Format() string {
	if node == nil {
		return ""
	}
	switch node.Kind {
	case ExprLiteral, ExprIdent:
		return node.Value
	case ExprUnary:
		if len(node.Children) == 1 {
			return node.Value + node.Children[0].Format()
		}
	case ExprBinOp:
		if len(node.Children) == 2 {
			return fmt.Sprintf("%s %s %s", node.Children[0].Format(), node.Value, node.Children[1].Format())
		}
	case ExprParen:
		if len(node.Children) == 1 {
			return "(" + node.Children[0].Format() + ")"
		}
	case ExprCall:
		args := make([]string, len(node.Children))
		for i, c := range node.Children {
			args[i] = c.Format()
		}
		return node.Value + "(" + strings.Join(args, ", ") + ")"
	}
	return node.Value
}

func isLiteralZero(n *ExprNode) bool {
	return n != nil && n.Kind == ExprLiteral && n.Value == "0"
}

func isLiteralOne(n *ExprNode) bool {
	return n != nil && n.Kind == ExprLiteral && n.Value == "1"
}

func isLiteralBool(n *ExprNode, val string) bool {
	return n != nil && n.Kind == ExprLiteral && n.Value == val
}
