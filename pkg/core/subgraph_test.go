package core

import (
	"reflect"
	"testing"
)

func TestSubgraphPartitioner(t *testing.T) {
	steps := []StepDefinition{
		{ID: "A"},
		{ID: "B"},
		{ID: "C", DependsOn: []string{"A", "B"}},
		{ID: "D", DependsOn: []string{"C"}},
	}
	dag, err := BuildDAG(steps)
	if err != nil {
		t.Fatalf("build DAG error: %v", err)
	}

	part := NewSubgraphPartitioner(dag)
	tiers, err := part.ExecutionTiers()
	if err != nil {
		t.Fatalf("execution tiers error: %v", err)
	}

	expectedTiers := [][]string{
		{"A", "B"},
		{"C"},
		{"D"},
	}
	if !reflect.DeepEqual(tiers, expectedTiers) {
		t.Errorf("expected tiers %v, got %v", expectedTiers, tiers)
	}

	downstream, err := part.ReachableDownstream("A")
	if err != nil {
		t.Fatalf("downstream error: %v", err)
	}
	expectedDownstream := []string{"C", "D"}
	if !reflect.DeepEqual(downstream, expectedDownstream) {
		t.Errorf("expected downstream %v, got %v", expectedDownstream, downstream)
	}
}
