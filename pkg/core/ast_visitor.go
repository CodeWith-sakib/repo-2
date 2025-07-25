package core

import (
	"fmt"
	"strings"
)

type ASTVisitor interface {
	VisitLiteral(node *MacroNode) error
	VisitVariable(node *MacroNode) error
	VisitBinaryOp(node *MacroNode) error
}

type ASTWalker struct {
	visitor ASTVisitor
}

func NewASTWalker(visitor ASTVisitor) *ASTWalker {
	return &ASTWalker{visitor: visitor}
}

func (w *ASTWalker) Walk(node *MacroNode) error {
	if node == nil {
		return nil
	}

	switch node.Type {
	case NodeLiteral:
		return w.visitor.VisitLiteral(node)
	case NodeVariable:
		return w.visitor.VisitVariable(node)
	case NodeBinaryOp:
		if err := w.visitor.VisitBinaryOp(node); err != nil {
			return err
		}
		if err := w.Walk(node.Left); err != nil {
			return err
		}
		return w.Walk(node.Right)
	default:
		return fmt.Errorf("unknown node type: %d", node.Type)
	}
}

type VariableCollector struct {
	variables map[string]bool
}

func NewVariableCollector() *VariableCollector {
	return &VariableCollector{variables: make(map[string]bool)}
}

func (c *VariableCollector) VisitLiteral(node *MacroNode) error { return nil }

func (c *VariableCollector) VisitVariable(node *MacroNode) error {
	c.variables[node.Value] = true
	return nil
}

func (c *VariableCollector) VisitBinaryOp(node *MacroNode) error { return nil }

func (c *VariableCollector) Variables() []string {
	var list []string
	for k := range c.variables {
		list = append(list, k)
	}
	return list
}

type ASTStringifier struct {
	sb strings.Builder
}

func NewASTStringifier() *ASTStringifier {
	return &ASTStringifier{}
}

func (s *ASTStringifier) VisitLiteral(node *MacroNode) error {
	s.sb.WriteString(node.Value)
	return nil
}

func (s *ASTStringifier) VisitVariable(node *MacroNode) error {
	s.sb.WriteString(node.Value)
	return nil
}

func (s *ASTStringifier) VisitBinaryOp(node *MacroNode) error {
	s.sb.WriteString("(")
	return nil
}

func (s *ASTStringifier) String() string {
	return s.sb.String()
}
