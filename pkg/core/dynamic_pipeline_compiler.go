package core

import (
	"fmt"
	"sort"
	"strings"
)

// PipelineStageIR represents an intermediate representation (IR) of an execution stage.
type PipelineStageIR struct {
	StageIndex int
	Parallel   bool
	StepIDs    []string
	DependsOn  []int // Indices of prerequisite stages
}

// CompiledPipelineIR contains the lowered intermediate representation of a workflow.
type CompiledPipelineIR struct {
	WorkflowID  string
	TotalStages int
	Stages      []PipelineStageIR
	CriticalHop int
}

// DynamicPipelineCompiler compiles raw workflow definitions into optimized staged IR execution plans.
type DynamicPipelineCompiler struct{}

// NewDynamicPipelineCompiler creates a new compiler instance.
func NewDynamicPipelineCompiler() *DynamicPipelineCompiler {
	return &DynamicPipelineCompiler{}
}

// Compile lowers a WorkflowDefinition into an ordered slice of PipelineStageIR stages.
func (c *DynamicPipelineCompiler) Compile(wf *WorkflowDefinition) (*CompiledPipelineIR, error) {
	if wf == nil {
		return nil, fmt.Errorf("workflow definition cannot be nil")
	}

	steps := wf.Steps
	if len(steps) == 0 {
		return &CompiledPipelineIR{
			WorkflowID: string(wf.ID),
		}, nil
	}

	// Validate DAG
	dag, err := BuildDAG(steps)
	if err != nil {
		return nil, fmt.Errorf("compilation failed to build DAG: %w", err)
	}

	topoOrder, err := dag.TopologicalSort()
	if err != nil {
		return nil, fmt.Errorf("topological sort failed: %w", err)
	}

	// Compute longest path from any root to each node to assign stage level
	stageLevel := make(map[string]int)
	for _, id := range topoOrder {
		deps := dag.GetDependencies(id)
		maxDepLevel := -1
		for _, d := range deps {
			if lvl := stageLevel[d]; lvl > maxDepLevel {
				maxDepLevel = lvl
			}
		}
		stageLevel[id] = maxDepLevel + 1
	}

	// Group steps by level
	levelGroups := make(map[int][]string)
	maxLevel := 0
	for id, lvl := range stageLevel {
		levelGroups[lvl] = append(levelGroups[lvl], id)
		if lvl > maxLevel {
			maxLevel = lvl
		}
	}

	var stages []PipelineStageIR
	for lvl := 0; lvl <= maxLevel; lvl++ {
		group := levelGroups[lvl]
		sort.Strings(group)

		var stageDeps []int
		if lvl > 0 {
			stageDeps = []int{lvl - 1}
		}

		stages = append(stages, PipelineStageIR{
			StageIndex: lvl,
			Parallel:   len(group) > 1,
			StepIDs:    group,
			DependsOn:  stageDeps,
		})
	}

	return &CompiledPipelineIR{
		WorkflowID:  string(wf.ID),
		TotalStages: len(stages),
		Stages:      stages,
		CriticalHop: maxLevel + 1,
	}, nil
}

// FormatIR produces a human-readable stage breakdown.
func (ir *CompiledPipelineIR) FormatIR() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Compiled Pipeline IR for %s (Total Stages: %d, Critical Hops: %d)\n",
		ir.WorkflowID, ir.TotalStages, ir.CriticalHop))

	for _, st := range ir.Stages {
		mode := "Sequential"
		if st.Parallel {
			mode = "Parallel"
		}
		sb.WriteString(fmt.Sprintf("  Stage [%d] (%s, %d steps): %s\n",
			st.StageIndex, mode, len(st.StepIDs), strings.Join(st.StepIDs, ", ")))
	}

	return sb.String()
}
