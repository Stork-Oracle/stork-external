package random

import (
	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/sources"
	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/types"
	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/utils"
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

func GetSourceSpecificConfig(sourceConfig types.DataProviderSourceConfig) (RandomConfig, error) {
	var config RandomConfig
	err := mapstructure.Decode(sourceConfig.Config, &config)
	return config, err
}

var _ types.DataSource = (*randomDataSource)(nil)
