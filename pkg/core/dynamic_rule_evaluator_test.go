package core

import (
	"testing"
)

func TestForwardChainingRuleEngine_Execution(t *testing.T) {
	engine := NewForwardChainingRuleEngine(10)

	// Rule 1: if age >= 18 -> set can_vote = true (salience 10)
	_ = engine.AddRule(&ProductionRule{
		Name:     "check_voting_eligibility",
		Salience: 10,
		Condition: func(wm map[string]interface{}) bool {
			age, ok := wm["age"].(int)
			return ok && age >= 18
		},
		Action: func(wm map[string]interface{}) error {
			wm["can_vote"] = true
			return nil
		},
	})

	// Rule 2: if can_vote == true -> set status = "registered_voter" (salience 5)
	_ = engine.AddRule(&ProductionRule{
		Name:     "assign_voter_status",
		Salience: 5,
		Condition: func(wm map[string]interface{}) bool {
			canVote, ok := wm["can_vote"].(bool)
			return ok && canVote
		},
		Action: func(wm map[string]interface{}) error {
			wm["status"] = "registered_voter"
			return nil
		},
	})

	wm := map[string]interface{}{
		"age": 20,
	}

	fired, err := engine.Execute(wm)
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	if fired != 2 {
		t.Errorf("expected 2 rules to fire, got %d", fired)
	}

	if wm["can_vote"] != true {
		t.Errorf("expected can_vote=true, got %v", wm["can_vote"])
	}
	if wm["status"] != "registered_voter" {
		t.Errorf("expected status='registered_voter', got %v", wm["status"])
	}
}
