package chain_pusher

import (
	"context"

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
	SuipushCmd.Flags().IntP(PollingFrequencyFlag, "p", 3, PollingFrequencyDesc)

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
	pollingFrequency, _ := cmd.Flags().GetInt(PollingFrequencyFlag)

	logger := SuiPusherLogger(chainRpcUrl, contractAddress)

	suiInteractor, err := NewSuiContractInteractor(chainRpcUrl, contractAddress, []byte(privateKeyFile), assetConfigFile, pollingFrequency, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize Sui contract interactor")
	}

	suiPusher := NewPusher(storkWsEndpoint, storkAuth, chainRpcUrl, contractAddress, assetConfigFile, batchingWindow, pollingFrequency, suiInteractor, &logger)
	suiPusher.Run(context.Background())
}
