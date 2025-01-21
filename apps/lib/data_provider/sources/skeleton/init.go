package skeleton

import (
	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/sources"
	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/types"
	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/utils"
	"github.com/mitchellh/mapstructure"
)

var SkeletonDataSourceId types.DataSourceId = types.DataSourceId(utils.GetCurrentDirName())

type skeletonDataSourceFactory struct{}

func (f *skeletonDataSourceFactory) Build(sourceConfig types.DataProviderSourceConfig) types.DataSource {
	return newSkeletonDataSource(sourceConfig)
}

func init() {
	sources.RegisterDataSourceFactory(SkeletonDataSourceId, &skeletonDataSourceFactory{})
}

// assert we're satisfying our interfaces
var (
	_ types.DataSource        = (*skeletonDataSource)(nil)
	_ types.DataSourceFactory = (*skeletonDataSourceFactory)(nil)
)

func GetSourceSpecificConfig(sourceConfig types.DataProviderSourceConfig) (SkeletonConfig, error) {
	var config SkeletonConfig
	err := mapstructure.Decode(sourceConfig.Config, &config)

	return config, err
}
