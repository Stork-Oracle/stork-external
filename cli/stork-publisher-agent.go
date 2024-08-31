package main

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"time"

	storkpublisheragent "github.com/Stork-Oracle/stork_external/stork-publisher-agent"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/spf13/cobra"
)

var Hex32Regex = regexp.MustCompile(`^0x[0-9a-fA-F]+$`)

var publisherAgentCmd = &cobra.Command{
	Use:   "publisher-agent",
	Short: "Start a process to sign price updates and make them available to the Stork network",
	RunE:  runPublisherAgent,
}

// required
const SignatureTypesFlag = "signature-types"
const OracleIdFlag = "oracle-id"
const StorkAuthFlag = "stork-auth"

// optional
const EvmPrivateKeyFlag = "evm-private-key"
const EvmPublicKeyFlag = "evm-public-key"
const StarkPrivateKeyFlag = "stark-private-key"
const StarkPublicKeyFlag = "stark-public-key"
const ClockUpdatePeriodFlag = "clock-update-period"
const DeltaUpdatePeriodFlag = "delta-update-period"
const ChangeThresholdPercentFlag = "change-threshold-percent"
const IncomingWsPortFlag = "incoming-ws-port"
const StorkRegistryBaseUrlFlag = "stork-registry-base-url"
const StorkRegistryRefreshIntervalFlag = "stork-registry-refresh-interval"
const BrokerReconnectDelayFlag = "broker-reconnect-delay"

const PullBasedWsUrlFlag = "pull-based-ws-url"
const PullBasedAuthFlag = "pull-based-auth"
const PullBasedSubscriptionRequestFlag = "pull-based-subscription-request"
const PullBasedReconnectDelayFlag = "pull-based-reconnect-delay"

const SignEveryUpdateFlag = "sign-every-update"

const ProdStorkRegistryBaseUrl = "https://rest.jp.stork-oracle.network"

func init() {
	publisherAgentCmd.Flags().StringSliceP(SignatureTypesFlag, "s", nil, "Signature Types, space-separated (valid types: [evm, stark])")
	publisherAgentCmd.Flags().StringP(OracleIdFlag, "o", "", "oracle id (must be 5 characters)")
	publisherAgentCmd.Flags().StringP(StorkAuthFlag, "", "", "The auth token for Stork broker servers")

	publisherAgentCmd.Flags().StringP(EvmPrivateKeyFlag, "", "", "Your EVM private key for signing updates")
	publisherAgentCmd.Flags().StringP(EvmPublicKeyFlag, "", "", "Your EVM public key for signing updates")
	publisherAgentCmd.Flags().StringP(StarkPrivateKeyFlag, "", "", "Your Stark private key for signing updates")
	publisherAgentCmd.Flags().StringP(StarkPublicKeyFlag, "", "", "Your Stark public key for signing updates")

	publisherAgentCmd.Flags().StringP(ClockUpdatePeriodFlag, "c", "500ms", "How frequently to update the price even if it's not changing much")
	publisherAgentCmd.Flags().StringP(DeltaUpdatePeriodFlag, "d", "10ms", "How frequently to check if we're hitting the change threshold")
	publisherAgentCmd.Flags().Float64P(ChangeThresholdPercentFlag, "t", 0.1, "Report prices immediately if they've changed by more than this percentage (1 means 1%)")
	publisherAgentCmd.Flags().IntP(IncomingWsPortFlag, "i", 0, "The port which you'll report prices to")
	publisherAgentCmd.Flags().StringP(StorkRegistryBaseUrlFlag, "", ProdStorkRegistryBaseUrl, "The base URL for the Stork Registry (defaults to the production Stork Registry)")
	publisherAgentCmd.Flags().StringP(StorkRegistryRefreshIntervalFlag, "", "10m", "How frequently to refresh brokers from the Stork Registry")
	publisherAgentCmd.Flags().StringP(BrokerReconnectDelayFlag, "", "5s", "The time to wait before reconnecting to a broker websocket after a failure")

	publisherAgentCmd.Flags().StringP(PullBasedWsUrlFlag, "u", "", "A websocket URL to pull price updates from")
	publisherAgentCmd.Flags().StringP(PullBasedAuthFlag, "a", "", "A Basic auth token needed to connect to the pull websocket")
	publisherAgentCmd.Flags().StringP(PullBasedSubscriptionRequestFlag, "x", "", "A subscription message for the pull websocket")
	publisherAgentCmd.Flags().StringP(PullBasedReconnectDelayFlag, "r", "5s", "The time to wait before reconnecting to the pull websocket after a failure")

	publisherAgentCmd.Flags().BoolP(SignEveryUpdateFlag, "b", false, "Just sign every update received without any clock or delta logic")

	publisherAgentCmd.MarkFlagRequired(SignatureTypesFlag)
	publisherAgentCmd.MarkFlagRequired(OracleIdFlag)
	publisherAgentCmd.MarkFlagRequired(StorkAuthFlag)
}

func runPublisherAgent(cmd *cobra.Command, args []string) error {
	signatureTypesStr, _ := cmd.Flags().GetStringSlice(SignatureTypesFlag)
	oracleId, _ := cmd.Flags().GetString(OracleIdFlag)
	evmPrivateKey, _ := cmd.Flags().GetString(EvmPrivateKeyFlag)
	evmPublicKey, _ := cmd.Flags().GetString(EvmPublicKeyFlag)
	starkPrivateKey, _ := cmd.Flags().GetString(StarkPrivateKeyFlag)
	starkPublicKey, _ := cmd.Flags().GetString(StarkPublicKeyFlag)
	clockUpdatePeriodStr, _ := cmd.Flags().GetString(ClockUpdatePeriodFlag)
	deltaUpdatePeriodStr, _ := cmd.Flags().GetString(DeltaUpdatePeriodFlag)
	changeThresholdPercent, _ := cmd.Flags().GetFloat64(ChangeThresholdPercentFlag)
	incomingWsPort, _ := cmd.Flags().GetInt(IncomingWsPortFlag)
	storkAuth, _ := cmd.Flags().GetString(StorkAuthFlag)
	storkRegistryBaseUrl, _ := cmd.Flags().GetString(StorkRegistryBaseUrlFlag)
	storkRegistryRefreshIntervalStr, _ := cmd.Flags().GetString(StorkRegistryRefreshIntervalFlag)
	brokerReconnectDelayStr, _ := cmd.Flags().GetString(BrokerReconnectDelayFlag)

	pullBasedWsUrl, _ := cmd.Flags().GetString(PullBasedWsUrlFlag)
	pullBasedAuth, _ := cmd.Flags().GetString(PullBasedAuthFlag)
	pullBasedSubscriptionRequest, _ := cmd.Flags().GetString(PullBasedSubscriptionRequestFlag)
	pullBasedReconnectDelay, _ := cmd.Flags().GetString(PullBasedReconnectDelayFlag)

	signEveryUpdate, _ := cmd.Flags().GetBool(SignEveryUpdateFlag)

	// validate cli options
	signatureTypes := make([]storkpublisheragent.SignatureType, 0)
	for _, signatureTypeStr := range signatureTypesStr {
		signatureType := storkpublisheragent.SignatureType(signatureTypeStr)

		switch signatureType {
		case storkpublisheragent.EvmSignatureType:
			if !Hex32Regex.MatchString(evmPrivateKey) {
				return errors.New("must pass a valid EVM private key")
			}
			if !Hex32Regex.MatchString(evmPublicKey) {
				return errors.New("must pass a valid EVM public key")
			}
		case storkpublisheragent.StarkSignatureType:
			if !Hex32Regex.MatchString(starkPrivateKey) {
				return errors.New("must pass a valid Stark private key")
			}
			if !Hex32Regex.MatchString(starkPublicKey) {
				return errors.New("must pass a valid Stark public key")
			}
		default:
			return fmt.Errorf("invalid signature type: %s", signatureType)
		}

		signatureTypes = append(signatureTypes, signatureType)
	}

	if len(oracleId) != 5 {
		return errors.New("oracle id length must be 5")
	}

	clockUpdatePeriod, err := time.ParseDuration(clockUpdatePeriodStr)
	if err != nil {
		return fmt.Errorf("invalid clock update period: %s", clockUpdatePeriodStr)
	}

	deltaUpdatePeriod, err := time.ParseDuration(deltaUpdatePeriodStr)
	if err != nil {
		return fmt.Errorf("invalid delta update period: %s", deltaUpdatePeriodStr)
	}
	if deltaUpdatePeriod.Nanoseconds() == 0 {
		return errors.New("delta update period must be positive")
	}

	storkRegistryRefreshDuration, err := time.ParseDuration(storkRegistryRefreshIntervalStr)
	if err != nil {
		return fmt.Errorf("invalid stork registry refresh duration: %s", storkRegistryRefreshIntervalStr)
	}
	if storkRegistryRefreshDuration.Nanoseconds() == 0 {
		return errors.New("stork registry refresh duration must be positive")
	}

	brokerReconnectDelayDuration, err := time.ParseDuration(brokerReconnectDelayStr)
	if err != nil {
		return fmt.Errorf("invalid broker reconnect duration: %s", brokerReconnectDelayStr)
	}
	if brokerReconnectDelayDuration.Nanoseconds() == 0 {
		return errors.New("broker reconnect duration must be positive")
	}

	if changeThresholdPercent <= 0 {
		return errors.New("change threshold percent must be positive")
	}

	if incomingWsPort > 65535 {
		return errors.New("incoming ws port must be between 0 and 65535")
	}

	pullBasedReconnectDuration, err := time.ParseDuration(pullBasedReconnectDelay)
	if err != nil {
		return fmt.Errorf("invalid pull-based websocket reconnect period: %s", pullBasedReconnectDuration)
	}

	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.DurationFieldUnit = time.Nanosecond
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	mainLogger := storkpublisheragent.MainLogger()

	mainLogger.Info().Msg("initializing")

	config := storkpublisheragent.NewStorkPublisherAgentConfig(
		signatureTypes,
		storkpublisheragent.EvmPrivateKey(evmPrivateKey),
		storkpublisheragent.EvmPublisherKey(evmPublicKey),
		storkpublisheragent.StarkPrivateKey(starkPrivateKey),
		storkpublisheragent.StarkPublisherKey(starkPublicKey),
		clockUpdatePeriod,
		deltaUpdatePeriod,
		changeThresholdPercent,
		storkpublisheragent.OracleId(oracleId),
		storkRegistryBaseUrl,
		storkRegistryRefreshDuration,
		brokerReconnectDelayDuration,
		storkpublisheragent.AuthToken(storkAuth),
		pullBasedWsUrl,
		storkpublisheragent.AuthToken(pullBasedAuth),
		pullBasedSubscriptionRequest,
		pullBasedReconnectDuration,
		signEveryUpdate,
	)

	var evmRunner *storkpublisheragent.PublisherAgentRunner[*storkpublisheragent.EvmSignature]
	var starkRunner *storkpublisheragent.PublisherAgentRunner[*storkpublisheragent.StarkSignature]
	for _, signatureType := range config.SignatureTypes {
		switch signatureType {
		case storkpublisheragent.EvmSignatureType:
			evmRunner = storkpublisheragent.NewPublisherAgentRunner[*storkpublisheragent.EvmSignature](*config, storkpublisheragent.EvmSignatureType, mainLogger)
			go evmRunner.Run()
		case storkpublisheragent.StarkSignatureType:
			starkRunner = storkpublisheragent.NewPublisherAgentRunner[*storkpublisheragent.StarkSignature](*config, storkpublisheragent.StarkSignatureType, mainLogger)
			go starkRunner.Run()
		default:
			return fmt.Errorf("invalid signature type: %s", signatureType)
		}
	}

	if incomingWsPort > 0 {
		newIncomingWsHandler := func(resp http.ResponseWriter, req *http.Request) {
			if evmRunner != nil {
				evmRunner.HandleNewIncomingWsConnection(resp, req)
			}
			if starkRunner != nil {
				starkRunner.HandleNewIncomingWsConnection(resp, req)
			}
		}
		http.HandleFunc("/publish", newIncomingWsHandler)
		mainLogger.Info().Msgf("starting incoming http server on port %d", incomingWsPort)
		err = http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", incomingWsPort), nil)
		mainLogger.Fatal().Err(err).Msg("incoming http server failed, process exiting")
	} else {
		mainLogger.Info().Msg("Not running incoming http server because incoming ws port is not specified")
		select {}
	}

	return nil
}
