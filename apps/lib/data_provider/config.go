package data_provider

import (
	"encoding/json"
	"fmt"
	"os"
)

func loadConfig(configPath string) (*DataProviderConfig, error) {
	configBytes, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config DataProviderConfig
	if err := json.Unmarshal(configBytes, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config file: %v", err)
	}
	return &config, nil
}
