package data_provider

import (
	"math/rand"
	"time"

	"github.com/mitchellh/mapstructure"
)

const RandomDataSourceId = "RANDOM_NUMBER"

type randomConfig struct {
	UpdateFrequency string  `json:"updateFrequency"`
	MinValue        float64 `json:"minValue"`
	MaxValue        float64 `json:"maxValue"`
}

type randomConnector struct {
	valueId         ValueId
	config          randomConfig
	updateFrequency time.Duration
}

func newRandomConnector(sourceConfig DataProviderSourceConfig) *randomConnector {
	var randomConfig randomConfig
	mapstructure.Decode(sourceConfig.Config, &randomConfig)

	updateFrequency, err := time.ParseDuration(randomConfig.UpdateFrequency)
	if err != nil {
		panic("unable to parse update frequency: " + randomConfig.UpdateFrequency)
	}

	return &randomConnector{
		valueId:         sourceConfig.Id,
		config:          randomConfig,
		updateFrequency: updateFrequency,
	}
}

func (r *randomConnector) GetUpdate() (DataSourceUpdateMap, error) {
	randValue := r.config.MinValue + rand.Float64()*(r.config.MaxValue-r.config.MinValue)

	updateMap := DataSourceUpdateMap{
		r.valueId: DataSourceValueUpdate{
			ValueId:      r.valueId,
			DataSourceId: r.GetDataSourceId(),
			Timestamp:    time.Now(),
			Value:        randValue,
		},
	}

	return updateMap, nil
}

func (r *randomConnector) GetUpdateFrequency() time.Duration {
	return r.updateFrequency
}

func (r *randomConnector) GetDataSourceId() DataSourceId {
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
