package chain_pusher

import (
	"context"

	"github.com/gagliardetto/solana-go"
	"github.com/spf13/cobra"
)

var SolanapushCmd = &cobra.Command{
	Use:   "solana",
	Short: "Push WebSocket prices to Solana contract",
	Run:   runSolanaPush,
}

func init() {
	SolanapushCmd.Flags().StringP(StorkWebsocketEndpointFlag, "w", "", StorkWebsocketEndpointDesc)
	SolanapushCmd.Flags().StringP(StorkAuthCredentialsFlag, "a", "", StorkAuthCredentialsDesc)
	SolanapushCmd.Flags().StringP(ChainRpcUrlFlag, "c", "", ChainRpcUrlDesc)
	SolanapushCmd.Flags().StringP(ChainWsUrlFlag, "u", "", ChainWsUrlDesc)
	SolanapushCmd.Flags().StringP(ContractAddressFlag, "x", "", ContractAddressDesc)
	SolanapushCmd.Flags().StringP(AssetConfigFileFlag, "f", "", AssetConfigFileDesc)
	SolanapushCmd.Flags().StringP(PrivateKeyFileFlag, "k", "", PrivateKeyFileDesc)
	SolanapushCmd.Flags().IntP(BatchingWindowFlag, "b", 5, BatchingWindowDesc)
	SolanapushCmd.Flags().IntP(PollingFrequencyFlag, "p", 3, PollingFrequencyDesc)
	SolanapushCmd.Flags().IntP(LimitPerSecondFlag, "l", 40, LimitPerSecondDesc)
	SolanapushCmd.Flags().IntP(BurstLimitFlag, "r", 10, BurstLimitDesc)
	SolanapushCmd.Flags().IntP(BatchSizeFlag, "s", 4, BatchSizeDesc)

	SolanapushCmd.MarkFlagRequired(StorkWebsocketEndpointFlag)
	SolanapushCmd.MarkFlagRequired(StorkAuthCredentialsFlag)
	SolanapushCmd.MarkFlagRequired(ChainRpcUrlFlag)
	SolanapushCmd.MarkFlagRequired(ContractAddressFlag)
	SolanapushCmd.MarkFlagRequired(AssetConfigFileFlag)
	SolanapushCmd.MarkFlagRequired(MnemonicFileFlag)
}

func runSolanaPush(cmd *cobra.Command, args []string) {
	storkWsEndpoint, _ := cmd.Flags().GetString(StorkWebsocketEndpointFlag)
	storkAuth, _ := cmd.Flags().GetString(StorkAuthCredentialsFlag)
	chainRpcUrl, _ := cmd.Flags().GetString(ChainRpcUrlFlag)
	chainWsUrl, _ := cmd.Flags().GetString(ChainWsUrlFlag)
	contractAddress, _ := cmd.Flags().GetString(ContractAddressFlag)
	assetConfigFile, _ := cmd.Flags().GetString(AssetConfigFileFlag)
	privateKeyFile, _ := cmd.Flags().GetString(PrivateKeyFileFlag)
	batchingWindow, _ := cmd.Flags().GetInt(BatchingWindowFlag)
	pollingFrequency, _ := cmd.Flags().GetInt(PollingFrequencyFlag)
	limitPerSecond, _ := cmd.Flags().GetInt(LimitPerSecondFlag)
	burstLimit, _ := cmd.Flags().GetInt(BurstLimitFlag)
	batchSize, _ := cmd.Flags().GetInt(BatchSizeFlag)

	logger := SolanaPusherLogger(chainRpcUrl, contractAddress)

	payer, err := solana.PrivateKeyFromSolanaKeygenFile(privateKeyFile)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to parse private key")
	}

	solanaInteractor, err := NewSolanaContractInteractor(chainRpcUrl, chainWsUrl, contractAddress, payer, assetConfigFile, pollingFrequency, logger, limitPerSecond, burstLimit, batchSize)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize Solana contract interactor")
	}
	solanaPusher := NewPusher(storkWsEndpoint, storkAuth, chainRpcUrl, contractAddress, assetConfigFile, batchingWindow, pollingFrequency, solanaInteractor, &logger)
	solanaPusher.Run(context.Background())
}
