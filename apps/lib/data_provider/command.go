package data_provider

import (
	"fmt"
	"time"

	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/utils"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/spf13/cobra"
)

var GenerateDataProviderCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate the necessary data provider source files",
	RunE:  generateDataProvider,
}

var StartDataProviderCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a process to fetch prices from data sources",
	RunE:  runDataProvider,
}

// required
const (
	ConfigFilePathFlag   = "config-file-path"
	OutputAddressFlag    = "output-address"
	DataProviderNameFlag = "data-provider-name"
)

func init() {
	StartDataProviderCmd.Flags().StringP(ConfigFilePathFlag, "c", "", "the path of your config json file")
	StartDataProviderCmd.Flags().StringP(
		OutputAddressFlag, "o", "", "a string representing an output address (e.g. ws://localhost:5216/)",
	)
	StartDataProviderCmd.MarkFlagRequired(ConfigFilePathFlag)

	GenerateDataProviderCmd.Flags().StringP(
		DataProviderNameFlag, "n", "", "the name of your data provider in PascalCase",
	)
	GenerateDataProviderCmd.MarkFlagRequired(DataProviderNameFlag)
}

func runDataProvider(cmd *cobra.Command, args []string) error {
	configFilePath, _ := cmd.Flags().GetString(ConfigFilePathFlag)
	outputAddress, _ := cmd.Flags().GetString(OutputAddressFlag)

	mainLogger := utils.MainLogger()

	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.DurationFieldUnit = time.Nanosecond
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	mainLogger.Info().Msg("Starting data provider")

	config, err := LoadConfig(configFilePath)
	if err != nil {
		return fmt.Errorf("error loading config: %v", err)
	}

	runner := NewDataProviderRunner(*config, outputAddress)
	runner.Run()

	return nil
}
