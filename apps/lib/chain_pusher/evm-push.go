package chain_pusher

import (
	"context"
	"os"

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
	EvmpushCmd.Flags().StringP(ChainWsUrlFlag, "u", "", ChainWsUrlDesc)
	EvmpushCmd.Flags().StringP(ContractAddressFlag, "x", "", ContractAddressDesc)
	EvmpushCmd.Flags().StringP(AssetConfigFileFlag, "f", "", AssetConfigFileDesc)
	EvmpushCmd.Flags().StringP(MnemonicFileFlag, "m", "", MnemonicFileDesc)
	EvmpushCmd.Flags().BoolP(VerifyPublishersFlag, "v", false, VerifyPublishersDesc)
	EvmpushCmd.Flags().IntP(BatchingWindowFlag, "b", 5, BatchingWindowDesc)
	EvmpushCmd.Flags().IntP(PollingPeriodFlag, "p", 3, PollingPeriodDesc)
	EvmpushCmd.Flags().Uint64P(GasLimitFlag, "g", 0, GasLimitDesc)

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
	chainWsUrl, _ := cmd.Flags().GetString(ChainWsUrlFlag)
	contractAddress, _ := cmd.Flags().GetString(ContractAddressFlag)
	assetConfigFile, _ := cmd.Flags().GetString(AssetConfigFileFlag)
	mnemonicFile, _ := cmd.Flags().GetString(MnemonicFileFlag)
	verifyPublishers, _ := cmd.Flags().GetBool(VerifyPublishersFlag)
	batchingWindow, _ := cmd.Flags().GetInt(BatchingWindowFlag)
	pollingPeriod, _ := cmd.Flags().GetInt(PollingPeriodFlag)
	gasLimit, _ := cmd.Flags().GetUint64(GasLimitFlag)

	logger := EvmPusherLogger(chainRpcUrl, contractAddress)

	mnemonic, err := os.ReadFile(mnemonicFile)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to read mnemonic file")
	}

	evmInteractor, err := NewEvmContractInteractor(chainRpcUrl, chainWsUrl, contractAddress, mnemonic, verifyPublishers, logger, gasLimit)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize Evm contract interactor")
	}

	evmPusher := NewPusher(storkWsEndpoint, storkAuth, chainRpcUrl, contractAddress, assetConfigFile, batchingWindow, pollingPeriod, evmInteractor, &logger)
	evmPusher.Run(context.Background())
}
