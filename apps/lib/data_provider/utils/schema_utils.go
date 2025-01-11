package utils

import (
	"embed"
	"fmt"

	"github.com/xeipuuv/gojsonschema"
)

func LoadSchema(schemaConfigPath string, resourcesFs embed.FS) (*gojsonschema.Schema, error) {
	schemaContent, err := resourcesFs.ReadFile(schemaConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema file for %s: %v", schemaConfigPath, err)
	}

	schemaLoader := gojsonschema.NewStringLoader(string(schemaContent))
	schema, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse schema for %s: %v", schemaConfigPath, err)
	}

	return schema, nil
}
