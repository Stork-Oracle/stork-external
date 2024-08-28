package main

import (
	"errors"
	"fmt"
	storkpublisheragent "github.com/Stork-Oracle/stork_external/stork-publisher-agent"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/spf13/cobra"
	"net/http"
	"regexp"
	"time"
)

var Hex32Regex = regexp.MustCompile(`^0x[0-9a-fA-F]+$`)

var publisherAgentCmd = &cobra.Command{
	Use:   "publisher-agent",
	Short: "Start a process to sign price updates and make them available to Stork",
	RunE:  runPublisherAgent,
}

// required
const SignatureTypeFlag = "signature-type"
const OracleIdFlag = "oracle-id"
const PrivateKeyFlag = "private-key"
const PublicKeyFlag = "public-key"
const StorkAuthFlag = "stork-auth"

// optional
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
	publisherAgentCmd.Flags().StringP(SignatureTypeFlag, "s", "", "Signature Type (evm or stark)")
	publisherAgentCmd.Flags().StringP(OracleIdFlag, "o", "", "oracle id (must be 5 characters)")
	publisherAgentCmd.Flags().StringP(PrivateKeyFlag, "p", "", "Your private key for signing updates")
	publisherAgentCmd.Flags().StringP(PublicKeyFlag, "k", "", "Your public key for signing updates")
	publisherAgentCmd.Flags().StringP(StorkAuthFlag, "", "", "The auth token for Stork broker servers")

	publisherAgentCmd.Flags().StringP(ClockUpdatePeriodFlag, "c", "500ms", "How frequently to update the price even if it's not changing much")
	publisherAgentCmd.Flags().StringP(DeltaUpdatePeriodFlag, "d", "10ms", "How frequently to check if we're hitting the change threshold")
	publisherAgentCmd.Flags().Float64P(ChangeThresholdPercentFlag, "t", 0.1, "Report prices immediately if they've changed by more than this percentage (1 means 1%)")
	publisherAgentCmd.Flags().IntP(IncomingWsPortFlag, "i", 5215, "The port which you'll report prices to")
	publisherAgentCmd.Flags().StringP(StorkRegistryBaseUrlFlag, "", ProdStorkRegistryBaseUrl, "The base URL for the Stork Registry (defaults to the production Stork Registry)")
	publisherAgentCmd.Flags().StringP(StorkRegistryRefreshIntervalFlag, "", "10m", "How frequently to refresh brokers from the Stork Registry")
	publisherAgentCmd.Flags().StringP(BrokerReconnectDelayFlag, "", "5s", "The time to wait before reconnecting to a broker websocket after a failure")

	publisherAgentCmd.Flags().StringP(PullBasedWsUrlFlag, "u", "", "A websocket URL to pull price updates from")
	publisherAgentCmd.Flags().StringP(PullBasedAuthFlag, "a", "", "A Basic auth token needed to connect to the pull websocket")
	publisherAgentCmd.Flags().StringP(PullBasedSubscriptionRequestFlag, "x", "", "A subscription message for the pull websocket")
	publisherAgentCmd.Flags().StringP(PullBasedReconnectDelayFlag, "r", "5s", "The time to wait before reconnecting to the pull websocket after a failure")

	publisherAgentCmd.Flags().BoolP(SignEveryUpdateFlag, "b", false, "Just sign every update received without any extra logic")

	publisherAgentCmd.MarkFlagRequired(SignatureTypeFlag)
	publisherAgentCmd.MarkFlagRequired(OracleIdFlag)
	publisherAgentCmd.MarkFlagRequired(PrivateKeyFlag)
	publisherAgentCmd.MarkFlagRequired(PublicKeyFlag)
	publisherAgentCmd.MarkFlagRequired(StorkAuthFlag)
}

func runPublisherAgent(cmd *cobra.Command, args []string) error {
	signatureTypeStr, _ := cmd.Flags().GetString(SignatureTypeFlag)
	oracleId, _ := cmd.Flags().GetString(OracleIdFlag)
	privateKey, _ := cmd.Flags().GetString(PrivateKeyFlag)
	publicKey, _ := cmd.Flags().GetString(PublicKeyFlag)
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
	signatureType := storkpublisheragent.SignatureType(signatureTypeStr)
	if !(signatureType == storkpublisheragent.EvmSignatureType || signatureType == storkpublisheragent.StarkSignatureType) {
		return fmt.Errorf("invalid signature type: %s", signatureType)
	}

	if len(oracleId) != 5 {
		return errors.New("oracle id length must be 5")
	}

	if !Hex32Regex.MatchString(privateKey) {
		return errors.New("private key must start with 0x and consist entirely of hex characters")
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

	if incomingWsPort <= 0 || incomingWsPort > 65535 {
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
		signatureType,
		storkpublisheragent.PrivateKey(privateKey),
		storkpublisheragent.PublisherKey(publicKey),
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

	switch config.SignatureType {
	case storkpublisheragent.EvmSignatureType:
		runner := *storkpublisheragent.NewPublisherAgentRunner[*storkpublisheragent.EvmSignature](*config, mainLogger)
		go runner.Run()

		http.HandleFunc("/evm/publish", runner.HandleNewIncomingWsConnection)
		mainLogger.Warn().Msgf("starting incoming http server on port %d", incomingWsPort)
		err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", incomingWsPort), nil)
		mainLogger.Fatal().Err(err).Msg("incoming http server failed, process exiting")
	case storkpublisheragent.StarkSignatureType:
		runner := *storkpublisheragent.NewPublisherAgentRunner[*storkpublisheragent.StarkSignature](*config, mainLogger)
		go runner.Run()

		http.HandleFunc("/stark/publish", runner.HandleNewIncomingWsConnection)
		mainLogger.Warn().Msgf("starting incoming http server on port %d", incomingWsPort)
		err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", incomingWsPort), nil)
		mainLogger.Fatal().Err(err).Msg("incoming http server failed, process exiting")
	default:
		return fmt.Errorf("invalid signature type: %s", signatureType)
	}

	return nil

}
