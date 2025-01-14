package utils

import (
	"embed"
	"encoding/json"
	"fmt"

	"github.com/xeipuuv/gojsonschema"
)

// this class is only used for

const configSchemaPath = "resources/config_schema.json"

func LoadSchema(resourcesFS embed.FS) (*gojsonschema.Schema, error) {
	schemaContent, err := resourcesFS.ReadFile(configSchemaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema file for %s: %v", configSchemaPath, err)
	}

	schemaLoader := gojsonschema.NewStringLoader(string(schemaContent))
	schema, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse schema for %s: %v", configSchemaPath, err)
	}

	return schema, nil
}

func ValidateConfig(configBytes []byte, schema *gojsonschema.Schema) error {
	var dataProviderConfig map[string]interface{}
	if err := json.Unmarshal(configBytes, &dataProviderConfig); err != nil {
		return fmt.Errorf("failed to parse config JSON: %v", err)
	}

	configLoader := gojsonschema.NewGoLoader(dataProviderConfig)
	result, err := schema.Validate(configLoader)
	if err != nil {
		return fmt.Errorf("error validating config: %v", err)
	}
	if !result.Valid() {
		return fmt.Errorf("config is invalid: %v", result.Errors())
	}

	return nil
}
