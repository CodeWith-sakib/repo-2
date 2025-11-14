package postgres

import (
	"encoding/json"
	"fmt"
	"strings"
)

// PlanNode represents a node in the PostgreSQL EXPLAIN JSON tree.
type PlanNode struct {
	NodeType     string     `json:"Node Type"`
	RelationName string     `json:"Relation Name,omitempty"`
	Alias        string     `json:"Alias,omitempty"`
	StartupCost  float64    `json:"Startup Cost"`
	TotalCost    float64    `json:"Total Cost"`
	PlanRows     int64      `json:"Plan Rows"`
	PlanWidth    int        `json:"Plan Width"`
	Filter       string     `json:"Filter,omitempty"`
	IndexName    string     `json:"Index Name,omitempty"`
	Plans        []PlanNode `json:"Plans,omitempty"`
}

// ExplainOutput represents the top-level container for EXPLAIN (FORMAT JSON).
type ExplainOutput struct {
	Plan PlanNode `json:"Plan"`
}

// QueryOptimizationFinding highlights a performance bottleneck in an execution plan.
type QueryOptimizationFinding struct {
	Severity       string // "HIGH", "MEDIUM", "LOW"
	Table          string
	NodeType       string
	Cost           float64
	Rows           int64
	Description    string
	Recommendation string
}

// QueryPlanAnalyzer evaluates PostgreSQL query execution plans and flags bottlenecks.
type QueryPlanAnalyzer struct {
	costThreshold float64
	rowsThreshold int64
}

// NewQueryPlanAnalyzer creates an analyzer with cost and scan row thresholds.
func NewQueryPlanAnalyzer(costThreshold float64, rowsThreshold int64) *QueryPlanAnalyzer {
	if costThreshold <= 0 {
		costThreshold = 1000.0
	}
	if rowsThreshold <= 0 {
		rowsThreshold = 10000
	}
	return &QueryPlanAnalyzer{
		costThreshold: costThreshold,
		rowsThreshold: rowsThreshold,
	}
}

// AnalyzeJSON parses and analyzes EXPLAIN (FORMAT JSON) bytes.
func (a *QueryPlanAnalyzer) AnalyzeJSON(data []byte) ([]QueryOptimizationFinding, error) {
	var outputs []ExplainOutput
	if err := json.Unmarshal(data, &outputs); err != nil {
		// Try single object fallback
		var single ExplainOutput
		if err2 := json.Unmarshal(data, &single); err2 != nil {
			return nil, fmt.Errorf("failed to parse explain json: %w", err)
		}
		outputs = []ExplainOutput{single}
	}

	if len(outputs) == 0 {
		return nil, fmt.Errorf("empty explain output")
	}

	var findings []QueryOptimizationFinding
	a.traverse(&outputs[0].Plan, &findings)
	return findings, nil
}

func (a *QueryPlanAnalyzer) traverse(node *PlanNode, findings *[]QueryOptimizationFinding) {
	if node == nil {
		return
	}

	// 1. Check for Sequential Scan on large relation
	if node.NodeType == "Seq Scan" && (node.PlanRows >= a.rowsThreshold || node.TotalCost >= a.costThreshold) {
		rec := "Add B-tree index on columns referenced in query filter"
		if node.Filter != "" {
			rec = fmt.Sprintf("Add index to satisfy filter condition: %s", node.Filter)
		}

		*findings = append(*findings, QueryOptimizationFinding{
			Severity:       "HIGH",
			Table:          node.RelationName,
			NodeType:       node.NodeType,
			Cost:           node.TotalCost,
			Rows:           node.PlanRows,
			Description:    fmt.Sprintf("Sequential scan on table %q with %d rows and cost %.2f", node.RelationName, node.PlanRows, node.TotalCost),
			Recommendation: rec,
		})
	}

	// 2. Check for high cost sort operations
	if strings.Contains(node.NodeType, "Sort") && node.TotalCost >= a.costThreshold {
		*findings = append(*findings, QueryOptimizationFinding{
			Severity:       "MEDIUM",
			Table:          node.RelationName,
			NodeType:       node.NodeType,
			Cost:           node.TotalCost,
			Rows:           node.PlanRows,
			Description:    fmt.Sprintf("Expensive sort operation (cost: %.2f) with %d rows", node.TotalCost, node.PlanRows),
			Recommendation: "Consider an ordered index to eliminate sorting step or increase work_mem",
		})
	}

	// Recursively inspect child plans
	for i := range node.Plans {
		a.traverse(&node.Plans[i], findings)
	}
}
