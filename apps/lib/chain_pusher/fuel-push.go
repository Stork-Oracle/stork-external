package chain_pusher

import (
	"context"
	"os"

	"github.com/spf13/cobra"
)

var FuelpushCmd = &cobra.Command{
	Use:   "fuel",
	Short: "Push WebSocket prices to Fuel contract",
	Run:   runFuelPush,
}

func init() {
	FuelpushCmd.Flags().StringP(StorkWebsocketEndpointFlag, "w", "", StorkWebsocketEndpointDesc)
	FuelpushCmd.Flags().StringP(StorkAuthCredentialsFlag, "a", "", StorkAuthCredentialsDesc)
	FuelpushCmd.Flags().StringP(ChainRpcUrlFlag, "c", "", ChainRpcUrlDesc)
	FuelpushCmd.Flags().StringP(ContractAddressFlag, "x", "", ContractAddressDesc)
	FuelpushCmd.Flags().StringP(AssetConfigFileFlag, "f", "", AssetConfigFileDesc)
	FuelpushCmd.Flags().StringP(PrivateKeyFileFlag, "k", "", PrivateKeyFileDesc)
	FuelpushCmd.Flags().IntP(BatchingWindowFlag, "b", 5, BatchingWindowDesc)
	FuelpushCmd.Flags().IntP(PollingPeriodFlag, "p", 3, PollingPeriodDesc)

	FuelpushCmd.MarkFlagRequired(StorkWebsocketEndpointFlag)
	FuelpushCmd.MarkFlagRequired(StorkAuthCredentialsFlag)
	FuelpushCmd.MarkFlagRequired(ChainRpcUrlFlag)
	FuelpushCmd.MarkFlagRequired(ContractAddressFlag)
	FuelpushCmd.MarkFlagRequired(AssetConfigFileFlag)
	FuelpushCmd.MarkFlagRequired(PrivateKeyFileFlag)
}

func runFuelPush(cmd *cobra.Command, args []string) {
	storkWsEndpoint, _ := cmd.Flags().GetString(StorkWebsocketEndpointFlag)
	storkAuth, _ := cmd.Flags().GetString(StorkAuthCredentialsFlag)
	chainRpcUrl, _ := cmd.Flags().GetString(ChainRpcUrlFlag)
	contractAddress, _ := cmd.Flags().GetString(ContractAddressFlag)
	assetConfigFile, _ := cmd.Flags().GetString(AssetConfigFileFlag)
	privateKeyFile, _ := cmd.Flags().GetString(PrivateKeyFileFlag)
	batchingWindow, _ := cmd.Flags().GetInt(BatchingWindowFlag)
	pollingPeriod, _ := cmd.Flags().GetInt(PollingPeriodFlag)

	logger := FuelPusherLogger(chainRpcUrl, contractAddress)

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

	fuelPusher := NewPusher(storkWsEndpoint, storkAuth, chainRpcUrl, contractAddress, assetConfigFile, batchingWindow, pollingPeriod, fuelInteractor, &logger)
	fuelPusher.Run(context.Background())
}
