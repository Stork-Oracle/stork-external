package chain_pusher

import (
	"context"
	"os"

	"github.com/spf13/cobra"
)

var SuipushCmd = &cobra.Command{
	Use:   "sui",
	Short: "Push WebSocket prices to Sui contract",
	Run:   runSuiPush,
}

func init() {
	SuipushCmd.Flags().StringP(StorkWebsocketEndpointFlag, "w", "", StorkWebsocketEndpointDesc)
	SuipushCmd.Flags().StringP(StorkAuthCredentialsFlag, "a", "", StorkAuthCredentialsDesc)
	SuipushCmd.Flags().StringP(ChainRpcUrlFlag, "c", "", ChainRpcUrlDesc)
	SuipushCmd.Flags().StringP(ContractAddressFlag, "x", "", ContractAddressDesc)
	SuipushCmd.Flags().StringP(AssetConfigFileFlag, "f", "", AssetConfigFileDesc)
	SuipushCmd.Flags().StringP(PrivateKeyFileFlag, "k", "", PrivateKeyFileDesc)
	SuipushCmd.Flags().IntP(BatchingWindowFlag, "b", 5, BatchingWindowDesc)
	SuipushCmd.Flags().IntP(PollingPeriodFlag, "p", 3, PollingPeriodDesc)

	SuipushCmd.MarkFlagRequired(StorkWebsocketEndpointFlag)
	SuipushCmd.MarkFlagRequired(StorkAuthCredentialsFlag)
	SuipushCmd.MarkFlagRequired(ChainRpcUrlFlag)
	SuipushCmd.MarkFlagRequired(ContractAddressFlag)
	SuipushCmd.MarkFlagRequired(AssetConfigFileFlag)
	SuipushCmd.MarkFlagRequired(PrivateKeyFileFlag)
}

func runSuiPush(cmd *cobra.Command, args []string) {
	storkWsEndpoint, _ := cmd.Flags().GetString(StorkWebsocketEndpointFlag)
	storkAuth, _ := cmd.Flags().GetString(StorkAuthCredentialsFlag)
	chainRpcUrl, _ := cmd.Flags().GetString(ChainRpcUrlFlag)
	contractAddress, _ := cmd.Flags().GetString(ContractAddressFlag)
	assetConfigFile, _ := cmd.Flags().GetString(AssetConfigFileFlag)
	privateKeyFile, _ := cmd.Flags().GetString(PrivateKeyFileFlag)
	batchingWindow, _ := cmd.Flags().GetInt(BatchingWindowFlag)
	pollingPeriod, _ := cmd.Flags().GetInt(PollingPeriodFlag)

	logger := SuiPusherLogger(chainRpcUrl, contractAddress)

	keyFileContent, err := os.ReadFile(privateKeyFile)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to read private key file")
	}

	suiInteractor, err := NewSuiContractInteractor(chainRpcUrl, contractAddress, keyFileContent, assetConfigFile, pollingPeriod, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize Sui contract interactor")
	}

	suiPusher := NewPusher(storkWsEndpoint, storkAuth, chainRpcUrl, contractAddress, assetConfigFile, batchingWindow, pollingPeriod, suiInteractor, &logger)
	suiPusher.Run(context.Background())
}
