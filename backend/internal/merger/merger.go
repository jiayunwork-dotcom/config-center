package merger

import (
	"encoding/json"
	"strings"

	"gopkg.in/yaml.v3"
)

type ConfigLevel string

const (
	LevelPublic    ConfigLevel = "public"
	LevelNamespace ConfigLevel = "namespace"
	LevelGroup     ConfigLevel = "group"
)

type MergedConfig struct {
	Value  string      `json:"value"`
	Source ConfigLevel `json:"source"`
	Level  string      `json:"level"`
}

func MergeConfigs(publicConfig, namespaceConfig, groupConfig map[string]string, format string) map[string]MergedConfig {
	result := make(map[string]MergedConfig)

	for key, value := range publicConfig {
		result[key] = MergedConfig{
			Value:  value,
			Source: LevelPublic,
			Level:  "public",
		}
	}

	for key, value := range namespaceConfig {
		if existing, ok := result[key]; ok {
			merged := mergeValues(existing.Value, value, format)
			result[key] = MergedConfig{
				Value:  merged,
				Source: LevelNamespace,
				Level:  "namespace",
			}
		} else {
			result[key] = MergedConfig{
				Value:  value,
				Source: LevelNamespace,
				Level:  "namespace",
			}
		}
	}

	for key, value := range groupConfig {
		if existing, ok := result[key]; ok {
			merged := mergeValues(existing.Value, value, format)
			result[key] = MergedConfig{
				Value:  merged,
				Source: LevelGroup,
				Level:  "group",
			}
		} else {
			result[key] = MergedConfig{
				Value:  value,
				Source: LevelGroup,
				Level:  "group",
			}
		}
	}

	return result
}

func mergeValues(parent, child, format string) string {
	switch strings.ToLower(format) {
	case "json":
		return mergeJSON(parent, child)
	case "yaml":
		return mergeYAML(parent, child)
	default:
		return child
	}
}

func mergeJSON(parent, child string) string {
	var parentMap, childMap map[string]interface{}

	if err := json.Unmarshal([]byte(parent), &parentMap); err != nil {
		return child
	}
	if err := json.Unmarshal([]byte(child), &childMap); err != nil {
		return child
	}

	merged := deepMergeMap(parentMap, childMap)
	result, _ := json.MarshalIndent(merged, "", "  ")
	return string(result)
}

func mergeYAML(parent, child string) string {
	var parentMap, childMap map[string]interface{}

	if err := yaml.Unmarshal([]byte(parent), &parentMap); err != nil {
		return child
	}
	if err := yaml.Unmarshal([]byte(child), &childMap); err != nil {
		return child
	}

	merged := deepMergeMap(parentMap, childMap)
	result, _ := yaml.Marshal(merged)
	return string(result)
}

func deepMergeMap(parent, child map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for k, v := range parent {
		result[k] = v
	}

	for k, v := range child {
		if parentVal, ok := result[k]; ok {
			if parentMap, ok1 := parentVal.(map[string]interface{}); ok1 {
				if childMap, ok2 := v.(map[string]interface{}); ok2 {
					result[k] = deepMergeMap(parentMap, childMap)
					continue
				}
			}
		}
		result[k] = v
	}

	return result
}
