package data_provider

import (
	"context"
	"math"
	"time"

	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/sources"
	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/transformations"
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

	transformations, err := transformations.BuildTransformations(r.config.Transformations, valueIds)
	if err != nil {
		panic("unable to build transformations: " + err.Error())
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, dataSource := range dataSources {
		go dataSource.RunDataSource(ctx, r.updatesCh)
	}

	go r.processUpdates(dataSources, transformations)

	r.writer.Run(r.outputCh)
}

func (r *DataProviderRunner) processUpdates(dataSources []types.DataSource, transformations []transformations.OrderedTransformation) {
	currentVals := make(map[string]types.DataSourceValueUpdate, len(dataSources)+len(transformations))
	resolveVarsTicker := time.NewTicker(100 * time.Nanosecond)

	for {
		select {
		case update := <-r.updatesCh:
			r.outputCh <- update
			for valueId, update := range update {
				currentVals["s."+string(valueId)] = update
			}
		case <-resolveVarsTicker.C:
			updateMap := make(types.DataSourceUpdateMap, len(currentVals))
			for _, transformation := range transformations {
				computed := types.DataSourceValueUpdate{
					ValueId:      transformation.Id,
					DataSourceId: types.DataSourceId(transformation.Id),
					Timestamp:    time.Now(),
					Value:        transformation.Transformation.Eval(currentVals),
				}

				if math.IsNaN(computed.Value) {
					continue
				}

				// Only add to updateMap if value has changed
				if existing, ok := currentVals["t."+string(transformation.Id)]; !ok || existing.Value != computed.Value {
					updateMap[transformation.Id] = computed
				}
				currentVals["t."+string(transformation.Id)] = computed
			}
			if len(updateMap) > 0 {
				r.outputCh <- updateMap
			}
		}
	}
}
