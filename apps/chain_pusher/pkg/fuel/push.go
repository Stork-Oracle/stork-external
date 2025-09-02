package fuel

import (
	"context"
	"os"

	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/pusher"
	"github.com/spf13/cobra"
)

func NewPushCmd() *cobra.Command {
	pushCmd := &cobra.Command{
		Use:   "fuel",
		Short: "Push WebSocket prices to Fuel contract",
		Run:   runPush,
	}

	pushCmd.Flags().StringP(pusher.StorkWebsocketEndpointFlag, "w", "", pusher.StorkWebsocketEndpointDesc)
	pushCmd.Flags().StringP(pusher.StorkAuthCredentialsFlag, "a", "", pusher.StorkAuthCredentialsDesc)
	pushCmd.Flags().StringP(pusher.ChainRpcUrlFlag, "c", "", pusher.ChainRpcUrlDesc)
	pushCmd.Flags().StringP(pusher.ContractAddressFlag, "x", "", pusher.ContractAddressDesc)
	pushCmd.Flags().StringP(pusher.AssetConfigFileFlag, "f", "", pusher.AssetConfigFileDesc)
	pushCmd.Flags().StringP(pusher.PrivateKeyFileFlag, "k", "", pusher.PrivateKeyFileDesc)
	pushCmd.Flags().IntP(pusher.BatchingWindowFlag, "b", pusher.DefaultBatchingWindow, pusher.BatchingWindowDesc)
	pushCmd.Flags().IntP(pusher.PollingPeriodFlag, "p", pusher.DefaultPollingPeriod, pusher.PollingPeriodDesc)

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
	contractAddress, _ := cmd.Flags().GetString(pusher.ContractAddressFlag)
	assetConfigFile, _ := cmd.Flags().GetString(pusher.AssetConfigFileFlag)
	privateKeyFile, _ := cmd.Flags().GetString(pusher.PrivateKeyFileFlag)
	batchingWindow, _ := cmd.Flags().GetInt(pusher.BatchingWindowFlag)
	pollingPeriod, _ := cmd.Flags().GetInt(pusher.PollingPeriodFlag)

	logger := PusherLogger(chainRpcUrl, contractAddress)

	keyFileContent, err := os.ReadFile(privateKeyFile)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to read private key file")
	}

	interactor, err := NewContractInteractor(chainRpcUrl, contractAddress, keyFileContent, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize contract interactor")
	}

	// Ensure cleanup on exit
	defer interactor.Close()

	pusher := pusher.NewPusher(storkWsEndpoint, storkAuth, chainRpcUrl, contractAddress, assetConfigFile, batchingWindow, pollingPeriod, interactor, &logger)
	pusher.Run(context.Background())
}
