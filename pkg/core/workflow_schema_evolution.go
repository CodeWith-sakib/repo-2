package core

import (
	"encoding/json"
	"fmt"
	"sync"
)

// SchemaCompatibilityMode defines backward/forward compatibility checks.
type SchemaCompatibilityMode string

const (
	CompatibilityBackward SchemaCompatibilityMode = "BACKWARD"
	CompatibilityForward  SchemaCompatibilityMode = "FORWARD"
	CompatibilityFull     SchemaCompatibilityMode = "FULL"
)

// SchemaFieldDefinition describes a step input/output property with type and default.
type SchemaFieldDefinition struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Required    bool        `json:"required"`
	DefaultValue interface{} `json:"default_value,omitempty"`
}

// WorkflowSchemaDefinition registers typed fields for a workflow version.
type WorkflowSchemaDefinition struct {
	Version int                              `json:"version"`
	Fields  map[string]SchemaFieldDefinition `json:"fields"`
}

// WorkflowSchemaEvolutionValidator checks schema compatibility across workflow definitions.
type WorkflowSchemaEvolutionValidator struct {
	mu      sync.RWMutex
	schemas map[int]WorkflowSchemaDefinition // version -> schema
}

// NewWorkflowSchemaEvolutionValidator creates a schema evolution validator.
func NewWorkflowSchemaEvolutionValidator() *WorkflowSchemaEvolutionValidator {
	return &WorkflowSchemaEvolutionValidator{
		schemas: make(map[int]WorkflowSchemaDefinition),
	}
}

// RegisterSchema adds a schema for a specific workflow version.
func (v *WorkflowSchemaEvolutionValidator) RegisterSchema(version int, schema WorkflowSchemaDefinition) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.schemas[version] = schema
}

// ValidateCompatibility checks if targetSchema can safely evolve from prevSchema under mode.
func (v *WorkflowSchemaEvolutionValidator) ValidateCompatibility(prevVersion, nextVersion int, mode SchemaCompatibilityMode) error {
	v.mu.RLock()
	defer v.mu.RUnlock()

	prev, existsPrev := v.schemas[prevVersion]
	next, existsNext := v.schemas[nextVersion]

	if !existsPrev || !existsNext {
		return fmt.Errorf("both versions %d and %d must be registered", prevVersion, nextVersion)
	}

	if mode == CompatibilityBackward || mode == CompatibilityFull {
		// New required fields in next without default value break backward compatibility
		for name, nextField := range next.Fields {
			if nextField.Required && nextField.DefaultValue == nil {
				if _, existsInPrev := prev.Fields[name]; !existsInPrev {
					return fmt.Errorf("field '%s' is newly required in v%d without default, breaking backward compatibility", name, nextVersion)
				}
			}
		}
	}

	if mode == CompatibilityForward || mode == CompatibilityFull {
		// Deleted required fields in next break forward compatibility
		for name, prevField := range prev.Fields {
			if prevField.Required {
				if _, existsInNext := next.Fields[name]; !existsInNext {
					return fmt.Errorf("required field '%s' in v%d was removed in v%d, breaking forward compatibility", name, prevVersion, nextVersion)
				}
			}
		}
	}

	return nil
}

// ValidatePayload verifies a raw JSON payload complies with schema definitions.
func (v *WorkflowSchemaEvolutionValidator) ValidatePayload(version int, rawJSON []byte) error {
	v.mu.RLock()
	schema, exists := v.schemas[version]
	v.mu.RUnlock()

	if !exists {
		return fmt.Errorf("schema version %d not found", version)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(rawJSON, &data); err != nil {
		return fmt.Errorf("invalid json payload: %w", err)
	}

	for name, field := range schema.Fields {
		if field.Required {
			if _, present := data[name]; !present {
				return fmt.Errorf("missing required field: %s", name)
			}
		}
	}

	return nil
}
