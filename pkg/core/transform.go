package core

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type TransformationRule struct {
	SourcePath string `json:"source_path"`
	TargetPath string `json:"target_path"`
	DefaultVal string `json:"default_val,omitempty"`
}

type PayloadTransformer struct {
	rules []TransformationRule
}

func NewPayloadTransformer(rules []TransformationRule) *PayloadTransformer {
	return &PayloadTransformer{rules: rules}
}

func ExtractJSONPath(data interface{}, path string) (interface{}, error) {
	if path == "" || path == "$" {
		return data, nil
	}

	trimmed := strings.TrimPrefix(path, "$.")
	trimmed = strings.TrimPrefix(trimmed, "$")
	segments := splitPathSegments(trimmed)

	var current = data
	for _, seg := range segments {
		if seg == "" {
			continue
		}

		if strings.HasSuffix(seg, "]") && strings.Contains(seg, "[") {
			parts := strings.Split(seg, "[")
			prop := parts[0]
			idxStr := strings.TrimSuffix(parts[1], "]")
			idx, err := strconv.Atoi(idxStr)
			if err != nil {
				return nil, fmt.Errorf("invalid array index %s: %w", idxStr, err)
			}

			if prop != "" {
				m, ok := current.(map[string]interface{})
				if !ok {
					return nil, fmt.Errorf("expected object at %s", prop)
				}
				current = m[prop]
			}

			arr, ok := current.([]interface{})
			if !ok {
				return nil, fmt.Errorf("expected array at %s", seg)
			}
			if idx < 0 || idx >= len(arr) {
				return nil, fmt.Errorf("array index out of bounds: %d", idx)
			}
			current = arr[idx]
		} else {
			m, ok := current.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("expected object at segment %s", seg)
			}
			val, exists := m[seg]
			if !exists {
				return nil, fmt.Errorf("key not found: %s", seg)
			}
			current = val
		}
	}

	return current, nil
}

func (pt *PayloadTransformer) Transform(sourceJSON []byte) ([]byte, error) {
	var src interface{}
	if err := json.Unmarshal(sourceJSON, &src); err != nil {
		return nil, fmt.Errorf("%w: invalid source json: %v", ErrValidationFailed, err)
	}

	dest := make(map[string]interface{})
	for _, rule := range pt.rules {
		val, err := ExtractJSONPath(src, rule.SourcePath)
		if err != nil {
			if rule.DefaultVal != "" {
				val = rule.DefaultVal
			} else {
				continue
			}
		}

		setNestedValue(dest, rule.TargetPath, val)
	}

	return json.Marshal(dest)
}

func splitPathSegments(path string) []string {
	return strings.Split(path, ".")
}

func setNestedValue(m map[string]interface{}, path string, val interface{}) {
	trimmed := strings.TrimPrefix(path, "$.")
	trimmed = strings.TrimPrefix(trimmed, "$")
	segments := splitPathSegments(trimmed)

	var current = m
	for i, seg := range segments {
		if i == len(segments)-1 {
			current[seg] = val
			return
		}

		if _, exists := current[seg]; !exists {
			current[seg] = make(map[string]interface{})
		}
		if next, ok := current[seg].(map[string]interface{}); ok {
			current = next
		} else {
			newMap := make(map[string]interface{})
			current[seg] = newMap
			current = newMap
		}
	}
}
