package core

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

var macroRegex = regexp.MustCompile(`\{\{\s*([a-zA-Z0-9_\.\-]+)\s*\}\}`)

// TemplateMacroExpander expands dynamic macro placeholders in workflow task payloads.
type TemplateMacroExpander struct {
	strict bool
}

func NewTemplateMacroExpander(strict bool) *TemplateMacroExpander {
	return &TemplateMacroExpander{strict: strict}
}

func (e *TemplateMacroExpander) ExpandString(template string, context map[string]interface{}) (string, error) {
	var firstErr error
	res := macroRegex.ReplaceAllStringFunc(template, func(match string) string {
		sub := macroRegex.FindStringSubmatch(match)
		if len(sub) < 2 {
			return match
		}
		path := sub[1]
		val, found := resolveDottedPath(context, path)
		if !found {
			if e.strict {
				firstErr = fmt.Errorf("missing macro context key: %s", path)
			}
			return match
		}
		return fmt.Sprintf("%v", val)
	})
	return res, firstErr
}

func (e *TemplateMacroExpander) ExpandPayload(payload []byte, context map[string]interface{}) ([]byte, error) {
	if len(payload) == 0 {
		return payload, nil
	}
	expandedStr, err := e.ExpandString(string(payload), context)
	if err != nil {
		return nil, err
	}
	var js json.RawMessage
	if json.Unmarshal(payload, &js) == nil {
		var js2 json.RawMessage
		if err := json.Unmarshal([]byte(expandedStr), &js2); err != nil {
			return nil, fmt.Errorf("expanded payload is invalid JSON: %w", err)
		}
	}
	return []byte(expandedStr), nil
}

func resolveDottedPath(context map[string]interface{}, path string) (interface{}, bool) {
	parts := strings.Split(path, ".")
	var curr interface{} = context
	for _, part := range parts {
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
