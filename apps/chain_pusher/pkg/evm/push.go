package evm

import (
	"context"
	"os"

	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/pusher"
	"github.com/spf13/cobra"
)

var EvmpushCmd = &cobra.Command{
	Use:   "evm",
	Short: "Push WebSocket prices to EVM contract",
	Run:   runEvmPush,
}

func init() {
	EvmpushCmd.Flags().StringP(pusher.StorkWebsocketEndpointFlag, "w", "", pusher.StorkWebsocketEndpointDesc)
	EvmpushCmd.Flags().StringP(pusher.StorkAuthCredentialsFlag, "a", "", pusher.StorkAuthCredentialsDesc)
	EvmpushCmd.Flags().StringP(pusher.ChainRpcUrlFlag, "c", "", pusher.ChainRpcUrlDesc)
	EvmpushCmd.Flags().StringP(pusher.ChainWsUrlFlag, "u", "", pusher.ChainWsUrlDesc)
	EvmpushCmd.Flags().StringP(pusher.ContractAddressFlag, "x", "", pusher.ContractAddressDesc)
	EvmpushCmd.Flags().StringP(pusher.AssetConfigFileFlag, "f", "", pusher.AssetConfigFileDesc)
	EvmpushCmd.Flags().StringP(pusher.MnemonicFileFlag, "m", "", pusher.MnemonicFileDesc)
	EvmpushCmd.Flags().BoolP(pusher.VerifyPublishersFlag, "v", false, pusher.VerifyPublishersDesc)
	EvmpushCmd.Flags().IntP(pusher.BatchingWindowFlag, "b", 5, pusher.BatchingWindowDesc)
	EvmpushCmd.Flags().IntP(pusher.PollingPeriodFlag, "p", 3, pusher.PollingPeriodDesc)
	EvmpushCmd.Flags().Uint64P(pusher.GasLimitFlag, "g", 0, pusher.GasLimitDesc)

	EvmpushCmd.MarkFlagRequired(pusher.StorkWebsocketEndpointFlag)
	EvmpushCmd.MarkFlagRequired(pusher.StorkAuthCredentialsFlag)
	EvmpushCmd.MarkFlagRequired(pusher.ChainRpcUrlFlag)
	EvmpushCmd.MarkFlagRequired(pusher.ContractAddressFlag)
	EvmpushCmd.MarkFlagRequired(pusher.AssetConfigFileFlag)
	EvmpushCmd.MarkFlagRequired(pusher.MnemonicFileFlag)
}

func runEvmPush(cmd *cobra.Command, args []string) {
	storkWsEndpoint, _ := cmd.Flags().GetString(pusher.StorkWebsocketEndpointFlag)
	storkAuth, _ := cmd.Flags().GetString(pusher.StorkAuthCredentialsFlag)
	chainRpcUrl, _ := cmd.Flags().GetString(pusher.ChainRpcUrlFlag)
	chainWsUrl, _ := cmd.Flags().GetString(pusher.ChainWsUrlFlag)
	contractAddress, _ := cmd.Flags().GetString(pusher.ContractAddressFlag)
	assetConfigFile, _ := cmd.Flags().GetString(pusher.AssetConfigFileFlag)
	mnemonicFile, _ := cmd.Flags().GetString(pusher.MnemonicFileFlag)
	verifyPublishers, _ := cmd.Flags().GetBool(pusher.VerifyPublishersFlag)
	batchingWindow, _ := cmd.Flags().GetInt(pusher.BatchingWindowFlag)
	pollingPeriod, _ := cmd.Flags().GetInt(pusher.PollingPeriodFlag)
	gasLimit, _ := cmd.Flags().GetUint64(pusher.GasLimitFlag)

	logger := pusher.EvmPusherLogger(chainRpcUrl, contractAddress)

	mnemonic, err := os.ReadFile(mnemonicFile)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to read mnemonic file")
	}

	evmInteractor, err := NewEvmContractInteractor(chainRpcUrl, chainWsUrl, contractAddress, mnemonic, verifyPublishers, logger, gasLimit)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize Evm contract interactor")
	}

	evmPusher := pusher.NewPusher(storkWsEndpoint, storkAuth, chainRpcUrl, contractAddress, assetConfigFile, batchingWindow, pollingPeriod, evmInteractor, &logger)
	evmPusher.Run(context.Background())
}
