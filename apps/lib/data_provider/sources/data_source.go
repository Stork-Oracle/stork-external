package sources

import (
	"fmt"

	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider"
)

type (
	DataSourceId      string
	DataPullerFactory func(config data_provider.DataProviderSourceConfig) DataPuller
)

type DataPuller interface {
	RunContinuousPull(updatesCh chan data_provider.DataSourceUpdateMap)
	GetDataSourceId() DataSourceId
}

var pullerFactories = map[string]DataPullerFactory{}

// Register a new factory for a specific puller type.
func RegisterDataPuller(name string, factory DataPullerFactory) {
	if _, exists := pullerFactories[name]; exists {
		panic(fmt.Sprintf("DataPullerFactory already registered for: %s", name))
	}
	pullerFactories[name] = factory
}

// Get a factory by name.
func GetDataPullerFactory(name string) (DataPullerFactory, error) {
	factory, exists := pullerFactories[name]
	if !exists {
		return nil, fmt.Errorf("no factory registered for: %s", name)
	}
	return factory, nil
}
