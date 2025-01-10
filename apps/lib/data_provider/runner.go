package data_provider

type DataProviderRunner struct {
	config      DataProviderConfig
	dataSources []dataSource
	writer      WebsocketWriter
	updatesCh   chan DataSourceUpdateMap
}

func NewDataProviderRunner(dataProviderConfig DataProviderConfig, wsUrl string) *DataProviderRunner {
	writer := NewWebsocketWriter(wsUrl)
	return &DataProviderRunner{
		config:    dataProviderConfig,
		updatesCh: make(chan DataSourceUpdateMap, 4096),
		writer:    *writer,
	}
}

func (r *DataProviderRunner) Run() {
	r.dataSources = buildDataSources(r.config)
	for _, dataSource := range r.dataSources {
		go dataSource.Run(r.updatesCh)
	}

	r.writer.Run(r.updatesCh)
}
