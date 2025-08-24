package core

import (
	"encoding/json"
	"testing"
)

func TestCompositeSchemaValidator_AllOf(t *testing.T) {
	minLen := 1
	s1, _ := CompileSchema([]byte(`{"type":"string","minLength":1}`))
	s2, _ := CompileSchema([]byte(`{"type":"string","maxLength":50}`))

	validator := NewCompositeSchemaValidator().AllOf(s1, s2)

	// valid: short string
	res := validator.Validate([]byte(`"hello"`))
	if !res.IsValid() {
		t.Errorf("expected valid allOf, got: %v", res.Errors)
	}

	// invalid: empty string fails minLength
	_ = minLen
	res2 := validator.Validate([]byte(`""`))
	if res2.IsValid() {
		t.Error("expected allOf violation for empty string")
	}
}

func TestCompositeSchemaValidator_AnyOf(t *testing.T) {
	sNum, _ := CompileSchema([]byte(`{"type":"number"}`))
	sStr, _ := CompileSchema([]byte(`{"type":"string"}`))

	validator := NewCompositeSchemaValidator().AnyOf(sNum, sStr)

	if res := validator.Validate([]byte(`42`)); !res.IsValid() {
		t.Errorf("number should satisfy anyOf: %v", res.Errors)
	}
	if res := validator.Validate([]byte(`"text"`)); !res.IsValid() {
		t.Errorf("string should satisfy anyOf: %v", res.Errors)
	}
	if res := validator.Validate([]byte(`true`)); res.IsValid() {
		t.Error("boolean should NOT satisfy anyOf(number|string)")
	}
}

func TestCompositeSchemaValidator_OneOf(t *testing.T) {
	sNum, _ := CompileSchema([]byte(`{"type":"number"}`))
	sStr, _ := CompileSchema([]byte(`{"type":"string"}`))

	validator := NewCompositeSchemaValidator().OneOf(sNum, sStr)

	if res := validator.Validate([]byte(`99`)); !res.IsValid() {
		t.Errorf("number should match exactly oneOf: %v", res.Errors)
	}
	// boolean matches neither
	if res := validator.Validate([]byte(`true`)); res.IsValid() {
		t.Error("boolean matches 0 schemas, oneOf should fail")
	}
}

func TestCompositeSchemaValidator_Not(t *testing.T) {
	sStr, _ := CompileSchema([]byte(`{"type":"string"}`))
	validator := NewCompositeSchemaValidator().Not(sStr)

	// number is NOT a string — should pass
	if res := validator.Validate([]byte(`42`)); !res.IsValid() {
		t.Errorf("number should pass 'not string' schema: %v", res.Errors)
	}

	// actual string should fail not
	if res := validator.Validate([]byte(`"fail"`)); res.IsValid() {
		t.Error("string should NOT pass 'not string' schema")
	}
}

func TestParseAndValidateJSON(t *testing.T) {
	schema := json.RawMessage(`{"type":"object","required":["id"],"properties":{"id":{"type":"string"}}}`)
	valid := json.RawMessage(`{"id":"abc-123"}`)
	invalid := json.RawMessage(`{"name":"no-id"}`)

	if res, err := ParseAndValidateJSON(schema, valid); err != nil || !res.IsValid() {
		t.Errorf("expected valid result, got err=%v violations=%v", err, res)
	}
	if res, err := ParseAndValidateJSON(schema, invalid); err != nil || res.IsValid() {
		t.Errorf("expected validation failure for missing id, got=%v", res)
	}
}
