package core

import (
	"sort"
	"time"
)

type CriticalPathNode struct {
	StepID      string
	Duration    time.Duration
	EarliestStart time.Duration
	LatestStart   time.Duration
	Slack         time.Duration
	IsCritical    bool
}

type CriticalPathAnalysis struct {
	TotalDuration time.Duration
	CriticalPath  []string
	Schedule      map[string]*CriticalPathNode
}

func AnalyzeCriticalPath(dag *DAG, durations map[string]time.Duration) *CriticalPathAnalysis {
	nodes := dag.Nodes()
	sorted, _ := dag.TopologicalSort()

	earliest := make(map[string]time.Duration)
	for _, step := range sorted {
		var maxPred time.Duration
		for _, pred := range dag.GetDependencies(step.ID) {
			predDur := durations[pred]
			finish := earliest[pred] + predDur
			if finish > maxPred {
				maxPred = finish
			}
		}
		earliest[step.ID] = maxPred
	}

	var maxProject time.Duration
	for _, step := range nodes {
		finish := earliest[step.ID] + durations[step.ID]
		if finish > maxProject {
			maxProject = finish
		}
	}

	latest := make(map[string]time.Duration)
	for i := len(sorted) - 1; i >= 0; i-- {
		step := sorted[i]
		dependents := dag.GetDependents(step.ID)
		if len(dependents) == 0 {
			latest[step.ID] = maxProject - durations[step.ID]
		} else {
			minSucc := maxProject
			for _, succ := range dependents {
				start := latest[succ]
				if start < minSucc {
					minSucc = start
				}
			}
			latest[step.ID] = minSucc - durations[step.ID]
		}
	}

	analysis := &CriticalPathAnalysis{
		TotalDuration: maxProject,
		Schedule:      make(map[string]*CriticalPathNode),
	}

	for _, step := range nodes {
		slack := latest[step.ID] - earliest[step.ID]
		isCritical := slack <= 0
		if isCritical {
			analysis.CriticalPath = append(analysis.CriticalPath, step.ID)
		}
		analysis.Schedule[step.ID] = &CriticalPathNode{
			StepID:        step.ID,
			Duration:      durations[step.ID],
			EarliestStart: earliest[step.ID],
			LatestStart:   latest[step.ID],
			Slack:         slack,
			IsCritical:    isCritical,
		}
	}

	sort.Strings(analysis.CriticalPath)
	return analysis
}
