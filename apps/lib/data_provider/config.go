package data_provider

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"

	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/sources"
	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/types"
	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/utils"
	"github.com/xeipuuv/gojsonschema"
)

//go:embed resources
var resourcesFS embed.FS

func loadConfig(configPath string) (*types.DataProviderConfig, error) {
	configBytes, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	err = validateConfig(configBytes)
	if err != nil {
		return nil, fmt.Errorf("config file is invalid: %v", err)
	}

	var config types.DataProviderConfig
	if err := json.Unmarshal(configBytes, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config file: %v", err)
	}
	return &config, nil
}

func validateSourceConfigs(sourceConfigsObj interface{}) error {
	sourceConfigs, ok := sourceConfigsObj.([]interface{})
	if !ok {
		return fmt.Errorf("invalid source configs type: %T", sourceConfigsObj)
	}

	for _, sourceConfig := range sourceConfigs {
		sourceConfigMap, ok := sourceConfig.(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid source config type: %v", sourceConfig)
		}
		dataSourceId := sourceConfigMap["dataSource"].(string)
		factory, err := sources.GetDataSourceFactory(types.DataSourceId(dataSourceId))
		if err != nil {
			return err
		}
		schema, err := factory.GetSchema()
		if err != nil {
			return err
		}

		sourceSpecificConfig := sourceConfigMap["config"]
		configLoader := gojsonschema.NewGoLoader(sourceSpecificConfig)
		result, err := schema.Validate(configLoader)
		if err != nil {
			return fmt.Errorf("error validating config: %v", err)
		}
		if !result.Valid() {
			return fmt.Errorf("config is invalid: %v", result.Errors())
		}
	}

	return nil
}

func validateConfig(configBytes []byte) error {
	var dataProviderConfig map[string]interface{}
	if err := json.Unmarshal(configBytes, &dataProviderConfig); err != nil {
		return fmt.Errorf("failed to parse config JSON: %v", err)
	}

	// validate top level of config
	schema, err := utils.LoadSchema("resources/config_schema.json", resourcesFS)
	if err != nil {
		return fmt.Errorf("error loading schema: %v", err)
	}

	configLoader := gojsonschema.NewGoLoader(dataProviderConfig)
	result, err := schema.Validate(configLoader)
	if err != nil {
		return fmt.Errorf("error validating config: %v", err)
	}
	if !result.Valid() {
		return fmt.Errorf("config is invalid: %v", result.Errors())
	}

	err = validateSourceConfigs(dataProviderConfig["sources"])
	if err != nil {
		return fmt.Errorf("error validating source configs: %v", err)
	}
	return nil
}
