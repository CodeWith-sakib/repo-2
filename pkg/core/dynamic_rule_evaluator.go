package core

import (
	"fmt"
	"sort"
	"sync"
)

// RuleAction is the callback invoked when a rule fires.
type RuleAction func(workingMemory map[string]interface{}) error

// RuleCondition evaluates working memory facts to decide if a rule activates.
type RuleCondition func(workingMemory map[string]interface{}) bool

// ProductionRule represents a single condition-action production rule.
type ProductionRule struct {
	Name      string
	Salience  int // higher salience = higher firing priority
	Condition RuleCondition
	Action    RuleAction
}

// ForwardChainingRuleEngine evaluates production rules against working memory facts until saturation.
type ForwardChainingRuleEngine struct {
	mu            sync.RWMutex
	rules         []*ProductionRule
	maxIterations int
}

// NewForwardChainingRuleEngine creates a new forward-chaining rule engine.
func NewForwardChainingRuleEngine(maxIterations int) *ForwardChainingRuleEngine {
	if maxIterations <= 0 {
		maxIterations = 100
	}
	return &ForwardChainingRuleEngine{
		maxIterations: maxIterations,
	}
}

// AddRule registers a rule.
func (e *ForwardChainingRuleEngine) AddRule(rule *ProductionRule) error {
	if rule == nil || rule.Name == "" {
		return fmt.Errorf("invalid rule: nil or empty name")
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	e.rules = append(e.rules, rule)
	return nil
}

// Execute runs the match-resolve-act cycle over working memory until no rules fire or max iterations reached.
func (e *ForwardChainingRuleEngine) Execute(workingMemory map[string]interface{}) (int, error) {
	e.mu.RLock()
	// Sort rules by salience descending
	sortedRules := make([]*ProductionRule, len(e.rules))
	copy(sortedRules, e.rules)
	e.mu.RUnlock()

	sort.Slice(sortedRules, func(i, j int) bool {
		return sortedRules[i].Salience > sortedRules[j].Salience
	})

	iterations := 0
	firedCount := 0
	firedRules := make(map[string]bool)

	for iterations < e.maxIterations {
		iterations++
		ruleFiredInRound := false

		// Conflict resolution: pick first un-fired rule with satisfied condition
		for _, r := range sortedRules {
			if firedRules[r.Name] {
				continue
			}

			if r.Condition != nil && r.Condition(workingMemory) {
				// Fire rule!
				if r.Action != nil {
					if err := r.Action(workingMemory); err != nil {
						return firedCount, fmt.Errorf("action in rule %q failed: %w", r.Name, err)
					}
				}
				firedRules[r.Name] = true
				firedCount++
				ruleFiredInRound = true
				break // restart match-resolve cycle with updated working memory
			}
		}

		if !ruleFiredInRound {
			// Saturated! No more rules can fire
			break
		}
	}

	return firedCount, nil
}
