package chain_pusher

import (
	"github.com/spf13/cobra"
)

var CosmwasmPushCmd = &cobra.Command{
	Use:   "cosmwasm",
	Short: "Push WebSocket prices to Cosmwasm contract",
	Run:   runCosmwasmPush,
}

func init() {
	CosmwasmPushCmd.Flags().StringP(StorkWebsocketEndpointFlag, "w", "", StorkWebsocketEndpointDesc)
	CosmwasmPushCmd.Flags().StringP(StorkAuthCredentialsFlag, "a", "", StorkAuthCredentialsDesc)
	CosmwasmPushCmd.Flags().StringP(ChainGrpcUrlFlag, "g", "", ChainGrpcUrlDesc)
	CosmwasmPushCmd.Flags().StringP(ContractAddressFlag, "x", "", ContractAddressDesc)
	CosmwasmPushCmd.Flags().StringP(AssetConfigFileFlag, "f", "", AssetConfigFileDesc)
	CosmwasmPushCmd.Flags().StringP(MnemonicFileFlag, "m", "", MnemonicFileDesc)
	CosmwasmPushCmd.Flags().IntP(BatchingWindowFlag, "b", 5, BatchingWindowDesc)
	CosmwasmPushCmd.Flags().IntP(PollingFrequencyFlag, "p", 3, PollingFrequencyDesc)

	CosmwasmPushCmd.MarkFlagRequired(StorkWebsocketEndpointFlag)
	CosmwasmPushCmd.MarkFlagRequired(StorkAuthCredentialsFlag)
	CosmwasmPushCmd.MarkFlagRequired(ChainGrpcUrlFlag)
	CosmwasmPushCmd.MarkFlagRequired(ContractAddressFlag)
	CosmwasmPushCmd.MarkFlagRequired(AssetConfigFileFlag)
	CosmwasmPushCmd.MarkFlagRequired(MnemonicFileFlag)
}

func runCosmwasmPush(cmd *cobra.Command, args []string) {
	storkWsEndpoint, _ := cmd.Flags().GetString(StorkWebsocketEndpointFlag)
	storkAuth, _ := cmd.Flags().GetString(StorkAuthCredentialsFlag)
	chainGrpcUrl, _ := cmd.Flags().GetString(ChainGrpcUrlFlag)
	contractAddress, _ := cmd.Flags().GetString(ContractAddressFlag)
	assetConfigFile, _ := cmd.Flags().GetString(AssetConfigFileFlag)
	mnemonicFile, _ := cmd.Flags().GetString(MnemonicFileFlag)
	batchingWindow, _ := cmd.Flags().GetInt(BatchingWindowFlag)
	pollingFrequency, _ := cmd.Flags().GetInt(PollingFrequencyFlag)

	logger := CosmwasmPusherLogger(chainGrpcUrl, contractAddress)

	cosmwasmInteracter, err := NewCosmwasmContractInteracter(chainGrpcUrl, contractAddress, mnemonicFile, batchingWindow, pollingFrequency, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create cosmwasm interacter")
	}

	cosmwasmPusher := NewPusher(storkWsEndpoint, storkAuth, chainGrpcUrl, contractAddress, assetConfigFile, batchingWindow, pollingFrequency, cosmwasmInteracter, &logger)
	cosmwasmPusher.Run()
}
