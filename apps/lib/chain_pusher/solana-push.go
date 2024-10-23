package chain_pusher

import (
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
	SolanapushCmd.Flags().StringP(ContractAddressFlag, "x", "", ContractAddressDesc)
	SolanapushCmd.Flags().StringP(AssetConfigFileFlag, "f", "", AssetConfigFileDesc)
	SolanapushCmd.Flags().StringP(MnemonicFileFlag, "m", "", MnemonicFileDesc)
	SolanapushCmd.Flags().BoolP(VerifyPublishersFlag, "v", false, VerifyPublishersDesc)
	SolanapushCmd.Flags().IntP(BatchingWindowFlag, "b", 5, BatchingWindowDesc)
	SolanapushCmd.Flags().IntP(PollingFrequencyFlag, "p", 3, PollingFrequencyDesc)

	SolanapushCmd.MarkFlagRequired(StorkWebsocketEndpointFlag)
	SolanapushCmd.MarkFlagRequired(StorkAuthCredentialsFlag)
	SolanapushCmd.MarkFlagRequired(ChainRpcUrlFlag)
	SolanapushCmd.MarkFlagRequired(ContractAddressFlag)
	SolanapushCmd.MarkFlagRequired(AssetConfigFileFlag)
	SolanapushCmd.MarkFlagRequired(MnemonicFileFlag)
}

func runSolanaPush(cmd *cobra.Command, args []string) {
	// storkWsEndpoint, _ := cmd.Flags().GetString(StorkWebsocketEndpointFlag)
	// storkAuth, _ := cmd.Flags().GetString(StorkAuthCredentialsFlag)
	// chainRpcUrl, _ := cmd.Flags().GetString(ChainRpcUrlFlag)
	// contractAddress, _ := cmd.Flags().GetString(ContractAddressFlag)
	// assetConfigFile, _ := cmd.Flags().GetString(AssetConfigFileFlag)
	// mnemonicFile, _ := cmd.Flags().GetString(MnemonicFileFlag)
	// verifyPublishers, _ := cmd.Flags().GetBool(VerifyPublishersFlag)
	// batchingWindow, _ := cmd.Flags().GetInt(BatchingWindowFlag)
	// pollingFrequency, _ := cmd.Flags().GetInt(PollingFrequencyFlag)

	// logger := SolanaPusherLogger(chainRpcUrl, contractAddress)

	// TODO - add solanaInteracter

	// evmPusher := NewPusher(storkWsEndpoint, storkAuth, chainRpcUrl, contractAddress, assetConfigFile, mnemonicFile, verifyPublishers, batchingWindow, pollingFrequency, solanaInteracter, &logger)
	// evmPusher.Run()
}
