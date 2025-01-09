package main

import (
	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider"
)

func main() {
	mainLogger := data_provider.MainLogger()
	config, err := data_provider.LoadConfig("config.json")
	if err != nil {
		mainLogger.Fatal().Err(err).Msg("could not load config file")
	}
	runner := data_provider.NewDataProviderRunner(*config)
	runner.Run()
}
