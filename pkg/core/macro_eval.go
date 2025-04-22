package core

import (
	"fmt"
	"strconv"
	"strings"
)

type MacroNodeType int

const (
	NodeLiteral MacroNodeType = iota
	NodeVariable
	NodeBinaryOp
)

type MacroNode struct {
	Type     MacroNodeType
	Value    string
	Operator string
	Left     *MacroNode
	Right    *MacroNode
}

type MacroEvaluator struct {
	vars map[string]float64
}

func NewMacroEvaluator(vars map[string]float64) *MacroEvaluator {
	if vars == nil {
		vars = make(map[string]float64)
	}
	return &MacroEvaluator{vars: vars}
}

func (e *MacroEvaluator) Eval(node *MacroNode) (float64, error) {
	if node == nil {
		return 0, fmt.Errorf("nil node")
	}

	switch node.Type {
	case NodeLiteral:
		val, err := strconv.ParseFloat(node.Value, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid literal %q: %w", node.Value, err)
		}
		return val, nil

	case NodeVariable:
		val, exists := e.vars[node.Value]
		if !exists {
			return 0, fmt.Errorf("unknown variable %q", node.Value)
		}
		return val, nil

	case NodeBinaryOp:
		leftVal, err := e.Eval(node.Left)
		if err != nil {
			return 0, err
		}
		rightVal, err := e.Eval(node.Right)
		if err != nil {
			return 0, err
		}

		switch node.Operator {
		case "+":
			return leftVal + rightVal, nil
		case "-":
			return leftVal - rightVal, nil
		case "*":
			return leftVal * rightVal, nil
		case "/":
			if rightVal == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			return leftVal / rightVal, nil
		default:
			return 0, fmt.Errorf("unsupported operator %q", node.Operator)
		}
	}

	return 0, fmt.Errorf("unknown node type")
}

// SimpleBinaryExpr parses single binary operations like "var + 10"
func SimpleBinaryExpr(expr string) (*MacroNode, error) {
	expr = strings.TrimSpace(expr)
	for _, op := range []string{"+", "-", "*", "/"} {
		idx := strings.Index(expr, " "+op+" ")
		if idx != -1 {
			leftStr := strings.TrimSpace(expr[:idx])
			rightStr := strings.TrimSpace(expr[idx+len(op)+2:])

			leftNode := parseSimpleTerm(leftStr)
			rightNode := parseSimpleTerm(rightStr)

			return &MacroNode{
				Type:     NodeBinaryOp,
				Operator: op,
				Left:     leftNode,
				Right:    rightNode,
			}, nil
		}
	}
	return parseSimpleTerm(expr), nil
}

func parseSimpleTerm(term string) *MacroNode {
	if _, err := strconv.ParseFloat(term, 64); err == nil {
		return &MacroNode{Type: NodeLiteral, Value: term}
	}
	return &MacroNode{Type: NodeVariable, Value: term}
}
