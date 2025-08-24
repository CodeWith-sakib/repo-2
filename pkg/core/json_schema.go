package core

import (
	"encoding/json"
	"fmt"
)

// JSONSchemaCompositor extends the existing JSONSchema with Draft-07 combiner keywords
// (allOf, anyOf, oneOf, not) and a multi-error validator interface.

// SchemaValidationResult collects all constraint violations (non-short-circuit).
type SchemaValidationResult struct {
	Errors []SchemaViolation
}

// SchemaViolation describes a single schema constraint violation.
type SchemaViolation struct {
	Path    string
	Message string
}

func (v SchemaViolation) Error() string {
	return fmt.Sprintf("[%s] %s", v.Path, v.Message)
}

// IsValid returns true when there are no violations.
func (r *SchemaValidationResult) IsValid() bool {
	return len(r.Errors) == 0
}

// CompositeSchemaValidator validates JSON against multiple schemas with allOf / anyOf / oneOf / not.
type CompositeSchemaValidator struct {
	allOf []*JSONSchema
	anyOf []*JSONSchema
	oneOf []*JSONSchema
	not   *JSONSchema
}

// NewCompositeSchemaValidator builds a composite validator.
func NewCompositeSchemaValidator() *CompositeSchemaValidator {
	return &CompositeSchemaValidator{}
}

// AllOf adds schemas that must all pass.
func (c *CompositeSchemaValidator) AllOf(schemas ...*JSONSchema) *CompositeSchemaValidator {
	c.allOf = append(c.allOf, schemas...)
	return c
}

// AnyOf adds schemas where at least one must pass.
func (c *CompositeSchemaValidator) AnyOf(schemas ...*JSONSchema) *CompositeSchemaValidator {
	c.anyOf = append(c.anyOf, schemas...)
	return c
}

// OneOf adds schemas where exactly one must pass.
func (c *CompositeSchemaValidator) OneOf(schemas ...*JSONSchema) *CompositeSchemaValidator {
	c.oneOf = append(c.oneOf, schemas...)
	return c
}

// Not adds a schema that must NOT pass.
func (c *CompositeSchemaValidator) Not(schema *JSONSchema) *CompositeSchemaValidator {
	c.not = schema
	return c
}

// Validate executes composite validation and returns all violations.
func (c *CompositeSchemaValidator) Validate(data []byte) *SchemaValidationResult {
	result := &SchemaValidationResult{}

	// allOf: every schema must pass
	for i, s := range c.allOf {
		if err := s.Validate(data); err != nil {
			result.Errors = append(result.Errors, SchemaViolation{
				Path:    fmt.Sprintf("allOf[%d]", i),
				Message: err.Error(),
			})
		}
	}

	// anyOf: at least one must pass
	if len(c.anyOf) > 0 {
		anyPassed := false
		for _, s := range c.anyOf {
			if s.Validate(data) == nil {
				anyPassed = true
				break
			}
		}
		if !anyPassed {
			result.Errors = append(result.Errors, SchemaViolation{
				Path:    "anyOf",
				Message: "value did not satisfy any of the anyOf schemas",
			})
		}
	}

	// oneOf: exactly one must pass
	if len(c.oneOf) > 0 {
		passCount := 0
		for _, s := range c.oneOf {
			if s.Validate(data) == nil {
				passCount++
			}
		}
		if passCount != 1 {
			result.Errors = append(result.Errors, SchemaViolation{
				Path:    "oneOf",
				Message: fmt.Sprintf("value matched %d schemas, exactly 1 required", passCount),
			})
		}
	}

	// not: must NOT pass
	if c.not != nil {
		if c.not.Validate(data) == nil {
			result.Errors = append(result.Errors, SchemaViolation{
				Path:    "not",
				Message: "value unexpectedly matched the 'not' schema",
			})
		}
	}

	return result
}

// ParseAndValidateJSON is a convenience function that parses a raw JSON byte slice
// against a compiled schema and returns all violations.
func ParseAndValidateJSON(schemaRaw json.RawMessage, data json.RawMessage) (*SchemaValidationResult, error) {
	s, err := CompileSchema(schemaRaw)
	if err != nil {
		return nil, fmt.Errorf("schema compile error: %w", err)
	}
	result := &SchemaValidationResult{}
	if err := s.Validate(data); err != nil {
		result.Errors = append(result.Errors, SchemaViolation{Path: "$", Message: err.Error()})
	}
	return result, nil
}
