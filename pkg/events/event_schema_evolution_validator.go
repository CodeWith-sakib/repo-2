package events

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
)

var (
	ErrIncompatibleSchemaEvolution = errors.New("incompatible event schema evolution detected")
)

// SchemaFieldDef defines field type and presence requirements.
type SchemaFieldDef struct {
	Type     string `json:"type"`
	Required bool   `json:"required"`
}

// EventSchemaEvolutionValidator checks backward-compatibility of event payload versions.
type EventSchemaEvolutionValidator struct {
	mu      sync.RWMutex
	schemas map[string]map[int]map[string]SchemaFieldDef // topic -> version -> fields
}

// NewEventSchemaEvolutionValidator initializes a schema registry and evolution checker.
func NewEventSchemaEvolutionValidator() *EventSchemaEvolutionValidator {
	return &EventSchemaEvolutionValidator{
		schemas: make(map[string]map[int]map[string]SchemaFieldDef),
	}
}

// RegisterSchema registers a new schema version for a topic and validates backward compatibility.
func (v *EventSchemaEvolutionValidator) RegisterSchema(topic string, version int, fields map[string]SchemaFieldDef) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if _, ok := v.schemas[topic]; !ok {
		v.schemas[topic] = make(map[int]map[string]SchemaFieldDef)
	}

	// If version > 1, check backward compatibility against version - 1
	if prevFields, exists := v.schemas[topic][version-1]; exists {
		for fName, prevDef := range prevFields {
			currDef, existsNow := fields[fName]
			if !existsNow && prevDef.Required {
				return fmt.Errorf("%w: deleted required field %s in v%d", ErrIncompatibleSchemaEvolution, fName, version)
			}
			if existsNow && currDef.Type != prevDef.Type {
				return fmt.Errorf("%w: altered type of field %s from %s to %s in v%d", ErrIncompatibleSchemaEvolution, fName, prevDef.Type, currDef.Type, version)
			}
		}
	}

	v.schemas[topic][version] = fields
	return nil
}

// ValidatePayload checks raw json against specified topic version requirements.
func (v *EventSchemaEvolutionValidator) ValidatePayload(topic string, version int, payload []byte) error {
	v.mu.RLock()
	defer v.mu.RUnlock()

	fields, ok := v.schemas[topic][version]
	if !ok {
		return fmt.Errorf("unknown schema for topic %s v%d", topic, version)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(payload, &data); err != nil {
		return err
	}

	for fName, def := range fields {
		if def.Required {
			if _, exists := data[fName]; !exists {
				return fmt.Errorf("missing required field %s", fName)
			}
		}
	}
	return nil
}
