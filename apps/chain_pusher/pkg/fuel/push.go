package fuel

import (
	"context"
	"os"

	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/pusher"
	"github.com/spf13/cobra"
)

var FuelpushCmd = &cobra.Command{
	Use:   "fuel",
	Short: "Push WebSocket prices to Fuel contract",
	Run:   runFuelPush,
}

func init() {
	FuelpushCmd.Flags().StringP(pusher.StorkWebsocketEndpointFlag, "w", "", pusher.StorkWebsocketEndpointDesc)
	FuelpushCmd.Flags().StringP(pusher.StorkAuthCredentialsFlag, "a", "", pusher.StorkAuthCredentialsDesc)
	FuelpushCmd.Flags().StringP(pusher.ChainRpcUrlFlag, "c", "", pusher.ChainRpcUrlDesc)
	FuelpushCmd.Flags().StringP(pusher.ContractAddressFlag, "x", "", pusher.ContractAddressDesc)
	FuelpushCmd.Flags().StringP(pusher.AssetConfigFileFlag, "f", "", pusher.AssetConfigFileDesc)
	FuelpushCmd.Flags().StringP(pusher.PrivateKeyFileFlag, "k", "", pusher.PrivateKeyFileDesc)
	FuelpushCmd.Flags().IntP(pusher.BatchingWindowFlag, "b", 5, pusher.BatchingWindowDesc)
	FuelpushCmd.Flags().IntP(pusher.PollingPeriodFlag, "p", 3, pusher.PollingPeriodDesc)

	FuelpushCmd.MarkFlagRequired(pusher.StorkWebsocketEndpointFlag)
	FuelpushCmd.MarkFlagRequired(pusher.StorkAuthCredentialsFlag)
	FuelpushCmd.MarkFlagRequired(pusher.ChainRpcUrlFlag)
	FuelpushCmd.MarkFlagRequired(pusher.ContractAddressFlag)
	FuelpushCmd.MarkFlagRequired(pusher.AssetConfigFileFlag)
	FuelpushCmd.MarkFlagRequired(pusher.PrivateKeyFileFlag)
}

func runFuelPush(cmd *cobra.Command, args []string) {
	storkWsEndpoint, _ := cmd.Flags().GetString(pusher.StorkWebsocketEndpointFlag)
	storkAuth, _ := cmd.Flags().GetString(pusher.StorkAuthCredentialsFlag)
	chainRpcUrl, _ := cmd.Flags().GetString(pusher.ChainRpcUrlFlag)
	contractAddress, _ := cmd.Flags().GetString(pusher.ContractAddressFlag)
	assetConfigFile, _ := cmd.Flags().GetString(pusher.AssetConfigFileFlag)
	privateKeyFile, _ := cmd.Flags().GetString(pusher.PrivateKeyFileFlag)
	batchingWindow, _ := cmd.Flags().GetInt(pusher.BatchingWindowFlag)
	pollingPeriod, _ := cmd.Flags().GetInt(pusher.PollingPeriodFlag)

	logger := pusher.FuelPusherLogger(chainRpcUrl, contractAddress)

	keyFileContent, err := os.ReadFile(privateKeyFile)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to read private key file")
	}

	// Extract private key from file content (assuming it's on the first line)
	privateKeyStr := string(keyFileContent)
	if len(privateKeyStr) > 0 && privateKeyStr[len(privateKeyStr)-1] == '\n' {
		privateKeyStr = privateKeyStr[:len(privateKeyStr)-1]
	}

	fuelInteractor, err := NewFuelContractInteractor(chainRpcUrl, contractAddress, privateKeyStr, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize Fuel contract interactor")
	}

	// Ensure cleanup on exit
	defer fuelInteractor.Close()

	fuelPusher := pusher.NewPusher(storkWsEndpoint, storkAuth, chainRpcUrl, contractAddress, assetConfigFile, batchingWindow, pollingPeriod, fuelInteractor, &logger)
	fuelPusher.Run(context.Background())
}
