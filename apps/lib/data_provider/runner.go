package data_provider

import (
	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/sources"
	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/types"
)

// This is the main orchestrator of the data provider package.
// It builds the data sources defined in the config and runs them in separate goroutines.
// Each data source is responsible for fetching data from its source and reporting it to the updates channel.
// The writer then writes this data to the output address.
// If no output address is provided, the messages are logged to the console but not sent anywhere.

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
	dataSources, err := sources.BuildDataSources(r.config.Sources)
	if err != nil {
		panic("unable to build data sources: " + err.Error())
	}
	for _, dataSource := range dataSources {
		go dataSource.RunDataSource(r.updatesCh)
	}

	r.writer.Run(r.updatesCh)
}
