package data_provider

func GetDataSourceBuilder(dataSourceId DataSourceId) func([]DataProviderSourceConfig) []dataSource {
	switch dataSourceId {
	case UniswapV2DataSourceId:
		return GetUniswapV2DataSources
	default:
		panic("unknown data source id " + dataSourceId)
	}
}
