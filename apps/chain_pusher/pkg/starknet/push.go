package starknet

import (
	"context"
	"os"

	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/pusher"
	"github.com/spf13/cobra"
)

func NewPushCmd() *cobra.Command {
	pushCmd := &cobra.Command{
		Use:   "starknet",
		Short: "Push WebSocket prices to Starknet contract",
		Run:   runPush,
	}

	pushCmd.Flags().StringP(pusher.StorkWebsocketEndpointFlag, "w", "", pusher.StorkWebsocketEndpointDesc)
	pushCmd.Flags().StringP(pusher.StorkAuthCredentialsFlag, "a", "", pusher.StorkAuthCredentialsDesc)
	pushCmd.Flags().StringP(pusher.ChainRpcUrlFlag, "r", "", pusher.ChainRpcUrlDesc)
	pushCmd.Flags().StringP(pusher.ChainWsUrlFlag, "c", "", pusher.ChainWsUrlDesc)
	pushCmd.Flags().StringP(pusher.ContractAddressFlag, "x", "", pusher.ContractAddressDesc)
	pushCmd.Flags().StringP(pusher.AccountAddressFlag, "n", "", pusher.AccountAddressDesc)
	pushCmd.Flags().StringP(pusher.FeeTokenAddressFlag, "t", "", pusher.FeeTokenAddressDesc)
	pushCmd.Flags().String(pusher.TipFlag, "", pusher.TipDesc)
	pushCmd.Flags().StringP(pusher.AssetConfigFileFlag, "f", "", pusher.AssetConfigFileDesc)
	pushCmd.Flags().StringP(pusher.PrivateKeyFileFlag, "k", "", pusher.PrivateKeyFileDesc)
	pushCmd.Flags().IntP(pusher.BatchingWindowFlag, "b", pusher.DefaultBatchingWindow, pusher.BatchingWindowDesc)
	pushCmd.Flags().String(pusher.BatchingWindowStrFlag, "", pusher.BatchingWindowStrDesc)
	pushCmd.Flags().IntP(pusher.PollingPeriodFlag, "p", pusher.DefaultPollingPeriod, pusher.PollingPeriodDesc)

	pushCmd.MarkFlagsMutuallyExclusive(pusher.BatchingWindowFlag, pusher.BatchingWindowStrFlag)

	_ = pushCmd.MarkFlagRequired(pusher.StorkWebsocketEndpointFlag)
	_ = pushCmd.MarkFlagRequired(pusher.StorkAuthCredentialsFlag)
	_ = pushCmd.MarkFlagRequired(pusher.ChainRpcUrlFlag)
	_ = pushCmd.MarkFlagRequired(pusher.ContractAddressFlag)
	_ = pushCmd.MarkFlagRequired(pusher.AccountAddressFlag)
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
	accountAddress, _ := cmd.Flags().GetString(pusher.AccountAddressFlag)
	feeTokenAddress, _ := cmd.Flags().GetString(pusher.FeeTokenAddressFlag)
	tip, _ := cmd.Flags().GetString(pusher.TipFlag)
	assetConfigFile, _ := cmd.Flags().GetString(pusher.AssetConfigFileFlag)
	privateKeyFile, _ := cmd.Flags().GetString(pusher.PrivateKeyFileFlag)
	batchingWindow, _ := cmd.Flags().GetInt(pusher.BatchingWindowFlag)
	batchingWindowStr, _ := cmd.Flags().GetString(pusher.BatchingWindowStrFlag)
	pollingPeriod, _ := cmd.Flags().GetInt(pusher.PollingPeriodFlag)

	logger := PusherLogger(chainRpcUrl, contractAddress)

	privateKey, err := os.ReadFile(privateKeyFile)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to read private key file")
	}

	interactor, err := NewContractInteractor(
		contractAddress,
		accountAddress,
		feeTokenAddress,
		tip,
		privateKey,
		logger,
	)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize contract interactor")
	}

	starknetPusher := pusher.NewPusher(
		storkWsEndpoint,
		storkAuth,
		chainRpcUrl,
		chainWsUrl,
		contractAddress,
		assetConfigFile,
		batchingWindowStr,
		batchingWindow,
		pollingPeriod,
		interactor,
		&logger,
	)
	starknetPusher.Run(context.Background())
}
