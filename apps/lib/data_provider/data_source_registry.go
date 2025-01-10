package data_provider

func GetDataSourceBuilder(dataSourceId DataSourceId) func([]DataProviderSourceConfig) []dataSource {
	switch dataSourceId {
	case UniswapV2DataSourceId:
		return getUniswapV2DataSources
	case RandomDataSourceId:
		return getRandomDataSource
	default:
		panic("unknown data source id " + dataSourceId)
	}
}
