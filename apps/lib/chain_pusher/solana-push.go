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
	SolanapushCmd.Flags().StringP(PrivateKeyFileFlag, "k", "", PrivateKeyFileDesc)
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
	storkWsEndpoint, _ := cmd.Flags().GetString(StorkWebsocketEndpointFlag)
	storkAuth, _ := cmd.Flags().GetString(StorkAuthCredentialsFlag)
	chainRpcUrl, _ := cmd.Flags().GetString(ChainRpcUrlFlag)
	contractAddress, _ := cmd.Flags().GetString(ContractAddressFlag)
	assetConfigFile, _ := cmd.Flags().GetString(AssetConfigFileFlag)
	privateKeyFile, _ := cmd.Flags().GetString(PrivateKeyFileFlag)
	batchingWindow, _ := cmd.Flags().GetInt(BatchingWindowFlag)
	pollingFrequency, _ := cmd.Flags().GetInt(PollingFrequencyFlag)

	logger := SolanaPusherLogger(chainRpcUrl, contractAddress)

	solanaInteracter, err := NewSolanaContractInteracter(chainRpcUrl, storkWsEndpoint, contractAddress, privateKeyFile, assetConfigFile, pollingFrequency, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize Solana contract interacter")
	}
	solanaPusher := NewPusher(storkWsEndpoint, storkAuth, chainRpcUrl, contractAddress, assetConfigFile, batchingWindow, pollingFrequency, solanaInteracter, &logger)
	solanaPusher.Run()
}
