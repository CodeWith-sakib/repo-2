package core

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var varPattern = regexp.MustCompile(`\$\{([^}]+)\}`)

// InterpolationContext provides the dictionary of step outputs and variables.
type InterpolationContext struct {
	Variables map[string]interface{}
}

// NewInterpolationContext initializes an empty interpolation context.
func NewInterpolationContext() *InterpolationContext {
	return &InterpolationContext{
		Variables: make(map[string]interface{}),
	}
}

// Set stores a root-level variable or object in the context.
func (c *InterpolationContext) Set(key string, val interface{}) *InterpolationContext {
	c.Variables[key] = val
	return c
}

// VariableInterpolator substitutes `${path:-default}` expressions inside strings and JSON payloads.
type VariableInterpolator struct{}

// NewVariableInterpolator creates an interpolator.
func NewVariableInterpolator() *VariableInterpolator {
	return &VariableInterpolator{}
}

// InterpolateString replaces all `${path}` occurrences in a template string.
func (vi *VariableInterpolator) InterpolateString(template string, ctx *InterpolationContext) (string, error) {
	var errOut error

	result := varPattern.ReplaceAllStringFunc(template, func(match string) string {
		inner := match[2 : len(match)-1] // strip ${ and }
		val, err := vi.resolveExpression(inner, ctx)
		if err != nil {
			errOut = err
			return match
		}
		return fmt.Sprintf("%v", val)
	})

	if errOut != nil {
		return "", errOut
	}
	return result, nil
}

// InterpolateJSON recursively walks a JSON RawMessage and interpolates strings.
func (vi *VariableInterpolator) InterpolateJSON(raw json.RawMessage, ctx *InterpolationContext) (json.RawMessage, error) {
	var val interface{}
	if err := json.Unmarshal(raw, &val); err != nil {
		return nil, fmt.Errorf("invalid json: %w", err)
	}

	interpolated, err := vi.walkAndInterpolate(val, ctx)
	if err != nil {
		return nil, err
	}

	return json.Marshal(interpolated)
}

func (vi *VariableInterpolator) walkAndInterpolate(val interface{}, ctx *InterpolationContext) (interface{}, error) {
	switch v := val.(type) {
	case string:
		return vi.InterpolateString(v, ctx)
	case []interface{}:
		res := make([]interface{}, len(v))
		for i, elem := range v {
			interp, err := vi.walkAndInterpolate(elem, ctx)
			if err != nil {
				return nil, err
			}
			res[i] = interp
		}
		return res, nil
	case map[string]interface{}:
		res := make(map[string]interface{}, len(v))
		for k, elem := range v {
			interp, err := vi.walkAndInterpolate(elem, ctx)
			if err != nil {
				return nil, err
			}
			res[k] = interp
		}
		return res, nil
	default:
		return val, nil
	}
}

func (vi *VariableInterpolator) resolveExpression(expr string, ctx *InterpolationContext) (interface{}, error) {
	var defaultVal string
	hasDefault := false

	// Check for default value: path:-default
	if idx := strings.Index(expr, ":-"); idx != -1 {
		defaultVal = expr[idx+2:]
		expr = expr[:idx]
		hasDefault = true
	}

	expr = strings.TrimSpace(expr)

	val, found := vi.lookupPath(expr, ctx.Variables)
	if !found {
		if hasDefault {
			return defaultVal, nil
		}
		return nil, fmt.Errorf("variable %q not found in context", expr)
	}

	return val, nil
}

func (vi *VariableInterpolator) lookupPath(path string, data map[string]interface{}) (interface{}, bool) {
	parts := strings.Split(path, ".")
	var curr interface{} = data

	for _, part := range parts {
		// Check for array index notation: e.g. items[0]
		bracketIdx := strings.Index(part, "[")
		if bracketIdx != -1 && strings.HasSuffix(part, "]") {
			field := part[:bracketIdx]
			idxStr := part[bracketIdx+1 : len(part)-1]
			idx, err := strconv.Atoi(idxStr)
			if err != nil || idx < 0 {
				return nil, false
			}

			// First resolve map field
			m, ok := curr.(map[string]interface{})
			if !ok {
				return nil, false
			}
			arrVal, ok := m[field]
			if !ok {
				return nil, false
			}

			// Next resolve array index
			arr, ok := arrVal.([]interface{})
			if !ok || idx >= len(arr) {
				return nil, false
			}
			curr = arr[idx]
			continue
		}

		m, ok := curr.(map[string]interface{})
		if !ok {
			return nil, false
		}
		val, exists := m[part]
		if !exists {
			return nil, false
		}
		curr = val
	}

	return curr, true
}
