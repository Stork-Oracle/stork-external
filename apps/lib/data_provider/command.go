package data_provider

import (
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/spf13/cobra"
)

var DataProviderCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a process to fetch prices from data sources",
	RunE:  runDataProvider,
}

// required
const ConfigFilePathFlag = "config-file-path"
const WebsocketUrl = "ws-url"

func init() {
	DataProviderCmd.Flags().StringP(ConfigFilePathFlag, "c", "", "the path of your config json file")
	DataProviderCmd.Flags().StringP(WebsocketUrl, "w", "", "the websocket url to write updates to")

	DataProviderCmd.MarkFlagRequired(ConfigFilePathFlag)
}

func runDataProvider(cmd *cobra.Command, args []string) error {
	configFilePath, _ := cmd.Flags().GetString(ConfigFilePathFlag)
	wsUrl, _ := cmd.Flags().GetString(WebsocketUrl)

	mainLogger := mainLogger()

	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.DurationFieldUnit = time.Nanosecond
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	mainLogger.Info().Msg("Starting data provider")

	config, err := loadConfig(configFilePath)
	if err != nil {
		return fmt.Errorf("error loading config: %v", err)
	}

	runner := NewDataProviderRunner(*config, wsUrl)
	runner.Run()

	return nil
}
