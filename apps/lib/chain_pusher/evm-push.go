package chain_pusher

import (
	"github.com/spf13/cobra"
)

var EvmpushCmd = &cobra.Command{
	Use:   "evm",
	Short: "Push WebSocket prices to EVM contract",
	Run:   runEvmPush,
}

func init() {
	EvmpushCmd.Flags().StringP(StorkWebsocketEndpointFlag, "w", "", StorkWebsocketEndpointDesc)
	EvmpushCmd.Flags().StringP(StorkAuthCredentialsFlag, "a", "", StorkAuthCredentialsDesc)
	EvmpushCmd.Flags().StringP(ChainRpcUrlFlag, "c", "", ChainRpcUrlDesc)
	EvmpushCmd.Flags().StringP(ContractAddressFlag, "x", "", ContractAddressDesc)
	EvmpushCmd.Flags().StringP(AssetConfigFileFlag, "f", "", AssetConfigFileDesc)
	EvmpushCmd.Flags().StringP(MnemonicFileFlag, "m", "", MnemonicFileDesc)
	EvmpushCmd.Flags().BoolP(VerifyPublishersFlag, "v", false, VerifyPublishersDesc)
	EvmpushCmd.Flags().IntP(BatchingWindowFlag, "b", 5, BatchingWindowDesc)
	EvmpushCmd.Flags().IntP(PollingFrequencyFlag, "p", 3, PollingFrequencyDesc)

	EvmpushCmd.MarkFlagRequired(StorkWebsocketEndpointFlag)
	EvmpushCmd.MarkFlagRequired(StorkAuthCredentialsFlag)
	EvmpushCmd.MarkFlagRequired(ChainRpcUrlFlag)
	EvmpushCmd.MarkFlagRequired(ContractAddressFlag)
	EvmpushCmd.MarkFlagRequired(AssetConfigFileFlag)
	EvmpushCmd.MarkFlagRequired(MnemonicFileFlag)
}

func runEvmPush(cmd *cobra.Command, args []string) {
	storkWsEndpoint, _ := cmd.Flags().GetString(StorkWebsocketEndpointFlag)
	storkAuth, _ := cmd.Flags().GetString(StorkAuthCredentialsFlag)
	chainRpcUrl, _ := cmd.Flags().GetString(ChainRpcUrlFlag)
	contractAddress, _ := cmd.Flags().GetString(ContractAddressFlag)
	assetConfigFile, _ := cmd.Flags().GetString(AssetConfigFileFlag)
	mnemonicFile, _ := cmd.Flags().GetString(MnemonicFileFlag)
	verifyPublishers, _ := cmd.Flags().GetBool(VerifyPublishersFlag)
	batchingWindow, _ := cmd.Flags().GetInt(BatchingWindowFlag)
	pollingFrequency, _ := cmd.Flags().GetInt(PollingFrequencyFlag)

	logger := EvmPusherLogger(chainRpcUrl, contractAddress)

	evmInteracter := NewEvmContractInteracter(chainRpcUrl, contractAddress, mnemonicFile, pollingFrequency, verifyPublishers, logger)

	evmPusher := NewPusher(storkWsEndpoint, storkAuth, chainRpcUrl, contractAddress, assetConfigFile, mnemonicFile, verifyPublishers, batchingWindow, pollingFrequency, evmInteracter, &logger)
	evmPusher.Run()
}
