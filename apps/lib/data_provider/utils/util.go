package utils

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/types"
)

func GetCurrentDirName() string {
	_, file, _, ok := runtime.Caller(1) // 1 means the caller of this function
	if !ok {
		return ""
	}
	return filepath.Base(filepath.Dir(file))
}

func GetDataSourceId(config any) (types.DataSourceId, error) {
	configMap, ok := config.(map[string]any)
	if !ok {
		return "", fmt.Errorf("config field is not interpretable as a map")
	}

	dataSourceId, exists := configMap["dataSource"]
	if !exists {
		return "", fmt.Errorf("no dataSource field in config map")
	}

	return types.DataSourceId(dataSourceId.(string)), nil
}
