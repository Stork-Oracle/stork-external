package publisher_agent

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Stork-Oracle/stork-external/apps/lib/signer"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/spf13/cobra"
)

var PublisherAgentCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a process to sign price updates and make them available to the Stork network",
	RunE:  runPublisherAgent,
}

// required
const ConfigFilePathFlag = "config-file-path"
const KeysFilePathFlag = "keys-file-path"

func init() {
	PublisherAgentCmd.Flags().StringP(ConfigFilePathFlag, "c", "", "the path of your config json file")
	PublisherAgentCmd.Flags().StringP(KeysFilePathFlag, "k", "", "The path of your keys json file")

	PublisherAgentCmd.MarkFlagRequired(ConfigFilePathFlag)
	PublisherAgentCmd.MarkFlagRequired(KeysFilePathFlag)
}

func runPublisherAgent(cmd *cobra.Command, args []string) error {
	configFilePath, _ := cmd.Flags().GetString(ConfigFilePathFlag)
	keysFilePath, _ := cmd.Flags().GetString(KeysFilePathFlag)

	config, err := LoadConfig(configFilePath, keysFilePath)
	if err != nil {
		return fmt.Errorf("error loading config: %v", err)
	}

	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.DurationFieldUnit = time.Nanosecond
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	mainLogger := MainLogger()
	mainLogger.Info().Msg("initializing publisher agent")

	valueUpdateChannels := make([]chan ValueUpdate, 0)
	var evmRunner *PublisherAgentRunner[*signer.EvmSignature]
	var starkRunner *PublisherAgentRunner[*signer.StarkSignature]
	for _, signatureType := range config.SignatureTypes {
		switch signatureType {
		case EvmSignatureType:
			mainLogger.Info().Msg("Starting EVM runner")
			logger := RunnerLogger(signatureType)
			thisSigner, err := signer.NewEvmSigner(config.EvmPrivateKey, logger)
			if err != nil {
				return fmt.Errorf("failed to create EVM signer: %v", err)
			}
			evmRunner = NewPublisherAgentRunner[*signer.EvmSignature](*config, thisSigner, signatureType, logger)
			valueUpdateChannels = append(valueUpdateChannels, evmRunner.ValueUpdateCh)
			go evmRunner.Run()
		case StarkSignatureType:
			mainLogger.Info().Msg("Starting Stark runner")
			logger := RunnerLogger(signatureType)
			thisSigner, err := signer.NewStarkSigner(config.StarkPrivateKey, string(config.StarkPublicKey), string(config.OracleId), logger)
			if err != nil {
				return fmt.Errorf("failed to create EVM signer: %v", err)
			}
			starkRunner = NewPublisherAgentRunner[*signer.StarkSignature](*config, thisSigner, signatureType, logger)
			valueUpdateChannels = append(valueUpdateChannels, starkRunner.ValueUpdateCh)
			go starkRunner.Run()
		default:
			return fmt.Errorf("invalid signature type: %s", signatureType)
		}
	}

	if len(config.PullBasedWsUrl) > 0 {
		incomingWsPuller := IncomingWebsocketPuller{
			Auth:                config.PullBasedAuth,
			Url:                 config.PullBasedWsUrl,
			SubscriptionRequest: config.PullBasedWsSubscriptionRequest,
			ReconnectDelay:      config.PullBasedWsReconnectDelay,
			ValueUpdateChannels: valueUpdateChannels,
			Logger:              IncomingLogger(),
			ReadTimeout:         config.PullBasedWsReadTimeout,
		}
		go incomingWsPuller.Run()
	}

	if config.IncomingWsPort > 0 {
		http.HandleFunc("/publish", func(resp http.ResponseWriter, req *http.Request) {
			HandleNewIncomingWsConnection(
				resp,
				req,
				IncomingLogger(),
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
