package sources

import (
	"fmt"

	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/types"
	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/utils"
)

var dataSourceFactories = map[types.DataSourceId]types.DataSourceFactory{}

// Register a new factory for a specific DataSource type.
func RegisterDataSourceFactory(dataSourceId types.DataSourceId, factory types.DataSourceFactory) {
	err := tryRegisterDataSourceFactory(dataSourceId, factory)
	if err != nil {
		panic(err)
	}
}

// exposed for testing
func tryRegisterDataSourceFactory(dataSourceId types.DataSourceId, factory types.DataSourceFactory) error {
	if _, exists := dataSourceFactories[dataSourceId]; exists {
		return fmt.Errorf("DataSourceFactory already registered for: %s", dataSourceId)
	}
	dataSourceFactories[dataSourceId] = factory
	return nil
}

// Get a factory by dataSourceId.
func GetDataSourceFactory(dataSourceId types.DataSourceId) (types.DataSourceFactory, error) {
	factory, exists := dataSourceFactories[dataSourceId]
	if !exists {
		return nil, fmt.Errorf("no factory registered for: %s", dataSourceId)
	}
	return factory, nil
}

func BuildDataSources(sourceConfigs []types.DataProviderSourceConfig) ([]types.DataSource, map[types.ValueId]any, error) {
	dataSources := make([]types.DataSource, 0)
	valueIds := make(map[types.ValueId]any)
	for _, source := range sourceConfigs {
		_, exists := valueIds[source.Id]
		if exists {
			return nil, nil, fmt.Errorf("duplicate value id in config: %s", source.Id)
		}
		valueIds[source.Id] = nil

		dataSourceId, err := utils.GetDataSourceId(source.Config)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to get data source id from source config %s: %v", source.Id, err)
		}
		factory, err := GetDataSourceFactory(dataSourceId)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to get data source factory for data source id %s: %v", dataSourceId, err)
		}
		dataSource := factory.Build(source)
		dataSources = append(dataSources, dataSource)

	}
	return dataSources, valueIds, nil
}
