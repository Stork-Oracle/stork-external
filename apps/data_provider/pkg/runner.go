package data_provider

import (
	"context"

	"github.com/Stork-Oracle/stork-external/apps/data_provider/pkg/sources"
	"github.com/Stork-Oracle/stork-external/apps/data_provider/pkg/transformations"
	"github.com/Stork-Oracle/stork-external/apps/data_provider/pkg/types"
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
	outputCh  chan types.DataSourceUpdateMap
}

func NewDataProviderRunner(dataProviderConfig types.DataProviderConfig, outputAddress string) *DataProviderRunner {
	writer, err := GetWriter(outputAddress)
	if err != nil {
		panic("unable to get data writer: " + err.Error())
	}
	return &DataProviderRunner{
		config:    dataProviderConfig,
		updatesCh: make(chan types.DataSourceUpdateMap, 4096),
		outputCh:  make(chan types.DataSourceUpdateMap, 4096),
		writer:    writer,
	}
}

func (r *DataProviderRunner) Run() {
	dataSources, valueIds, err := sources.BuildDataSources(r.config.Sources)
	if err != nil {
		panic("unable to build data sources: " + err.Error())
	}

	transformationGraph, err := transformations.BuildTransformationGraph(r.config.Transformations, valueIds)
	if err != nil {
		panic("unable to build transformations: " + err.Error())
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, dataSource := range dataSources {
		go dataSource.RunDataSource(ctx, r.updatesCh)
	}

	go r.processUpdates(transformationGraph)

	r.writer.Run(r.outputCh)
}

func (r *DataProviderRunner) processUpdates(transformationGraph *transformations.TransformationGraph) {
	for {
		select {
		case sourceUpdates := <-r.updatesCh:
			processedUpdates := transformationGraph.ProcessSourceUpdates(sourceUpdates)
			r.outputCh <- processedUpdates
		}
	}
}
