package solana

import (
	"context"

	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/pusher"
	"github.com/gagliardetto/solana-go"
	"github.com/spf13/cobra"
)

var SolanapushCmd = &cobra.Command{
	Use:   "solana",
	Short: "Push WebSocket prices to Solana contract",
	Run:   runSolanaPush,
}

func init() {
	SolanapushCmd.Flags().StringP(pusher.StorkWebsocketEndpointFlag, "w", "", pusher.StorkWebsocketEndpointDesc)
	SolanapushCmd.Flags().StringP(pusher.StorkAuthCredentialsFlag, "a", "", pusher.StorkAuthCredentialsDesc)
	SolanapushCmd.Flags().StringP(pusher.ChainRpcUrlFlag, "c", "", pusher.ChainRpcUrlDesc)
	SolanapushCmd.Flags().StringP(pusher.ChainWsUrlFlag, "u", "", pusher.ChainWsUrlDesc)
	SolanapushCmd.Flags().StringP(pusher.ContractAddressFlag, "x", "", pusher.ContractAddressDesc)
	SolanapushCmd.Flags().StringP(pusher.AssetConfigFileFlag, "f", "", pusher.AssetConfigFileDesc)
	SolanapushCmd.Flags().StringP(pusher.PrivateKeyFileFlag, "k", "", pusher.PrivateKeyFileDesc)
	SolanapushCmd.Flags().IntP(pusher.BatchingWindowFlag, "b", 5, pusher.BatchingWindowDesc)
	SolanapushCmd.Flags().IntP(pusher.PollingPeriodFlag, "p", 3, pusher.PollingPeriodDesc)
	SolanapushCmd.Flags().IntP(pusher.LimitPerSecondFlag, "l", 40, pusher.LimitPerSecondDesc)
	SolanapushCmd.Flags().IntP(pusher.BurstLimitFlag, "r", 10, pusher.BurstLimitDesc)
	SolanapushCmd.Flags().IntP(pusher.BatchSizeFlag, "s", 4, pusher.BatchSizeDesc)

	SolanapushCmd.MarkFlagRequired(pusher.StorkWebsocketEndpointFlag)
	SolanapushCmd.MarkFlagRequired(pusher.StorkAuthCredentialsFlag)
	SolanapushCmd.MarkFlagRequired(pusher.ChainRpcUrlFlag)
	SolanapushCmd.MarkFlagRequired(pusher.ContractAddressFlag)
	SolanapushCmd.MarkFlagRequired(pusher.AssetConfigFileFlag)
	SolanapushCmd.MarkFlagRequired(pusher.MnemonicFileFlag)
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

	solanaInteractor, err := NewSolanaContractInteractor(chainRpcUrl, chainWsUrl, contractAddress, payer, assetConfigFile, pollingPeriod, logger, limitPerSecond, burstLimit, batchSize)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize Solana contract interactor")
	}
	solanaPusher := pusher.NewPusher(storkWsEndpoint, storkAuth, chainRpcUrl, contractAddress, assetConfigFile, batchingWindow, pollingPeriod, solanaInteractor, &logger)
	solanaPusher.Run(context.Background())
}
