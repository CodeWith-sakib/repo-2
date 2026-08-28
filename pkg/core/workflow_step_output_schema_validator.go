package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
)

var (
	ErrOutputSchemaMismatch = errors.New("step output failed contract schema validation")
)

// OutputContractField defines expectation for intermediate step outputs.
type OutputContractField struct {
	Type     string `json:"type"` // "string", "number", "boolean", "object", "array"
	Required bool   `json:"required"`
}

// StepOutputSchemaValidator enforces runtime data contracts between producer and consumer steps in a DAG.
type StepOutputSchemaValidator struct {
	mu        sync.RWMutex
	contracts map[string]map[string]OutputContractField // taskType -> field -> contract
}

// NewStepOutputSchemaValidator initializes a contract validator.
func NewStepOutputSchemaValidator() *StepOutputSchemaValidator {
	return &StepOutputSchemaValidator{
		contracts: make(map[string]map[string]OutputContractField),
	}
}

// RegisterContract declares the expected JSON schema for a taskType.
func (v *StepOutputSchemaValidator) RegisterContract(taskType string, fields map[string]OutputContractField) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.contracts[taskType] = fields
}

// ValidateJSON verifies that raw JSON payload matches registered contracts.
func (v *StepOutputSchemaValidator) ValidateJSON(taskType string, payload []byte) error {
	v.mu.RLock()
	defer v.mu.RUnlock()

	fields, ok := v.contracts[taskType]
	if !ok {
		return nil // no contract registered
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(payload, &parsed); err != nil {
		return fmt.Errorf("%w: invalid JSON payload: %v", ErrOutputSchemaMismatch, err)
	}

	for fieldName, contract := range fields {
		val, exists := parsed[fieldName]
		if !exists {
			if contract.Required {
				return fmt.Errorf("%w: missing required field %s", ErrOutputSchemaMismatch, fieldName)
			}
			continue
		}

		// Type validation
		switch contract.Type {
		case "string":
			if _, isStr := val.(string); !isStr {
				return fmt.Errorf("%w: field %s must be string", ErrOutputSchemaMismatch, fieldName)
			}
		case "number":
			if _, isNum := val.(float64); !isNum {
				return fmt.Errorf("%w: field %s must be number", ErrOutputSchemaMismatch, fieldName)
			}
		case "boolean":
			if _, isBool := val.(bool); !isBool {
				return fmt.Errorf("%w: field %s must be boolean", ErrOutputSchemaMismatch, fieldName)
			}
		}
	}

	return nil
}
