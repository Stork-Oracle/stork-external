package random

import (
	"math/rand"
	"time"

	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider"
	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/sources"
	"github.com/mitchellh/mapstructure"
)

const RandomDataSourceId = "RANDOM_NUMBER"

type randomConfig struct {
	UpdateFrequency string  `json:"updateFrequency"`
	MinValue        float64 `json:"minValue"`
	MaxValue        float64 `json:"maxValue"`
}

type randomConnector struct {
	valueId         data_provider.ValueId
	config          randomConfig
	updateFrequency time.Duration
}

func init() {
	sources.RegisterDataPuller(RandomDataSourceId, func(sourceConfig data_provider.DataProviderSourceConfig) sources.DataPuller {
		var randomConfig randomConfig
		err := mapstructure.Decode(sourceConfig.Config, &randomConfig)
		if err != nil {
			panic("unable to decode random config: " + err.Error())
		}

		updateFrequency, err := time.ParseDuration(randomConfig.UpdateFrequency)
		if err != nil {
			panic("unable to parse update frequency: " + randomConfig.UpdateFrequency)
		}

		return &randomConnector{
			valueId:         sourceConfig.Id,
			config:          randomConfig,
			updateFrequency: updateFrequency,
		}
	})
}

func (r *randomConnector) GetDataSourceId() sources.DataSourceId {
	return RandomDataSourceId
}

func (r *randomConnector) RunContinuousPull(updatesCh chan data_provider.DataSourceUpdateMap) {
	scheduler := sources.NewScheduler(r.updateFrequency, r.getUpdate)
	scheduler.Run(updatesCh)
}

func (r *randomConnector) getUpdate() (data_provider.DataSourceUpdateMap, error) {
	randValue := r.config.MinValue + rand.Float64()*(r.config.MaxValue-r.config.MinValue)

	updateMap := data_provider.DataSourceUpdateMap{
		r.valueId: data_provider.DataSourceValueUpdate{
			ValueId:   r.valueId,
			Timestamp: time.Now(),
			Value:     randValue,
		},
	}

	return updateMap, nil
}

// Compile-time check
var _ sources.DataPuller = (*randomConnector)(nil)
