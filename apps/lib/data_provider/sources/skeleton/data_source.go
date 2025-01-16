package skeleton

import (
	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/types"
)

type skeletonDataSource struct {
	// TODO: set any necessary parameters
}

func newSkeletonDataSource(sourceConfig types.DataProviderSourceConfig) *skeletonDataSource {
	skeletonConfig, err := GetSourceSpecificConfig(sourceConfig)
	if err != nil {
		panic("unable to decode config: " + err.Error())
	}

	// TODO: add any necessary initialization code
	panic("implement me")
}

func (r skeletonDataSource) RunDataSource(updatesCh chan types.DataSourceUpdateMap) {
	// TODO: Write all logic to fetch datapoints and add to updatesCh
	panic("implement me")
}
