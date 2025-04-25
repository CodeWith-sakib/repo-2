package config

import (
	"encoding/json"
)

type ConfigMerger struct{}

func NewConfigMerger() *ConfigMerger {
	return &ConfigMerger{}
}

// DeepMerge merges overrides into base recursively, giving priority to overrides.
func (m *ConfigMerger) DeepMerge(base, overrides map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{})
	for k, v := range base {
		out[k] = v
	}

	for k, v := range overrides {
		if baseVal, exists := out[k]; exists {
			baseMap, baseIsMap := baseVal.(map[string]interface{})
			overMap, overIsMap := v.(map[string]interface{})
			if baseIsMap && overIsMap {
				out[k] = m.DeepMerge(baseMap, overMap)
				continue
			}
		}
		out[k] = v
	}

	return out
}

func (m *ConfigMerger) MergeJSON(baseJSON, overridesJSON []byte) ([]byte, error) {
	var baseMap, overMap map[string]interface{}
	if err := json.Unmarshal(baseJSON, &baseMap); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(overridesJSON, &overMap); err != nil {
		return nil, err
	}

	merged := m.DeepMerge(baseMap, overMap)
	return json.Marshal(merged)
}
