package data_provider

import (
	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/sources"
	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/types"
)

type DataProviderRunner struct {
	config    types.DataProviderConfig
	writer    Writer
	updatesCh chan types.DataSourceUpdateMap
}

func NewDataProviderRunner(dataProviderConfig types.DataProviderConfig, outputAddress string) *DataProviderRunner {
	writer, err := GetWriter(outputAddress)
	if err != nil {
		panic("unable to get data writer: " + err.Error())
	}
	return &DataProviderRunner{
		config:    dataProviderConfig,
		updatesCh: make(chan types.DataSourceUpdateMap, 4096),
		writer:    writer,
	}
}

func (r *DataProviderRunner) Run() {
	dataSources := sources.BuildDataSources(r.config.Sources)
	for _, dataSource := range dataSources {
		go dataSource.RunDataSource(r.updatesCh)
	}

	r.writer.Run(r.updatesCh)
}
