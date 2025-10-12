package core

import (
	"fmt"
	"sort"
	"strings"
)

// GraphVisualizer generates Graphviz DOT and ASCII text representations of workflow DAGs.
type GraphVisualizer struct{}

// NewGraphVisualizer creates a visualizer instance.
func NewGraphVisualizer() *GraphVisualizer {
	return &GraphVisualizer{}
}

// ToDOT formats a workflow definition into Graphviz DOT language.
func (v *GraphVisualizer) ToDOT(wf *WorkflowDefinition) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("digraph %q {\n", wf.Name))
	sb.WriteString("  rankdir=LR;\n")
	sb.WriteString("  node [shape=box, style=rounded, fontname=\"Helvetica\"];\n")
	sb.WriteString("  edge [fontname=\"Helvetica\"];\n\n")

	// Sort steps for deterministic output
	steps := make([]StepDefinition, len(wf.Steps))
	copy(steps, wf.Steps)
	sort.Slice(steps, func(i, j int) bool {
		return steps[i].ID < steps[j].ID
	})

	// Nodes
	for _, s := range steps {
		sb.WriteString(fmt.Sprintf("  %q [label=%q];\n", s.ID, fmt.Sprintf("%s\\n(%s)", s.ID, s.TaskType)))
	}
	sb.WriteString("\n")

	// Edges
	for _, s := range steps {
		sortedDeps := make([]string, len(s.DependsOn))
		copy(sortedDeps, s.DependsOn)
		sort.Strings(sortedDeps)

		for _, dep := range sortedDeps {
			sb.WriteString(fmt.Sprintf("  %q -> %q;\n", dep, s.ID))
		}
	}

	sb.WriteString("}\n")
	return sb.String()
}

// ToASCII generates an indented hierarchy tree from root nodes down to leaves.
func (v *GraphVisualizer) ToASCII(wf *WorkflowDefinition) string {
	steps := wf.Steps
	dag, err := BuildDAG(steps)
	if err != nil {
		return fmt.Sprintf("Error building DAG: %v", err)
	}

	// Find roots (steps with no dependencies)
	var roots []string
	for _, s := range steps {
		if len(s.DependsOn) == 0 {
			roots = append(roots, s.ID)
		}
	}
	sort.Strings(roots)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Workflow: %s (v%d)\n", wf.Name, wf.Version))

	visited := make(map[string]bool)
	for _, r := range roots {
		v.renderASCIINode(&sb, r, dag, 0, visited)
	}

	return sb.String()
}

func (v *GraphVisualizer) renderASCIINode(sb *strings.Builder, id string, dag *DAG, depth int, visited map[string]bool) {
	indent := strings.Repeat("  ", depth)
	prefix := "└─ "
	if depth == 0 {
		prefix = "• "
	}

	sb.WriteString(fmt.Sprintf("%s%s%s\n", indent, prefix, id))

	dependents := dag.GetDependents(id)
	sort.Strings(dependents)

	for _, child := range dependents {
		v.renderASCIINode(sb, child, dag, depth+1, visited)
	}
}
