package aptos

import (
	"context"
	"os"

	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/pusher"
	"github.com/spf13/cobra"
)

var AptospushCmd = &cobra.Command{
	Use:   "aptos",
	Short: "Push WebSocket prices to Aptos contract",
	Run:   runAptosPush,
}

func init() {
	AptospushCmd.Flags().StringP(pusher.StorkWebsocketEndpointFlag, "w", "", pusher.StorkWebsocketEndpointDesc)
	AptospushCmd.Flags().StringP(pusher.StorkAuthCredentialsFlag, "a", "", pusher.StorkAuthCredentialsDesc)
	AptospushCmd.Flags().StringP(pusher.ChainRpcUrlFlag, "c", "", pusher.ChainRpcUrlDesc)
	AptospushCmd.Flags().StringP(pusher.ContractAddressFlag, "x", "", pusher.ContractAddressDesc)
	AptospushCmd.Flags().StringP(pusher.AssetConfigFileFlag, "f", "", pusher.AssetConfigFileDesc)
	AptospushCmd.Flags().StringP(pusher.PrivateKeyFileFlag, "k", "", pusher.PrivateKeyFileDesc)
	AptospushCmd.Flags().IntP(pusher.BatchingWindowFlag, "b", 5, pusher.BatchingWindowDesc)
	AptospushCmd.Flags().IntP(pusher.PollingPeriodFlag, "p", 3, pusher.PollingPeriodDesc)

	AptospushCmd.MarkFlagRequired(pusher.StorkWebsocketEndpointFlag)
	AptospushCmd.MarkFlagRequired(pusher.StorkAuthCredentialsFlag)
	AptospushCmd.MarkFlagRequired(pusher.ChainRpcUrlFlag)
	AptospushCmd.MarkFlagRequired(pusher.ContractAddressFlag)
	AptospushCmd.MarkFlagRequired(pusher.AssetConfigFileFlag)
	AptospushCmd.MarkFlagRequired(pusher.PrivateKeyFileFlag)
}

func runAptosPush(cmd *cobra.Command, args []string) {
	storkWsEndpoint, _ := cmd.Flags().GetString(pusher.StorkWebsocketEndpointFlag)
	storkAuth, _ := cmd.Flags().GetString(pusher.StorkAuthCredentialsFlag)
	chainRpcUrl, _ := cmd.Flags().GetString(pusher.ChainRpcUrlFlag)
	contractAddress, _ := cmd.Flags().GetString(pusher.ContractAddressFlag)
	assetConfigFile, _ := cmd.Flags().GetString(pusher.AssetConfigFileFlag)
	privateKeyFile, _ := cmd.Flags().GetString(pusher.PrivateKeyFileFlag)
	batchingWindow, _ := cmd.Flags().GetInt(pusher.BatchingWindowFlag)
	pollingPeriod, _ := cmd.Flags().GetInt(pusher.PollingPeriodFlag)

	logger := pusher.AptosPusherLogger(chainRpcUrl, contractAddress)

	keyFileContent, err := os.ReadFile(privateKeyFile)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to read private key file")
	}

	aptosInteractor, err := NewAptosContractInteractor(chainRpcUrl, contractAddress, keyFileContent, pollingPeriod, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize Aptos contract interactor")
	}

	aptosPusher := pusher.NewPusher(storkWsEndpoint, storkAuth, chainRpcUrl, contractAddress, assetConfigFile, batchingWindow, pollingPeriod, aptosInteractor, &logger)
	aptosPusher.Run(context.Background())
}
