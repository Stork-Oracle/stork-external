package random

import (
	"github.com/Stork-Oracle/stork-external/apps/data_provider/lib/sources"
	"github.com/Stork-Oracle/stork-external/apps/data_provider/lib/types"
	"github.com/Stork-Oracle/stork-external/apps/data_provider/lib/utils"
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
var _ types.DataSource = (*randomDataSource)(nil)
var _ types.DataSourceFactory = (*randomDataSourceFactory)(nil)

func GetSourceSpecificConfig(sourceConfig types.DataProviderSourceConfig) (RandomConfig, error) {
	var config RandomConfig
	err := mapstructure.Decode(sourceConfig.Config, &config)
	return config, err
}
