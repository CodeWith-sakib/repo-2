package core

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

// WorkflowVariableInterpolator handles complex JSON string template interpolations with dot-notation.
type WorkflowVariableInterpolator struct {
	mu sync.RWMutex
}

// NewWorkflowVariableInterpolator creates a variable interpolator.
func NewWorkflowVariableInterpolator() *WorkflowVariableInterpolator {
	return &WorkflowVariableInterpolator{}
}

// InterpolateJSONRaw parses a JSON template string, replacing dot-path parameters (${inputs.customer.id}).
func (vi *WorkflowVariableInterpolator) InterpolateJSONRaw(rawJSON []byte, context map[string]interface{}) ([]byte, error) {
	if len(rawJSON) == 0 {
		return rawJSON, nil
	}

	str := string(rawJSON)
	for {
		start := strings.Index(str, "${")
		if start == -1 {
			break
		}
		end := strings.Index(str[start:], "}")
		if end == -1 {
			break
		}
		end = start + end

		path := str[start+2 : end]
		val, exists := extractPathValue(context, path)
		if !exists {
			return nil, fmt.Errorf("unresolved context variable path: %s", path)
		}

		valStr := fmt.Sprintf("%v", val)
		str = str[:start] + valStr + str[end+1:]
	}

	// Validate resulting JSON syntax
	var dummy interface{}
	if err := json.Unmarshal([]byte(str), &dummy); err != nil {
		return nil, fmt.Errorf("interpolated JSON invalid: %w", err)
	}

	return []byte(str), nil
}

func extractPathValue(ctx map[string]interface{}, path string) (interface{}, bool) {
	parts := strings.Split(path, ".")
	var curr interface{} = ctx

	for _, part := range parts {
		m, ok := curr.(map[string]interface{})
		if !ok {
			return nil, false
		}
		v, exists := m[part]
		if !exists {
			return nil, false
		}
		curr = v
	}

	return curr, true
}
