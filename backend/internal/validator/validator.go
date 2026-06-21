package validator

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

type Format string

const (
	FormatJSON       Format = "json"
	FormatYAML       Format = "yaml"
	FormatProperties Format = "properties"
	FormatTOML       Format = "toml"
)

func ValidateFormat(value string, format string) error {
	switch Format(strings.ToLower(format)) {
	case FormatJSON:
		return validateJSON(value)
	case FormatYAML:
		return validateYAML(value)
	case FormatProperties:
		return validateProperties(value)
	case FormatTOML:
		return validateTOML(value)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

func validateJSON(value string) error {
	var v interface{}
	err := json.Unmarshal([]byte(value), &v)
	if err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	return nil
}

func validateYAML(value string) error {
	var v interface{}
	err := yaml.Unmarshal([]byte(value), &v)
	if err != nil {
		return fmt.Errorf("invalid YAML: %w", err)
	}
	return nil
}

func validateProperties(value string) error {
	lines := strings.Split(value, "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "!") {
			continue
		}
		if !strings.Contains(line, "=") && !strings.Contains(line, ":") {
			return fmt.Errorf("invalid properties format at line %d: missing '=' or ':'", i+1)
		}
	}
	return nil
}

func validateTOML(value string) error {
	var v interface{}
	_, err := toml.Decode(value, &v)
	if err != nil {
		return fmt.Errorf("invalid TOML: %w", err)
	}
	return nil
}

func ValidateWithSchema(value string, schema string) error {
	var schemaMap map[string]interface{}
	if err := json.Unmarshal([]byte(schema), &schemaMap); err != nil {
		return fmt.Errorf("invalid JSON Schema: %w", err)
	}

	var valueMap interface{}
	if err := json.Unmarshal([]byte(value), &valueMap); err != nil {
		return fmt.Errorf("invalid JSON value: %w", err)
	}

	return validateAgainstSchema(valueMap, schemaMap, "")
}

func validateAgainstSchema(value interface{}, schema map[string]interface{}, path string) error {
	schemaType, ok := schema["type"].(string)
	if !ok {
		return nil
	}

	switch schemaType {
	case "object":
		return validateObject(value, schema, path)
	case "array":
		return validateArray(value, schema, path)
	case "string":
		return validateString(value, schema, path)
	case "number", "integer":
		return validateNumber(value, schema, path)
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("field %s: expected boolean, got %T", path, value)
		}
	}

	return nil
}

func validateObject(value interface{}, schema map[string]interface{}, path string) error {
	obj, ok := value.(map[string]interface{})
	if !ok {
		return fmt.Errorf("field %s: expected object, got %T", path, value)
	}

	if required, ok := schema["required"].([]interface{}); ok {
		for _, r := range required {
			if key, ok := r.(string); ok {
				if _, exists := obj[key]; !exists {
					return fmt.Errorf("field %s.%s: is required", path, key)
				}
			}
		}
	}

	if properties, ok := schema["properties"].(map[string]interface{}); ok {
		for key, propSchema := range properties {
			if propVal, exists := obj[key]; exists {
				if propSchemaMap, ok := propSchema.(map[string]interface{}); ok {
					if err := validateAgainstSchema(propVal, propSchemaMap, path+"."+key); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func validateArray(value interface{}, schema map[string]interface{}, path string) error {
	arr, ok := value.([]interface{})
	if !ok {
		return fmt.Errorf("field %s: expected array, got %T", path, value)
	}

	if items, ok := schema["items"].(map[string]interface{}); ok {
		for i, item := range arr {
			if err := validateAgainstSchema(item, items, fmt.Sprintf("%s[%d]", path, i)); err != nil {
				return err
			}
		}
	}

	return nil
}

func validateString(value interface{}, schema map[string]interface{}, path string) error {
	s, ok := value.(string)
	if !ok {
		return fmt.Errorf("field %s: expected string, got %T", path, value)
	}

	if minLen, ok := schema["minLength"].(float64); ok {
		if len(s) < int(minLen) {
			return fmt.Errorf("field %s: min length is %d, got %d", path, int(minLen), len(s))
		}
	}

	if maxLen, ok := schema["maxLength"].(float64); ok {
		if len(s) > int(maxLen) {
			return fmt.Errorf("field %s: max length is %d, got %d", path, int(maxLen), len(s))
		}
	}

	return nil
}

func validateNumber(value interface{}, schema map[string]interface{}, path string) error {
	num, ok := value.(float64)
	if !ok {
		return fmt.Errorf("field %s: expected number, got %T", path, value)
	}

	if minimum, ok := schema["minimum"].(float64); ok {
		if num < minimum {
			return fmt.Errorf("field %s: minimum is %v, got %v", path, minimum, num)
		}
	}

	if maximum, ok := schema["maximum"].(float64); ok {
		if num > maximum {
			return fmt.Errorf("field %s: maximum is %v, got %v", path, maximum, num)
		}
	}

	return nil
}
