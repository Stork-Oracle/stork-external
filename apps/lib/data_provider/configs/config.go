package configs

import (
	"embed"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/types"
	"github.com/xeipuuv/gojsonschema"
)

//go:embed resources
var resourcesFS embed.FS

const configSchemaPath = "resources/data_provider_config.schema.json"

// exposed for testing
func LoadConfigFromBytes(configBytes []byte) (*types.DataProviderConfig, error) {
	schema, err := loadSchema(resourcesFS)
	if err != nil {
		return nil, fmt.Errorf("error loading schema: %v", err)
	}

	err = validateConfig(configBytes, schema)
	if err != nil {
		return nil, fmt.Errorf("configs file is invalid: %v", err)
	}

	var config types.DataProviderConfig
	if err := json.Unmarshal(configBytes, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configs file: %v", err)
	}
	return &config, nil
}

func loadSchema(resourcesFS embed.FS) (*gojsonschema.Schema, error) {
	schemaContent, err := resourcesFS.ReadFile(configSchemaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema file for %s: %v", configSchemaPath, err)
	}

	loader := gojsonschema.NewSchemaLoader()

	// add all source schema configs to schema loader
	sourceSchemaDir := "resources/source_config_schemas"
	sourceSchemaFiles, err := resourcesFS.ReadDir(sourceSchemaDir)
	if err != nil {
		return nil, err
	}
	for _, sourceSchemaFile := range sourceSchemaFiles {
		sourceSchemaPath := filepath.Join(sourceSchemaDir, sourceSchemaFile.Name())
		schemaBytes, err := resourcesFS.ReadFile(sourceSchemaPath)
		if err != nil {
			return nil, err
		}
		schemaFileLoader := gojsonschema.NewBytesLoader(schemaBytes)
		err = loader.AddSchema(sourceSchemaPath, schemaFileLoader)
		if err != nil {
			return nil, err
		}
	}

	topLevelSchemaLoader := gojsonschema.NewStringLoader(string(schemaContent))

	schema, err := loader.Compile(topLevelSchemaLoader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse schema for %s: %v", configSchemaPath, err)
	}

	return schema, nil
}

func validateConfig(configBytes []byte, schema *gojsonschema.Schema) error {
	var dataProviderConfig map[string]interface{}
	if err := json.Unmarshal(configBytes, &dataProviderConfig); err != nil {
		return fmt.Errorf("failed to parse configs JSON: %v", err)
	}

	configLoader := gojsonschema.NewGoLoader(dataProviderConfig)
	result, err := schema.Validate(configLoader)
	if err != nil {
		return fmt.Errorf("error validating configs: %v", err)
	}
	if !result.Valid() {
		return fmt.Errorf("configs is invalid: %v", result.Errors())
	}

	return nil
}
