package sui

import (
	"context"
	"os"

	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/pusher"
	"github.com/spf13/cobra"
)

var SuipushCmd = &cobra.Command{
	Use:   "sui",
	Short: "Push WebSocket prices to Sui contract",
	Run:   runSuiPush,
}

func init() {
	SuipushCmd.Flags().StringP(pusher.StorkWebsocketEndpointFlag, "w", "", pusher.StorkWebsocketEndpointDesc)
	SuipushCmd.Flags().StringP(pusher.StorkAuthCredentialsFlag, "a", "", pusher.StorkAuthCredentialsDesc)
	SuipushCmd.Flags().StringP(pusher.ChainRpcUrlFlag, "c", "", pusher.ChainRpcUrlDesc)
	SuipushCmd.Flags().StringP(pusher.ContractAddressFlag, "x", "", pusher.ContractAddressDesc)
	SuipushCmd.Flags().StringP(pusher.AssetConfigFileFlag, "f", "", pusher.AssetConfigFileDesc)
	SuipushCmd.Flags().StringP(pusher.PrivateKeyFileFlag, "k", "", pusher.PrivateKeyFileDesc)
	SuipushCmd.Flags().IntP(pusher.BatchingWindowFlag, "b", 5, pusher.BatchingWindowDesc)
	SuipushCmd.Flags().IntP(pusher.PollingPeriodFlag, "p", 3, pusher.PollingPeriodDesc)

	SuipushCmd.MarkFlagRequired(pusher.StorkWebsocketEndpointFlag)
	SuipushCmd.MarkFlagRequired(pusher.StorkAuthCredentialsFlag)
	SuipushCmd.MarkFlagRequired(pusher.ChainRpcUrlFlag)
	SuipushCmd.MarkFlagRequired(pusher.ContractAddressFlag)
	SuipushCmd.MarkFlagRequired(pusher.AssetConfigFileFlag)
	SuipushCmd.MarkFlagRequired(pusher.PrivateKeyFileFlag)
}

func runSuiPush(cmd *cobra.Command, args []string) {
	storkWsEndpoint, _ := cmd.Flags().GetString(pusher.StorkWebsocketEndpointFlag)
	storkAuth, _ := cmd.Flags().GetString(pusher.StorkAuthCredentialsFlag)
	chainRpcUrl, _ := cmd.Flags().GetString(pusher.ChainRpcUrlFlag)
	contractAddress, _ := cmd.Flags().GetString(pusher.ContractAddressFlag)
	assetConfigFile, _ := cmd.Flags().GetString(pusher.AssetConfigFileFlag)
	privateKeyFile, _ := cmd.Flags().GetString(pusher.PrivateKeyFileFlag)
	batchingWindow, _ := cmd.Flags().GetInt(pusher.BatchingWindowFlag)
	pollingPeriod, _ := cmd.Flags().GetInt(pusher.PollingPeriodFlag)

	logger := pusher.SuiPusherLogger(chainRpcUrl, contractAddress)

	keyFileContent, err := os.ReadFile(privateKeyFile)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to read private key file")
	}

	suiInteractor, err := NewSuiContractInteractor(chainRpcUrl, contractAddress, keyFileContent, assetConfigFile, pollingPeriod, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize Sui contract interactor")
	}

	suiPusher := pusher.NewPusher(storkWsEndpoint, storkAuth, chainRpcUrl, contractAddress, assetConfigFile, batchingWindow, pollingPeriod, suiInteractor, &logger)
	suiPusher.Run(context.Background())
}
