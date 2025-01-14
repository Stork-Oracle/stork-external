package data_provider

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"

	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/types"
	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/utils"
)

//go:embed resources
var resourcesFS embed.FS

func loadConfig(configPath string) (*types.DataProviderConfig, error) {
	configBytes, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	// validate top level of config
	schema, err := utils.LoadSchema(resourcesFS)
	if err != nil {
		return nil, fmt.Errorf("error loading schema: %v", err)
	}

	err = utils.ValidateConfig(configBytes, schema)
	if err != nil {
		return nil, fmt.Errorf("config file is invalid: %v", err)
	}

	var config types.DataProviderConfig
	if err := json.Unmarshal(configBytes, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config file: %v", err)
	}
	return &config, nil
}
