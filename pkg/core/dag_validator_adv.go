package core

import (
	"fmt"
	"sort"
	"strings"
)

// StructuralIssueType classifies the structural defect found in a DAG.
type StructuralIssueType string

const (
	IssueUnreachableStep   StructuralIssueType = "UNREACHABLE_STEP"
	IssueDanglingReference StructuralIssueType = "DANGLING_REFERENCE"
	IssueCycleDetected     StructuralIssueType = "CYCLE_DETECTED"
	IssueMultipleRoots     StructuralIssueType = "MULTIPLE_ROOTS"
	IssueNoTerminals       StructuralIssueType = "NO_TERMINALS"
	IssueDisconnectedGraph StructuralIssueType = "DISCONNECTED_GRAPH"
)

// StructuralIssue describes a specific problem found in a DAG definition.
type StructuralIssue struct {
	Type        StructuralIssueType
	StepID      string
	Description string
	Remediation string
}

// AdvancedDAGValidationReport summarizes full graph structural health.
type AdvancedDAGValidationReport struct {
	IsValid             bool
	TotalSteps          int
	TotalEdges          int
	RootStepIDs         []string
	TerminalStepIDs     []string
	ConnectedComponents int
	Issues              []StructuralIssue
}

// AdvancedDAGValidator performs deep graph topology and connectivity inspection.
type AdvancedDAGValidator struct{}

// NewAdvancedDAGValidator creates an advanced validator.
func NewAdvancedDAGValidator() *AdvancedDAGValidator {
	return &AdvancedDAGValidator{}
}

// ValidateGraph performs comprehensive topological health checks on the given steps.
func (v *AdvancedDAGValidator) ValidateGraph(steps []StepDefinition) AdvancedDAGValidationReport {
	report := AdvancedDAGValidationReport{
		TotalSteps: len(steps),
		IsValid:    true,
	}

	if len(steps) == 0 {
		return report
	}

	stepMap := make(map[string]StepDefinition, len(steps))
	for _, s := range steps {
		stepMap[s.ID] = s
	}

	// 1. Check for dangling references (dependencies pointing to non-existent steps)
	edgeCount := 0
	for _, s := range steps {
		for _, dep := range s.DependsOn {
			edgeCount++
			if _, exists := stepMap[dep]; !exists {
				report.Issues = append(report.Issues, StructuralIssue{
					Type:        IssueDanglingReference,
					StepID:      s.ID,
					Description: fmt.Sprintf("Step %q depends on non-existent step %q", s.ID, dep),
					Remediation: fmt.Sprintf("Remove dependency %q or add step definition for it", dep),
				})
				report.IsValid = false
			}
		}
	}
	report.TotalEdges = edgeCount

	// 2. Identify roots and terminals
	dependents := make(map[string][]string)
	for _, s := range steps {
		if len(s.DependsOn) == 0 {
			report.RootStepIDs = append(report.RootStepIDs, s.ID)
		}
		for _, dep := range s.DependsOn {
			dependents[dep] = append(dependents[dep], s.ID)
		}
	}
	sort.Strings(report.RootStepIDs)

	for _, s := range steps {
		if len(dependents[s.ID]) == 0 {
			report.TerminalStepIDs = append(report.TerminalStepIDs, s.ID)
		}
	}
	sort.Strings(report.TerminalStepIDs)

	if len(report.RootStepIDs) == 0 {
		report.Issues = append(report.Issues, StructuralIssue{
			Type:        IssueCycleDetected,
			Description: "No root steps found: all steps have dependencies (probable global cycle)",
			Remediation: "Ensure at least one starting step has no dependencies",
		})
		report.IsValid = false
	}

	if len(report.TerminalStepIDs) == 0 {
		report.Issues = append(report.Issues, StructuralIssue{
			Type:        IssueNoTerminals,
			Description: "No terminal steps found: all steps have dependents",
			Remediation: "Ensure leaf steps terminate without circular dependencies",
		})
		report.IsValid = false
	}

	// 3. Check for cycles via standard DAG builder
	dag, err := BuildDAG(steps)
	if err != nil {
		report.Issues = append(report.Issues, StructuralIssue{
			Type:        IssueCycleDetected,
			Description: fmt.Sprintf("Graph validation failed: %v", err),
			Remediation: "Break cyclical step dependencies",
		})
		report.IsValid = false
	} else {
		if _, topoErr := dag.TopologicalSort(); topoErr != nil {
			report.Issues = append(report.Issues, StructuralIssue{
				Type:        IssueCycleDetected,
				Description: fmt.Sprintf("Cycle detected during topological sorting: %v", topoErr),
				Remediation: "Break cyclical step dependencies",
			})
			report.IsValid = false
		}
	}

	// 4. Check connected components (undirected BFS)
	undirected := make(map[string][]string)
	for _, s := range steps {
		for _, dep := range s.DependsOn {
			undirected[s.ID] = append(undirected[s.ID], dep)
			undirected[dep] = append(undirected[dep], s.ID)
		}
	}

	visited := make(map[string]bool)
	compCount := 0
	for _, s := range steps {
		if !visited[s.ID] {
			compCount++
			// Explore component
			queue := []string{s.ID}
			visited[s.ID] = true
			for len(queue) > 0 {
				curr := queue[0]
				queue = queue[1:]
				for _, neighbor := range undirected[curr] {
					if !visited[neighbor] {
						visited[neighbor] = true
						queue = append(queue, neighbor)
					}
				}
			}
		}
	}
	report.ConnectedComponents = compCount

	if compCount > 1 {
		report.Issues = append(report.Issues, StructuralIssue{
			Type:        IssueDisconnectedGraph,
			Description: fmt.Sprintf("Graph has %d disconnected sub-graphs", compCount),
			Remediation: "Connect disjoint workflow segments or split into separate workflow definitions",
		})
	}

	return report
}

// Summary generates an indented human-readable diagnostic report.
func (r AdvancedDAGValidationReport) Summary() string {
	var sb strings.Builder
	status := "VALID"
	if !r.IsValid {
		status = "INVALID"
	}
	sb.WriteString(fmt.Sprintf("DAG Structural Health: %s (Steps: %d, Edges: %d, Components: %d)\n",
		status, r.TotalSteps, r.TotalEdges, r.ConnectedComponents))
	sb.WriteString(fmt.Sprintf("Roots: %s\n", strings.Join(r.RootStepIDs, ", ")))
	sb.WriteString(fmt.Sprintf("Terminals: %s\n", strings.Join(r.TerminalStepIDs, ", ")))

	if len(r.Issues) > 0 {
		sb.WriteString("Issues:\n")
		for _, is := range r.Issues {
			sb.WriteString(fmt.Sprintf("  [%s] %s (Remediation: %s)\n", is.Type, is.Description, is.Remediation))
		}
	}

	return sb.String()
}
