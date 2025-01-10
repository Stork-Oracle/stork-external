package data_provider

type dataSource interface {
	// Add all value updates to updatesCh
	Run(updatesCh chan DataSourceUpdateMap)
	GetDataSourceId() DataSourceId
}

func buildDataSources(config DataProviderConfig) []dataSource {
	// group by data source id to support batched feeds
	sourceConfigsByDataSource := make(map[DataSourceId][]DataProviderSourceConfig)
	for _, sourceConfig := range config.Sources {
		dataSourceId := sourceConfig.DataSourceId
		if _, ok := sourceConfigsByDataSource[dataSourceId]; !ok {
			sourceConfigsByDataSource[dataSourceId] = make([]DataProviderSourceConfig, 0)
		}
		sourceConfigsByDataSource[dataSourceId] = append(sourceConfigsByDataSource[dataSourceId], sourceConfig)

	}

	// initialize data sources
	allDataSources := make([]dataSource, 0)
	for dataSourceId, sourceConfigs := range sourceConfigsByDataSource {
		dataSourceBuilder := GetDataSourceBuilder(dataSourceId)
		dataSources := dataSourceBuilder(sourceConfigs)

		allDataSources = append(allDataSources, dataSources...)
	}

	return allDataSources
}
