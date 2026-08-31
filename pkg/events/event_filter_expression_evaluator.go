package events

import (
	"strings"
	"sync"
)

// HeaderFilterOperator defines comparison logic for header attributes.
type HeaderFilterOperator string

const (
	HeaderOpEquals   HeaderFilterOperator = "EQUALS"
	HeaderOpContains HeaderFilterOperator = "CONTAINS"
	HeaderOpPrefix   HeaderFilterOperator = "PREFIX"
)

// HeaderFilterRule defines a predicate evaluated against event metadata attributes.
type HeaderFilterRule struct {
	Attribute string               `json:"attribute"`
	Operator  HeaderFilterOperator `json:"operator"`
	Value     string               `json:"value"`
}

// EventFilterExpressionEvaluator evaluates boolean predicate filter expressions over event headers.
type EventFilterExpressionEvaluator struct {
	mu sync.RWMutex
}

// NewEventFilterExpressionEvaluator creates an expression evaluator.
func NewEventFilterExpressionEvaluator() *EventFilterExpressionEvaluator {
	return &EventFilterExpressionEvaluator{}
}

// Matches returns true if metadata satisfies all rules (AND semantics).
func (e *EventFilterExpressionEvaluator) Matches(rules []HeaderFilterRule, metadata map[string]string) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()

	for _, rule := range rules {
		actual, exists := metadata[rule.Attribute]
		if !exists {
			return false
		}

		switch rule.Operator {
		case HeaderOpEquals:
			if actual != rule.Value {
				return false
			}
		case HeaderOpContains:
			if !strings.Contains(actual, rule.Value) {
				return false
			}
		case HeaderOpPrefix:
			if !strings.HasPrefix(actual, rule.Value) {
				return false
			}
		default:
			if actual != rule.Value {
				return false
			}
		}
	}

	return true
}
