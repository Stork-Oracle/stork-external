package random

import (
	"math/rand"
	"time"

	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/sources"
	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/types"
	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/utils"
	"github.com/mitchellh/mapstructure"
	"github.com/rs/zerolog"
)

type randomDataSource struct {
	valueId         types.ValueId
	config          randomConfig
	updateFrequency time.Duration
	logger          zerolog.Logger
}

func newRandomDataSource(sourceConfig types.DataProviderSourceConfig) *randomDataSource {
	var randomConfig randomConfig
	err := mapstructure.Decode(sourceConfig.Config, &randomConfig)
	if err != nil {
		panic("unable to decode random config: " + err.Error())
	}

	updateFrequency, err := time.ParseDuration(randomConfig.UpdateFrequency)
	if err != nil {
		panic("unable to parse update frequency: " + randomConfig.UpdateFrequency)
	}

	return &randomDataSource{
		valueId:         sourceConfig.Id,
		config:          randomConfig,
		updateFrequency: updateFrequency,
		logger:          utils.DataSourceLogger(RandomDataSourceId),
	}
}

func (r randomDataSource) RunDataSource(updatesCh chan types.DataSourceUpdateMap) {
	updater := func() (types.DataSourceUpdateMap, error) { return r.getUpdate() }
	scheduler := sources.NewScheduler(
		r.updateFrequency,
		updater,
		sources.GetErrorLogHandler(r.logger, zerolog.WarnLevel),
	)
	scheduler.RunScheduler(updatesCh)
}

func (r randomDataSource) getUpdate() (types.DataSourceUpdateMap, error) {
	randValue := r.config.MinValue + rand.Float64()*(r.config.MaxValue-r.config.MinValue)

	updateMap := types.DataSourceUpdateMap{
		r.valueId: types.DataSourceValueUpdate{
			ValueId:      r.valueId,
			DataSourceId: RandomDataSourceId,
			Timestamp:    time.Now(),
			Value:        randValue,
		},
	}

	return updateMap, nil
}
