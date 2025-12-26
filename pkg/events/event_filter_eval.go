package events

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// FilterOp defines comparison operator for event property matching.
type FilterOp string

const (
	OpEquals      FilterOp = "eq"
	OpNotEquals   FilterOp = "neq"
	OpContains    FilterOp = "contains"
	OpPrefix      FilterOp = "prefix"
	OpRegexMatch  FilterOp = "regex"
	OpExists      FilterOp = "exists"
)

// EventPropertyRule specifies a predicate against an event attribute.
type EventPropertyRule struct {
	Field    string   `json:"field"`
	Op       FilterOp `json:"op"`
	Value    string   `json:"value"`
	regexVal *regexp.Regexp
}

// Compile prepares pre-compiled regular expressions for the rule.
func (r *EventPropertyRule) Compile() error {
	if r.Op == OpRegexMatch {
		compiled, err := regexp.Compile(r.Value)
		if err != nil {
			return fmt.Errorf("invalid regex '%s': %w", r.Value, err)
		}
		r.regexVal = compiled
	}
	return nil
}

// EventFilterGroup evaluates a slice of property rules using AND logic.
type EventFilterGroup struct {
	Rules []EventPropertyRule `json:"rules"`
}

// NewEventFilterGroup creates a filter group from rules.
func NewEventFilterGroup(rules []EventPropertyRule) (*EventFilterGroup, error) {
	group := &EventFilterGroup{
		Rules: make([]EventPropertyRule, len(rules)),
	}
	for i, r := range rules {
		if err := r.Compile(); err != nil {
			return nil, err
		}
		group.Rules[i] = r
	}
	return group, nil
}

// Matches checks whether an incoming JSON event matches all group rules.
func (g *EventFilterGroup) Matches(payload []byte) bool {
	if len(g.Rules) == 0 {
		return true
	}

	var data map[string]interface{}
	if err := json.Unmarshal(payload, &data); err != nil {
		return false
	}

	for _, rule := range g.Rules {
		rawVal, exists := extractNestedField(data, rule.Field)
		if rule.Op == OpExists {
			if !exists {
				return false
			}
			continue
		}

		if !exists {
			return false
		}

		strVal := fmt.Sprintf("%v", rawVal)
		switch rule.Op {
		case OpEquals:
			if strVal != rule.Value {
				return false
			}
		case OpNotEquals:
			if strVal == rule.Value {
				return false
			}
		case OpContains:
			if !strings.Contains(strVal, rule.Value) {
				return false
			}
		case OpPrefix:
			if !strings.HasPrefix(strVal, rule.Value) {
				return false
			}
		case OpRegexMatch:
			if rule.regexVal != nil && !rule.regexVal.MatchString(strVal) {
				return false
			}
		default:
			return false
		}
	}

	return true
}

func extractNestedField(data map[string]interface{}, path string) (interface{}, bool) {
	parts := strings.Split(path, ".")
	var current interface{} = data

	for _, part := range parts {
		m, ok := current.(map[string]interface{})
		if !ok {
			return nil, false
		}
		val, exists := m[part]
		if !exists {
			return nil, false
		}
		current = val
	}

	return current, true
}
