package core

import (
	"fmt"
	"regexp"
	"strings"
)

var variablePlaceholderRegex = regexp.MustCompile(`\$\{([a-zA-Z0-9_.\-]+)\}`)

// VariableSubstitutionEngine resolves parameterized expressions (${var.name}) in strings and maps.
type VariableSubstitutionEngine struct{}

// NewVariableSubstitutionEngine creates a substitution engine.
func NewVariableSubstitutionEngine() *VariableSubstitutionEngine {
	return &VariableSubstitutionEngine{}
}

// SubstituteString replaces all ${var} placeholders with values from environment map.
func (e *VariableSubstitutionEngine) SubstituteString(template string, env map[string]string) (string, error) {
	var missingVars []string

	result := variablePlaceholderRegex.ReplaceAllStringFunc(template, func(match string) string {
		varName := strings.TrimSuffix(strings.TrimPrefix(match, "${"), "}")
		if val, exists := env[varName]; exists {
			return val
		}
		missingVars = append(missingVars, varName)
		return match
	})

	if len(missingVars) > 0 {
		return result, fmt.Errorf("unresolved variables: %s", strings.Join(missingVars, ", "))
	}

	return result, nil
}

// SubstituteMap resolves variables across all string values in a map.
func (e *VariableSubstitutionEngine) SubstituteMap(input map[string]string, env map[string]string) (map[string]string, error) {
	out := make(map[string]string, len(input))
	for k, v := range input {
		resolved, err := e.SubstituteString(v, env)
		if err != nil {
			return nil, fmt.Errorf("field '%s': %w", k, err)
		}
		out[k] = resolved
	}
	return out, nil
}
