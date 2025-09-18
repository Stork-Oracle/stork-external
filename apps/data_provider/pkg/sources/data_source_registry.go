package sources

import (
	"fmt"

	"github.com/Stork-Oracle/stork-external/apps/data_provider/pkg/types"
	"github.com/Stork-Oracle/stork-external/apps/data_provider/pkg/utils"
)

var dataSourceFactories = map[types.DataSourceID]types.DataSourceFactory{}

// Register a new factory for a specific DataSource type.
func RegisterDataSourceFactory(dataSourceID types.DataSourceID, factory types.DataSourceFactory) {
	err := tryRegisterDataSourceFactory(dataSourceID, factory)
	if err != nil {
		panic(err)
	}
}

// exposed for testing
func tryRegisterDataSourceFactory(dataSourceID types.DataSourceID, factory types.DataSourceFactory) error {
	if _, exists := dataSourceFactories[dataSourceID]; exists {
		return fmt.Errorf("DataSourceFactory already registered for: %s", dataSourceID)
	}
	dataSourceFactories[dataSourceID] = factory
	return nil
}

// Get a factory by dataSourceID.
func GetDataSourceFactory(dataSourceID types.DataSourceID) (types.DataSourceFactory, error) {
	factory, exists := dataSourceFactories[dataSourceID]
	if !exists {
		return nil, fmt.Errorf("no factory registered for: %s", dataSourceID)
	}
	return factory, nil
}

func BuildDataSources(
	sourceConfigs []types.DataProviderSourceConfig,
) ([]types.DataSource, map[types.ValueID]any, error) {
	dataSources := make([]types.DataSource, 0)
	valueIDs := make(map[types.ValueID]any)
	for _, source := range sourceConfigs {
		_, exists := valueIDs[source.ID]
		if exists {
			return nil, nil, fmt.Errorf("duplicate value id in config: %s", source.ID)
		}
		valueIDs[source.ID] = nil

		dataSourceID, err := utils.GetDataSourceID(source.Config)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to get data source id from source config %s: %v", source.ID, err)
		}
		factory, err := GetDataSourceFactory(dataSourceID)
		if err != nil {
			return nil, nil, fmt.Errorf(
				"unable to get data source factory for data source id %s: %v",
				dataSourceID,
				err,
			)
		}
		dataSource := factory.Build(source)
		dataSources = append(dataSources, dataSource)

	}
	return dataSources, valueIDs, nil
}
