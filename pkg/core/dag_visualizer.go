package core

import (
	"fmt"
	"strings"
)

type DAGVisualizer struct {
	dag *DAG
}

func NewDAGVisualizer(dag *DAG) *DAGVisualizer {
	return &DAGVisualizer{dag: dag}
}

func (v *DAGVisualizer) ToDOT() string {
	var sb strings.Builder
	sb.WriteString("digraph WorkflowDAG {\n")
	sb.WriteString("  rankdir=LR;\n")
	sb.WriteString("  node [shape=box, style=rounded, fontname=\"sans-serif\"];\n")

	nodes := v.dag.Nodes()
	for _, n := range nodes {
		sb.WriteString(fmt.Sprintf("  \"%s\" [label=\"%s\\n(%s)\"];\n", n.ID, n.ID, n.TaskType))
	}

	for _, n := range nodes {
		for _, dep := range n.DependsOn {
			sb.WriteString(fmt.Sprintf("  \"%s\" -> \"%s\";\n", dep, n.ID))
		}
	}

	sb.WriteString("}\n")
	return sb.String()
}

func (v *DAGVisualizer) ToMermaid() string {
	var sb strings.Builder
	sb.WriteString("graph LR\n")

	nodes := v.dag.Nodes()
	for _, n := range nodes {
		sb.WriteString(fmt.Sprintf("    %s[\"%s<br/>(%s)\"]\n", n.ID, n.ID, n.TaskType))
	}

	for _, n := range nodes {
		for _, dep := range n.DependsOn {
			sb.WriteString(fmt.Sprintf("    %s --> %s\n", dep, n.ID))
		}
	}

	return sb.String()
}

func (v *DAGVisualizer) ToASCII() string {
	roots := v.dag.RootNodes()
	var sb strings.Builder
	sb.WriteString("Workflow DAG:\n")

	for _, r := range roots {
		v.renderASCIINode(&sb, r, "", true)
	}
	return sb.String()
}

func (v *DAGVisualizer) renderASCIINode(sb *strings.Builder, nodeID, prefix string, isTail bool) {
	node, exists := v.dag.GetNode(nodeID)
	if !exists {
		return
	}

	connector := "├── "
	if isTail {
		connector = "└── "
	}

	sb.WriteString(fmt.Sprintf("%s%s[%s (%s)]\n", prefix, connector, node.ID, node.TaskType))

	dependents := v.dag.GetDependents(nodeID)
	for i, dep := range dependents {
		childPrefix := prefix + "│   "
		if isTail {
			childPrefix = prefix + "    "
		}
		v.renderASCIINode(sb, dep, childPrefix, i == len(dependents)-1)
	}
}
