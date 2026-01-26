package solana

import (
	"context"

	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/pusher"
	"github.com/gagliardetto/solana-go"
	"github.com/spf13/cobra"
)

const (
	DefaultLimitPerSecond = 40
	DefaultBurstLimit     = 10
	DefaultBatchSize      = 4
)

func NewPushCmd() *cobra.Command {
	pushCmd := &cobra.Command{
		Use:   "solana",
		Short: "Push WebSocket prices to Solana contract",
		Run:   runSolanaPush,
	}

	pushCmd.Flags().StringP(pusher.StorkWebsocketEndpointFlag, "w", "", pusher.StorkWebsocketEndpointDesc)
	pushCmd.Flags().StringP(pusher.StorkAuthCredentialsFlag, "a", "", pusher.StorkAuthCredentialsDesc)
	pushCmd.Flags().StringP(pusher.ChainRpcUrlFlag, "c", "", pusher.ChainRpcUrlDesc)
	pushCmd.Flags().StringP(pusher.ChainWsUrlFlag, "u", "", pusher.ChainWsUrlDesc)
	pushCmd.Flags().StringP(pusher.ContractAddressFlag, "x", "", pusher.ContractAddressDesc)
	pushCmd.Flags().StringP(pusher.AssetConfigFileFlag, "f", "", pusher.AssetConfigFileDesc)
	pushCmd.Flags().StringP(pusher.PrivateKeyFileFlag, "k", "", pusher.PrivateKeyFileDesc)
	pushCmd.Flags().StringP(pusher.BatchingWindowFlag, "b", pusher.DefaultBatchingWindow, pusher.BatchingWindowDesc)
	pushCmd.Flags().IntP(pusher.PollingPeriodFlag, "p", pusher.DefaultPollingPeriod, pusher.PollingPeriodDesc)
	pushCmd.Flags().IntP(pusher.LimitPerSecondFlag, "l", DefaultLimitPerSecond, pusher.LimitPerSecondDesc)
	pushCmd.Flags().IntP(pusher.BurstLimitFlag, "r", DefaultBurstLimit, pusher.BurstLimitDesc)
	pushCmd.Flags().IntP(pusher.BatchSizeFlag, "s", DefaultBatchSize, pusher.BatchSizeDesc)

	_ = pushCmd.MarkFlagRequired(pusher.StorkWebsocketEndpointFlag)
	_ = pushCmd.MarkFlagRequired(pusher.StorkAuthCredentialsFlag)
	_ = pushCmd.MarkFlagRequired(pusher.ChainRpcUrlFlag)
	_ = pushCmd.MarkFlagRequired(pusher.ContractAddressFlag)
	_ = pushCmd.MarkFlagRequired(pusher.AssetConfigFileFlag)
	_ = pushCmd.MarkFlagRequired(pusher.MnemonicFileFlag)

	return pushCmd
}

func runSolanaPush(cmd *cobra.Command, args []string) {
	storkWsEndpoint, _ := cmd.Flags().GetString(pusher.StorkWebsocketEndpointFlag)
	storkAuth, _ := cmd.Flags().GetString(pusher.StorkAuthCredentialsFlag)
	chainRpcUrl, _ := cmd.Flags().GetString(pusher.ChainRpcUrlFlag)
	chainWsUrl, _ := cmd.Flags().GetString(pusher.ChainWsUrlFlag)
	contractAddress, _ := cmd.Flags().GetString(pusher.ContractAddressFlag)
	assetConfigFile, _ := cmd.Flags().GetString(pusher.AssetConfigFileFlag)
	privateKeyFile, _ := cmd.Flags().GetString(pusher.PrivateKeyFileFlag)
	batchingWindow, _ := cmd.Flags().GetString(pusher.BatchingWindowFlag)
	pollingPeriod, _ := cmd.Flags().GetInt(pusher.PollingPeriodFlag)
	limitPerSecond, _ := cmd.Flags().GetInt(pusher.LimitPerSecondFlag)
	burstLimit, _ := cmd.Flags().GetInt(pusher.BurstLimitFlag)
	batchSize, _ := cmd.Flags().GetInt(pusher.BatchSizeFlag)

	logger := PusherLogger(chainRpcUrl, contractAddress)

	payer, err := solana.PrivateKeyFromSolanaKeygenFile(privateKeyFile)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to parse private key")
	}

	interactor, err := NewContractInteractor(
		context.Background(),
		contractAddress,
		payer,
		assetConfigFile,
		pollingPeriod,
		logger,
		limitPerSecond,
		burstLimit,
		batchSize,
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
