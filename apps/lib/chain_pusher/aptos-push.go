package chain_pusher

import (
	"github.com/spf13/cobra"
)

var AptospushCmd = &cobra.Command{
	Use:   "aptos",
	Short: "Push WebSocket prices to Aptos contract",
	Run:   runAptosPush,
}

func init() {
	AptospushCmd.Flags().StringP(StorkWebsocketEndpointFlag, "w", "", StorkWebsocketEndpointDesc)
	AptospushCmd.Flags().StringP(StorkAuthCredentialsFlag, "a", "", StorkAuthCredentialsDesc)
	AptospushCmd.Flags().StringP(ChainRpcUrlFlag, "c", "", ChainRpcUrlDesc)
	AptospushCmd.Flags().StringP(ContractAddressFlag, "x", "", ContractAddressDesc)
	AptospushCmd.Flags().StringP(AssetConfigFileFlag, "f", "", AssetConfigFileDesc)
	AptospushCmd.Flags().StringP(PrivateKeyFileFlag, "k", "", PrivateKeyFileDesc)
	AptospushCmd.Flags().IntP(BatchingWindowFlag, "b", 5, BatchingWindowDesc)
	AptospushCmd.Flags().IntP(PollingFrequencyFlag, "p", 3, PollingFrequencyDesc)

	AptospushCmd.MarkFlagRequired(StorkWebsocketEndpointFlag)
	AptospushCmd.MarkFlagRequired(StorkAuthCredentialsFlag)
	AptospushCmd.MarkFlagRequired(ChainRpcUrlFlag)
	AptospushCmd.MarkFlagRequired(ContractAddressFlag)
	AptospushCmd.MarkFlagRequired(AssetConfigFileFlag)
	AptospushCmd.MarkFlagRequired(PrivateKeyFileFlag)
}

func runAptosPush(cmd *cobra.Command, args []string) {
	storkWsEndpoint, _ := cmd.Flags().GetString(StorkWebsocketEndpointFlag)
	storkAuth, _ := cmd.Flags().GetString(StorkAuthCredentialsFlag)
	chainRpcUrl, _ := cmd.Flags().GetString(ChainRpcUrlFlag)
	contractAddress, _ := cmd.Flags().GetString(ContractAddressFlag)
	assetConfigFile, _ := cmd.Flags().GetString(AssetConfigFileFlag)
	privateKeyFile, _ := cmd.Flags().GetString(PrivateKeyFileFlag)
	batchingWindow, _ := cmd.Flags().GetInt(BatchingWindowFlag)
	pollingFrequency, _ := cmd.Flags().GetInt(PollingFrequencyFlag)

	logger := AptosPusherLogger(chainRpcUrl, contractAddress)

	aptosInteracter, err := NewAptosContractInteracter(chainRpcUrl, contractAddress, privateKeyFile, assetConfigFile, pollingFrequency, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize Aptos contract interacter")
	}

	aptosPusher := NewPusher(storkWsEndpoint, storkAuth, chainRpcUrl, contractAddress, assetConfigFile, batchingWindow, pollingFrequency, aptosInteracter, &logger)
	aptosPusher.Run()
}
