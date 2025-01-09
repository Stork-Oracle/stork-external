package data_provider

import (
	"embed"
	"encoding/json"
	"fmt"
	"path/filepath"
)

//go:embed resources
var resourcesFS embed.FS

const ConfigDir = "resources/configs"

func LoadConfig(configName string) (*DataProviderConfig, error) {
	configBytes, err := resourcesFS.ReadFile(filepath.Join(ConfigDir, configName))
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config DataProviderConfig
	if err := json.Unmarshal(configBytes, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config file: %v", err)
	}
	return &config, nil
}
