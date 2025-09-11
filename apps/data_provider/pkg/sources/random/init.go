package random

import (
	"github.com/Stork-Oracle/stork-external/apps/data_provider/pkg/sources"
	"github.com/Stork-Oracle/stork-external/apps/data_provider/pkg/types"
	"github.com/Stork-Oracle/stork-external/apps/data_provider/pkg/utils"
	"github.com/mitchellh/mapstructure"
)

var RandomDataSourceId types.DataSourceId = types.DataSourceId(utils.GetCurrentDirName())

type randomDataSourceFactory struct{}

func (f *randomDataSourceFactory) Build(sourceConfig types.DataProviderSourceConfig) types.DataSource {
	return newRandomDataSource(sourceConfig)
}

func init() {
	sources.RegisterDataSourceFactory(RandomDataSourceId, &randomDataSourceFactory{})
}

// assert we're satisfying our interfaces
var (
	_ types.DataSource        = (*randomDataSource)(nil)
	_ types.DataSourceFactory = (*randomDataSourceFactory)(nil)
)

func GetSourceSpecificConfig(sourceConfig types.DataProviderSourceConfig) (RandomConfig, error) {
	var config RandomConfig
	err := mapstructure.Decode(sourceConfig.Config, &config)
	return config, err
}
