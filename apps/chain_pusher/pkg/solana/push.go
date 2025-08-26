package solana

import (
	"context"

	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/pusher"
	"github.com/gagliardetto/solana-go"
	"github.com/spf13/cobra"
)

var PushCmd = &cobra.Command{
	Use:   "solana",
	Short: "Push WebSocket prices to Solana contract",
	Run:   runSolanaPush,
}

func init() {
	PushCmd.Flags().StringP(pusher.StorkWebsocketEndpointFlag, "w", "", pusher.StorkWebsocketEndpointDesc)
	PushCmd.Flags().StringP(pusher.StorkAuthCredentialsFlag, "a", "", pusher.StorkAuthCredentialsDesc)
	PushCmd.Flags().StringP(pusher.ChainRpcUrlFlag, "c", "", pusher.ChainRpcUrlDesc)
	PushCmd.Flags().StringP(pusher.ChainWsUrlFlag, "u", "", pusher.ChainWsUrlDesc)
	PushCmd.Flags().StringP(pusher.ContractAddressFlag, "x", "", pusher.ContractAddressDesc)
	PushCmd.Flags().StringP(pusher.AssetConfigFileFlag, "f", "", pusher.AssetConfigFileDesc)
	PushCmd.Flags().StringP(pusher.PrivateKeyFileFlag, "k", "", pusher.PrivateKeyFileDesc)
	PushCmd.Flags().IntP(pusher.BatchingWindowFlag, "b", 5, pusher.BatchingWindowDesc)
	PushCmd.Flags().IntP(pusher.PollingPeriodFlag, "p", 3, pusher.PollingPeriodDesc)
	PushCmd.Flags().IntP(pusher.LimitPerSecondFlag, "l", 40, pusher.LimitPerSecondDesc)
	PushCmd.Flags().IntP(pusher.BurstLimitFlag, "r", 10, pusher.BurstLimitDesc)
	PushCmd.Flags().IntP(pusher.BatchSizeFlag, "s", 4, pusher.BatchSizeDesc)

	PushCmd.MarkFlagRequired(pusher.StorkWebsocketEndpointFlag)
	PushCmd.MarkFlagRequired(pusher.StorkAuthCredentialsFlag)
	PushCmd.MarkFlagRequired(pusher.ChainRpcUrlFlag)
	PushCmd.MarkFlagRequired(pusher.ContractAddressFlag)
	PushCmd.MarkFlagRequired(pusher.AssetConfigFileFlag)
	PushCmd.MarkFlagRequired(pusher.MnemonicFileFlag)
}

func runSolanaPush(cmd *cobra.Command, args []string) {
	storkWsEndpoint, _ := cmd.Flags().GetString(pusher.StorkWebsocketEndpointFlag)
	storkAuth, _ := cmd.Flags().GetString(pusher.StorkAuthCredentialsFlag)
	chainRpcUrl, _ := cmd.Flags().GetString(pusher.ChainRpcUrlFlag)
	chainWsUrl, _ := cmd.Flags().GetString(pusher.ChainWsUrlFlag)
	contractAddress, _ := cmd.Flags().GetString(pusher.ContractAddressFlag)
	assetConfigFile, _ := cmd.Flags().GetString(pusher.AssetConfigFileFlag)
	privateKeyFile, _ := cmd.Flags().GetString(pusher.PrivateKeyFileFlag)
	batchingWindow, _ := cmd.Flags().GetInt(pusher.BatchingWindowFlag)
	pollingPeriod, _ := cmd.Flags().GetInt(pusher.PollingPeriodFlag)
	limitPerSecond, _ := cmd.Flags().GetInt(pusher.LimitPerSecondFlag)
	burstLimit, _ := cmd.Flags().GetInt(pusher.BurstLimitFlag)
	batchSize, _ := cmd.Flags().GetInt(pusher.BatchSizeFlag)

	logger := pusher.SolanaPusherLogger(chainRpcUrl, contractAddress)

	payer, err := solana.PrivateKeyFromSolanaKeygenFile(privateKeyFile)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to parse private key")
	}

	interactor, err := NewContractInteractor(chainRpcUrl, chainWsUrl, contractAddress, payer, assetConfigFile, pollingPeriod, logger, limitPerSecond, burstLimit, batchSize)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize contract interactor")
	}
	pusher := pusher.NewPusher(storkWsEndpoint, storkAuth, chainRpcUrl, contractAddress, assetConfigFile, batchingWindow, pollingPeriod, interactor, &logger)
	pusher.Run(context.Background())
}
