package core

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

type SchemaType string

const (
	TypeString  SchemaType = "string"
	TypeNumber  SchemaType = "number"
	TypeInteger SchemaType = "integer"
	TypeBoolean SchemaType = "boolean"
	TypeArray   SchemaType = "array"
	TypeObject  SchemaType = "object"
	TypeNull    SchemaType = "null"
)

type JSONSchema struct {
	Type                 SchemaType             `json:"type,omitempty"`
	Title                string                 `json:"title,omitempty"`
	Description          string                 `json:"description,omitempty"`
	Required             []string               `json:"required,omitempty"`
	Properties           map[string]*JSONSchema `json:"properties,omitempty"`
	Items                *JSONSchema            `json:"items,omitempty"`
	MinLength            *int                   `json:"minLength,omitempty"`
	MaxLength            *int                   `json:"maxLength,omitempty"`
	Pattern              string                 `json:"pattern,omitempty"`
	Minimum              *float64               `json:"minimum,omitempty"`
	Maximum              *float64               `json:"maximum,omitempty"`
	Enum                 []interface{}          `json:"enum,omitempty"`
	AdditionalProperties *bool                  `json:"additionalProperties,omitempty"`
	compiledPattern      *regexp.Regexp
}

func CompileSchema(data []byte) (*JSONSchema, error) {
	var s JSONSchema
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("%w: invalid schema json: %v", ErrValidationFailed, err)
	}
	if err := s.compile(); err != nil {
		return nil, err
	}
	return &s, nil
}

func (s *JSONSchema) compile() error {
	if s.Pattern != "" {
		re, err := regexp.Compile(s.Pattern)
		if err != nil {
			return fmt.Errorf("%w: invalid regex pattern %s: %v", ErrValidationFailed, s.Pattern, err)
		}
		s.compiledPattern = re
	}
	for _, prop := range s.Properties {
		if err := prop.compile(); err != nil {
			return err
		}
	}
	if s.Items != nil {
		if err := s.Items.compile(); err != nil {
			return err
		}
	}
	return nil
}

func (s *JSONSchema) Validate(data []byte) error {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("%w: invalid JSON to validate: %v", ErrValidationFailed, err)
	}
	return s.validateValue("", v)
}

func (s *JSONSchema) validateValue(path string, val interface{}) error {
	if s.Type != "" {
		if err := s.validateType(path, val); err != nil {
			return err
		}
	}

	if len(s.Enum) > 0 {
		matched := false
		for _, e := range s.Enum {
			if fmt.Sprintf("%v", e) == fmt.Sprintf("%v", val) {
				matched = true
				break
			}
		}
		if !matched {
			return fmt.Errorf("%w: value at %s not in enum %v", ErrValidationFailed, path, s.Enum)
		}
	}

	switch s.Type {
	case TypeString:
		if str, ok := val.(string); ok {
			if s.MinLength != nil && len(str) < *s.MinLength {
				return fmt.Errorf("%w: string at %s shorter than minLength %d", ErrValidationFailed, path, *s.MinLength)
			}
			if s.MaxLength != nil && len(str) > *s.MaxLength {
				return fmt.Errorf("%w: string at %s longer than maxLength %d", ErrValidationFailed, path, *s.MaxLength)
			}
			if s.compiledPattern != nil && !s.compiledPattern.MatchString(str) {
				return fmt.Errorf("%w: string at %s does not match pattern %s", ErrValidationFailed, path, s.Pattern)
			}
		}
	case TypeNumber, TypeInteger:
		if num, ok := toFloat(val); ok {
			if s.Minimum != nil && num < *s.Minimum {
				return fmt.Errorf("%w: number at %s less than minimum %v", ErrValidationFailed, path, *s.Minimum)
			}
			if s.Maximum != nil && num > *s.Maximum {
				return fmt.Errorf("%w: number at %s greater than maximum %v", ErrValidationFailed, path, *s.Maximum)
			}
		}
	case TypeObject:
		if obj, ok := val.(map[string]interface{}); ok {
			for _, req := range s.Required {
				if _, exists := obj[req]; !exists {
					return fmt.Errorf("%w: missing required property %s%s", ErrValidationFailed, path, req)
				}
			}
			for k, v := range obj {
				subPath := path + "." + k
				if propSchema, ok := s.Properties[k]; ok {
					if err := propSchema.validateValue(subPath, v); err != nil {
						return err
					}
				} else if s.AdditionalProperties != nil && !*s.AdditionalProperties {
					return fmt.Errorf("%w: unexpected property %s", ErrValidationFailed, subPath)
				}
			}
		}
	case TypeArray:
		if arr, ok := val.([]interface{}); ok {
			if s.Items != nil {
				for i, item := range arr {
					subPath := fmt.Sprintf("%s[%d]", path, i)
					if err := s.Items.validateValue(subPath, item); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func (s *JSONSchema) validateType(path string, val interface{}) error {
	switch s.Type {
	case TypeString:
		if _, ok := val.(string); !ok {
			return fmt.Errorf("%w: expected string at %s, got %T", ErrValidationFailed, path, val)
		}
	case TypeNumber:
		if _, ok := toFloat(val); !ok {
			return fmt.Errorf("%w: expected number at %s, got %T", ErrValidationFailed, path, val)
		}
	case TypeInteger:
		if _, ok := val.(float64); !ok {
			return fmt.Errorf("%w: expected integer at %s, got %T", ErrValidationFailed, path, val)
		}
	case TypeBoolean:
		if _, ok := val.(bool); !ok {
			return fmt.Errorf("%w: expected boolean at %s, got %T", ErrValidationFailed, path, val)
		}
	case TypeArray:
		if _, ok := val.([]interface{}); !ok {
			return fmt.Errorf("%w: expected array at %s, got %T", ErrValidationFailed, path, val)
		}
	case TypeObject:
		if _, ok := val.(map[string]interface{}); !ok {
			return fmt.Errorf("%w: expected object at %s, got %T", ErrValidationFailed, path, val)
		}
	}
	return nil
}
