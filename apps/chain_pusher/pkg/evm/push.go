package evm

import (
	"context"
	"os"

	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/pusher"
	"github.com/spf13/cobra"
)

func NewPushCmd() *cobra.Command {
	pushCmd := &cobra.Command{
		Use:   "evm",
		Short: "Push WebSocket prices to EVM contract",
		Run:   runPush,
	}

	pushCmd.Flags().StringP(pusher.StorkWebsocketEndpointFlag, "w", "", pusher.StorkWebsocketEndpointDesc)
	pushCmd.Flags().StringP(pusher.StorkAuthCredentialsFlag, "a", "", pusher.StorkAuthCredentialsDesc)
	pushCmd.Flags().StringP(pusher.ChainRpcUrlFlag, "c", "", pusher.ChainRpcUrlDesc)
	pushCmd.Flags().StringP(pusher.ChainWsUrlFlag, "u", "", pusher.ChainWsUrlDesc)
	pushCmd.Flags().StringP(pusher.ContractAddressFlag, "x", "", pusher.ContractAddressDesc)
	pushCmd.Flags().StringP(pusher.AssetConfigFileFlag, "f", "", pusher.AssetConfigFileDesc)
	pushCmd.Flags().StringP(pusher.PrivateKeyFileFlag, "k", "", pusher.PrivateKeyFileDesc)
	pushCmd.Flags().BoolP(pusher.VerifyPublishersFlag, "v", false, pusher.VerifyPublishersDesc)
	pushCmd.Flags().IntP(pusher.BatchingWindowFlag, "b", pusher.DefaultBatchingWindow, pusher.BatchingWindowDesc)
	pushCmd.Flags().IntP(pusher.PollingPeriodFlag, "p", pusher.DefaultPollingPeriod, pusher.PollingPeriodDesc)
	pushCmd.Flags().Uint64P(pusher.GasLimitFlag, "g", 0, pusher.GasLimitDesc)

	_ = pushCmd.MarkFlagRequired(pusher.StorkWebsocketEndpointFlag)
	_ = pushCmd.MarkFlagRequired(pusher.StorkAuthCredentialsFlag)
	_ = pushCmd.MarkFlagRequired(pusher.ChainRpcUrlFlag)
	_ = pushCmd.MarkFlagRequired(pusher.ContractAddressFlag)
	_ = pushCmd.MarkFlagRequired(pusher.AssetConfigFileFlag)
	_ = pushCmd.MarkFlagRequired(pusher.PrivateKeyFileFlag)

	return pushCmd
}

func runPush(cmd *cobra.Command, args []string) {
	storkWsEndpoint, _ := cmd.Flags().GetString(pusher.StorkWebsocketEndpointFlag)
	storkAuth, _ := cmd.Flags().GetString(pusher.StorkAuthCredentialsFlag)
	chainRpcUrl, _ := cmd.Flags().GetString(pusher.ChainRpcUrlFlag)
	chainWsUrl, _ := cmd.Flags().GetString(pusher.ChainWsUrlFlag)
	contractAddress, _ := cmd.Flags().GetString(pusher.ContractAddressFlag)
	assetConfigFile, _ := cmd.Flags().GetString(pusher.AssetConfigFileFlag)
	privateKeyFile, _ := cmd.Flags().GetString(pusher.PrivateKeyFileFlag)
	verifyPublishers, _ := cmd.Flags().GetBool(pusher.VerifyPublishersFlag)
	batchingWindow, _ := cmd.Flags().GetInt(pusher.BatchingWindowFlag)
	pollingPeriod, _ := cmd.Flags().GetInt(pusher.PollingPeriodFlag)
	gasLimit, _ := cmd.Flags().GetUint64(pusher.GasLimitFlag)

	logger := PusherLogger(chainRpcUrl, contractAddress)

	keyFileContent, err := os.ReadFile(privateKeyFile)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to read private key file")
	}

	interactor, err := NewContractInteractor(
		contractAddress,
		keyFileContent,
		verifyPublishers,
		logger,
		gasLimit,
	)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize contract interactor")
	}

	pusher := pusher.NewPusher(
		storkWsEndpoint,
		storkAuth,
		chainRpcUrl,
		chainWsUrl,
		contractAddress,
		assetConfigFile,
		batchingWindow,
		pollingPeriod,
		interactor,
		&logger,
	)
	pusher.Run(context.Background())
}
