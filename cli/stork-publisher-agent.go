package main

import (
	"fmt"
	"net/http"
	"time"

	storkpublisheragent "github.com/Stork-Oracle/stork_external/stork-publisher-agent"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/spf13/cobra"
)

var publisherAgentCmd = &cobra.Command{
	Use:   "publisher-agent",
	Short: "Start a process to sign price updates and make them available to the Stork network",
	RunE:  runPublisherAgent,
}

// required
const ConfigFilePathFlag = "config-file-path"
const KeysFilePathFlag = "keys-file-path"

func init() {
	publisherAgentCmd.Flags().StringP(ConfigFilePathFlag, "c", "", "the path of your config json file")
	publisherAgentCmd.Flags().StringP(KeysFilePathFlag, "k", "", "The path of your keys json file")

	publisherAgentCmd.MarkFlagRequired(ConfigFilePathFlag)
	publisherAgentCmd.MarkFlagRequired(KeysFilePathFlag)
}

func runPublisherAgent(cmd *cobra.Command, args []string) error {
	configFilePath, _ := cmd.Flags().GetString(ConfigFilePathFlag)
	keysFilePath, _ := cmd.Flags().GetString(KeysFilePathFlag)

	config, err := storkpublisheragent.LoadConfig(configFilePath, keysFilePath)
	if err != nil {
		return fmt.Errorf("error loading config: %v", err)
	}

	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.DurationFieldUnit = time.Nanosecond
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	mainLogger := storkpublisheragent.MainLogger()
	mainLogger.Info().Msg("initializing publisher agent")

	valueUpdateChannels := make([]chan storkpublisheragent.ValueUpdate, 0)
	var evmRunner *storkpublisheragent.PublisherAgentRunner[*storkpublisheragent.EvmSignature]
	var starkRunner *storkpublisheragent.PublisherAgentRunner[*storkpublisheragent.StarkSignature]
	for _, signatureType := range config.SignatureTypes {
		switch signatureType {
		case storkpublisheragent.EvmSignatureType:
			mainLogger.Info().Msg("Starting EVM runner")
			evmRunner = storkpublisheragent.NewPublisherAgentRunner[*storkpublisheragent.EvmSignature](*config, signatureType, storkpublisheragent.RunnerLogger(signatureType))
			valueUpdateChannels = append(valueUpdateChannels, evmRunner.ValueUpdateCh)
			go evmRunner.Run()
		case storkpublisheragent.StarkSignatureType:
			mainLogger.Info().Msg("Starting Stark runner")
			starkRunner = storkpublisheragent.NewPublisherAgentRunner[*storkpublisheragent.StarkSignature](*config, signatureType, storkpublisheragent.RunnerLogger(signatureType))
			valueUpdateChannels = append(valueUpdateChannels, starkRunner.ValueUpdateCh)
			go starkRunner.Run()
		default:
			return fmt.Errorf("invalid signature type: %s", signatureType)
		}
	}

	if len(config.PullBasedWsUrl) > 0 {
		incomingWsPuller := storkpublisheragent.IncomingWebsocketPuller{
			Auth:                config.PullBasedAuth,
			Url:                 config.PullBasedWsUrl,
			SubscriptionRequest: config.PullBasedWsSubscriptionRequest,
			ReconnectDelay:      config.PullBasedWsReconnectDelay,
			ValueUpdateChannels: valueUpdateChannels,
			Logger:              storkpublisheragent.IncomingLogger(),
		}
		go incomingWsPuller.Run()
	}

	if config.IncomingWsPort > 0 {
		http.HandleFunc("/publish", func(resp http.ResponseWriter, req *http.Request) {
			storkpublisheragent.HandleNewIncomingWsConnection(
				resp,
				req,
				storkpublisheragent.IncomingLogger(),
				valueUpdateChannels,
			)
		})
		mainLogger.Info().Msgf("starting incoming http server on port %d", config.IncomingWsPort)
		err = http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", config.IncomingWsPort), nil)
		mainLogger.Fatal().Err(err).Msg("incoming http server failed, process exiting")
	} else {
		mainLogger.Info().Msg("Not running incoming http server because incoming ws port is not specified")
		select {}
	}

	return nil
}
