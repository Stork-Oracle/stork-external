package data_provider

import (
	"math/rand"
	"time"

	"github.com/mitchellh/mapstructure"
)

const RandomDataSourceId = "RANDOM_NUMBER"

type randomPullerConfig struct {
	UpdateFrequency string  `json:"updateFrequency"`
	MinValue        float64 `json:"minValue"`
	MaxValue        float64 `json:"maxValue"`
}

type randomPuller struct {
	valueId         ValueId
	config          randomPullerConfig
	updateFrequency time.Duration
}

func newRandomConnector(sourceConfig DataProviderSourceConfig) *randomPuller {
	var randomConfig randomPullerConfig
	mapstructure.Decode(sourceConfig.Config, &randomConfig)

	updateFrequency, err := time.ParseDuration(randomConfig.UpdateFrequency)
	if err != nil {
		panic("unable to parse update frequency: " + randomConfig.UpdateFrequency)
	}

	return &randomPuller{
		valueId:         sourceConfig.Id,
		config:          randomConfig,
		updateFrequency: updateFrequency,
	}
}

func (r *randomPuller) GetUpdate() (DataSourceUpdateMap, error) {
	randValue := r.config.MinValue + rand.Float64()*(r.config.MaxValue-r.config.MinValue)

	updateMap := DataSourceUpdateMap{
		r.valueId: DataSourceValueUpdate{
			ValueId:   r.valueId,
			Timestamp: time.Now(),
			Value:     randValue,
		},
	}

	return updateMap, nil
}

func (r *randomPuller) GetUpdateFrequency() time.Duration {
	return r.updateFrequency
}

func (r *randomPuller) GetDataSourceId() DataSourceId {
	return RandomDataSourceId
}

func getRandomDataSource(sourceConfigs []DataProviderSourceConfig) []dataSource {
	dataSources := make([]dataSource, 0)
	for _, sourceConfig := range sourceConfigs {
		connector := newRandomConnector(sourceConfig)
		dataSource := newScheduledDataSource(connector)
		dataSources = append(dataSources, dataSource)
	}
	return dataSources
}
