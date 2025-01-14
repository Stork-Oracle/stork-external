package sources

import (
	"fmt"

	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/types"
	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/utils"
)

var dataSourceFactories = map[types.DataSourceId]types.DataSourceFactory{}

// Register a new factory for a specific DataSource type.
func RegisterDataSourceFactory(dataSourceId types.DataSourceId, factory types.DataSourceFactory) {
	if _, exists := dataSourceFactories[dataSourceId]; exists {
		panic(fmt.Sprintf("DataSourceFactory already registered for: %s", dataSourceId))
	}
	dataSourceFactories[dataSourceId] = factory
}

// Get a factory by dataSourceId.
func GetDataSourceFactory(dataSourceId types.DataSourceId) (types.DataSourceFactory, error) {
	factory, exists := dataSourceFactories[dataSourceId]
	if !exists {
		return nil, fmt.Errorf("no factory registered for: %s", dataSourceId)
	}
	return factory, nil
}

func BuildDataSources(sourceConfigs []types.DataProviderSourceConfig) []types.DataSource {
	dataSources := make([]types.DataSource, 0)
	for _, source := range sourceConfigs {
		dataSourceId, err := utils.GetDataSourceId(source.Config)
		if err != nil {
			panic("unable to get data source id from source config " + string(source.Id) + ": " + err.Error())
		}
		factory, err := GetDataSourceFactory(dataSourceId)
		if err != nil {
			panic(err)
		}
		dataSource := factory.Build(source)
		dataSources = append(dataSources, dataSource)
	}
	return dataSources
}
