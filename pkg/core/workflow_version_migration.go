package core

import (
	"errors"
	"fmt"
	"sync"
)

// StepMigrationRule specifies mutation rules when moving step configs between workflow schema versions.
type StepMigrationRule struct {
	SourceStepID string
	TargetStepID string
	FieldRenames map[string]string
	DefaultValues map[string]interface{}
}

// WorkflowVersionMigrator transforms workflow runs from legacy definition schemas to newer versions.
type WorkflowVersionMigrator struct {
	mu    sync.RWMutex
	rules map[int]map[int][]StepMigrationRule // fromVersion -> toVersion -> rules
}

// NewWorkflowVersionMigrator creates a version migrator instance.
func NewWorkflowVersionMigrator() *WorkflowVersionMigrator {
	return &WorkflowVersionMigrator{
		rules: make(map[int]map[int][]StepMigrationRule),
	}
}

// RegisterMigration defines a migration path between two sequential versions.
func (m *WorkflowVersionMigrator) RegisterMigration(fromVersion, toVersion int, rules []StepMigrationRule) error {
	if fromVersion >= toVersion {
		return errors.New("fromVersion must be strictly less than toVersion")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.rules[fromVersion]; !exists {
		m.rules[fromVersion] = make(map[int][]StepMigrationRule)
	}
	m.rules[fromVersion][toVersion] = rules
	return nil
}

// MigrateStepConfig applies translation rules to migrate step inputs/configuration.
func (m *WorkflowVersionMigrator) MigrateStepConfig(fromVersion, toVersion int, stepID string, config map[string]interface{}) (map[string]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	targetRules, exists := m.rules[fromVersion][toVersion]
	if !exists {
		return nil, fmt.Errorf("no registered migration from version %d to %d", fromVersion, toVersion)
	}

	migrated := make(map[string]interface{}, len(config))
	for k, v := range config {
		migrated[k] = v
	}

	for _, rule := range targetRules {
		if rule.SourceStepID == stepID || rule.SourceStepID == "*" {
			for oldKey, newKey := range rule.FieldRenames {
				if val, found := migrated[oldKey]; found {
					delete(migrated, oldKey)
					migrated[newKey] = val
				}
			}
			for defKey, defVal := range rule.DefaultValues {
				if _, found := migrated[defKey]; !found {
					migrated[defKey] = defVal
				}
			}
		}
	}

	return migrated, nil
}
