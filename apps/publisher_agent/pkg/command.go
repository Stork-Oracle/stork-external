package publisher_agent

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Stork-Oracle/stork-external/shared"
	"github.com/Stork-Oracle/stork-external/shared/signer"
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

// not required
const KeysFilePathFlag = "keys-file-path"
const BrokerFilePathFlag = "broker-file-path" // TODO: should this be required?

func init() {
	PublisherAgentCmd.Flags().StringP(ConfigFilePathFlag, "c", "", "the path of your config json file")
	PublisherAgentCmd.Flags().StringP(KeysFilePathFlag, "k", "", "the path of your keys json file")
	PublisherAgentCmd.Flags().StringP(BrokerFilePathFlag, "b", "", "the path of your broker json file")

	PublisherAgentCmd.MarkFlagRequired(ConfigFilePathFlag)
}

func runPublisherAgent(cmd *cobra.Command, args []string) error {
	configFilePath, _ := cmd.Flags().GetString(ConfigFilePathFlag)
	keysFilePath, _ := cmd.Flags().GetString(KeysFilePathFlag)
	brokerFilePath, _ := cmd.Flags().GetString(BrokerFilePathFlag)

	config, secrets, err := LoadConfig(configFilePath, keysFilePath, brokerFilePath)
	if err != nil {
		return fmt.Errorf("error loading config: %v", err)
	}

	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.DurationFieldUnit = time.Nanosecond
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	mainLogger := MainLogger()
	mainLogger.Info().Msg("initializing publisher agent")

	valueUpdateChannels := make([]chan ValueUpdate, 0)
	var evmRunner *PublisherAgentRunner[*shared.EvmSignature]
	var starkRunner *PublisherAgentRunner[*shared.StarkSignature]
	for _, signatureType := range config.SignatureTypes {
		switch signatureType {
		case shared.EvmSignatureType:
			mainLogger.Info().Msg("Starting EVM runner")
			logger := RunnerLogger(signatureType)
			thisSigner, err := signer.NewEvmSigner(secrets.EvmPrivateKey, logger)
			if err != nil {
				return fmt.Errorf("failed to create EVM signer: %v", err)
			}
			evmAuthSigner, err := signer.NewEvmAuthSigner(secrets.EvmPrivateKey, logger)
			if err != nil {
				return fmt.Errorf("failed to create EVM auth signer: %v", err)
			}
			evmRunner = NewPublisherAgentRunner[*shared.EvmSignature](
				*config,
				thisSigner,
				evmAuthSigner,
				signatureType,
				logger,
			)
			valueUpdateChannels = append(valueUpdateChannels, evmRunner.ValueUpdateCh)
			go evmRunner.Run()
		case shared.StarkSignatureType:
			mainLogger.Info().Msg("Starting Stark runner")
			logger := RunnerLogger(signatureType)
			thisSigner, err := signer.NewStarkSigner(
				secrets.StarkPrivateKey,
				string(config.StarkPublicKey),
				string(config.OracleID),
				logger,
			)
			if err != nil {
				return fmt.Errorf("failed to create Stark signer: %v", err)
			}
			starkAuthSigner, err := signer.NewStarkAuthSigner(
				secrets.StarkPrivateKey,
				string(config.StarkPublicKey),
				logger,
			)
			if err != nil {
				return fmt.Errorf("failed to create Stark auth signer: %v", err)
			}
			starkRunner = NewPublisherAgentRunner[*shared.StarkSignature](
				*config,
				thisSigner,
				starkAuthSigner,
				signatureType,
				logger,
			)
			valueUpdateChannels = append(valueUpdateChannels, starkRunner.ValueUpdateCh)
			go starkRunner.Run()
		default:
			return fmt.Errorf("invalid signature type: %s", signatureType)
		}
	}

	if len(config.PullBasedWsUrl) > 0 {
		incomingWsPuller := IncomingWebsocketPuller{
			Auth:                secrets.PullBasedAuth,
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
