// Code initially generated by gen.go.
// This file contains the implementation for pulling data from the data source and putting it on the updatesCh.

package monadgasfees

import (
	"context"
	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/types"
	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/utils"
	"github.com/rs/zerolog"
)

type monadGasFeesDataSource struct {
	monadGasFeesConfig MonadGasFeesConfig
	valueId 	types.ValueId
	logger 		zerolog.Logger
	// TODO: set any necessary parameters
}

func newMonadGasFeesDataSource(sourceConfig types.DataProviderSourceConfig) *monadGasFeesDataSource {
	monadGasFeesConfig, err := GetSourceSpecificConfig(sourceConfig)
	if err != nil {
		panic("unable to decode config: " + err.Error())
	}

	// TODO: add any necessary initialization code
	return &monadGasFeesDataSource{
		monadGasFeesConfig: monadGasFeesConfig,
		valueId: 	sourceConfig.Id,
		logger: 	utils.DataSourceLogger(MonadGasFeesDataSourceId),
	}
}

func (r monadGasFeesDataSource) RunDataSource(ctx context.Context, updatesCh chan types.DataSourceUpdateMap) {
	// TODO: Write all logic to fetch data points and report them to updatesCh
	panic("implement me")
}
