package data_provider

import (
	"fmt"
	"os"

	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/configs"
	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/types"
)

func LoadConfig(configPath string) (*types.DataProviderConfig, error) {
	configBytes, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	return configs.LoadConfigFromBytes(configBytes)
}
