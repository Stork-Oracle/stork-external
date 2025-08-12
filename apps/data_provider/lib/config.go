package data_provider

import (
	"fmt"
	"os"

	"github.com/Stork-Oracle/stork-external/apps/data_provider/lib/configs"
	"github.com/Stork-Oracle/stork-external/apps/data_provider/lib/types"
)

func LoadConfig(configPath string) (*types.DataProviderConfig, error) {
	configBytes, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	return configs.LoadConfigFromBytes(configBytes)
}
