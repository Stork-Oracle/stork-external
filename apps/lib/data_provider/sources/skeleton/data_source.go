package skeleton

import (
	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/types"
)

type skeletonDataSource struct {
	SkeletonConfig SkeletonConfig
	// TODO: set any necessary parameters
}

func newSkeletonDataSource(sourceConfig types.DataProviderSourceConfig) *skeletonDataSource {
	skeletonConfig, err := GetSourceSpecificConfig(sourceConfig)
	if err != nil {
		panic("unable to decode config: " + err.Error())
	}

	// TODO: add any necessary initialization code
	return &skeletonDataSource{
		SkeletonConfig: skeletonConfig,
	}
}

func (r skeletonDataSource) RunDataSource(updatesCh chan types.DataSourceUpdateMap) {
	// TODO: Write all logic to fetch data points and report them to updatesCh
	panic("implement me")
}
