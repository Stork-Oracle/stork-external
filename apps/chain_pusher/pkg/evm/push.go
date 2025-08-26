package evm

import (
	"context"
	"os"

	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/pusher"
	"github.com/spf13/cobra"
)

var PushCmd = &cobra.Command{
	Use:   "evm",
	Short: "Push WebSocket prices to EVM contract",
	Run:   runPush,
}

func init() {
	PushCmd.Flags().StringP(pusher.StorkWebsocketEndpointFlag, "w", "", pusher.StorkWebsocketEndpointDesc)
	PushCmd.Flags().StringP(pusher.StorkAuthCredentialsFlag, "a", "", pusher.StorkAuthCredentialsDesc)
	PushCmd.Flags().StringP(pusher.ChainRpcUrlFlag, "c", "", pusher.ChainRpcUrlDesc)
	PushCmd.Flags().StringP(pusher.ChainWsUrlFlag, "u", "", pusher.ChainWsUrlDesc)
	PushCmd.Flags().StringP(pusher.ContractAddressFlag, "x", "", pusher.ContractAddressDesc)
	PushCmd.Flags().StringP(pusher.AssetConfigFileFlag, "f", "", pusher.AssetConfigFileDesc)
	PushCmd.Flags().StringP(pusher.MnemonicFileFlag, "m", "", pusher.MnemonicFileDesc)
	PushCmd.Flags().BoolP(pusher.VerifyPublishersFlag, "v", false, pusher.VerifyPublishersDesc)
	PushCmd.Flags().IntP(pusher.BatchingWindowFlag, "b", 5, pusher.BatchingWindowDesc)
	PushCmd.Flags().IntP(pusher.PollingPeriodFlag, "p", 3, pusher.PollingPeriodDesc)
	PushCmd.Flags().Uint64P(pusher.GasLimitFlag, "g", 0, pusher.GasLimitDesc)

	PushCmd.MarkFlagRequired(pusher.StorkWebsocketEndpointFlag)
	PushCmd.MarkFlagRequired(pusher.StorkAuthCredentialsFlag)
	PushCmd.MarkFlagRequired(pusher.ChainRpcUrlFlag)
	PushCmd.MarkFlagRequired(pusher.ContractAddressFlag)
	PushCmd.MarkFlagRequired(pusher.AssetConfigFileFlag)
	PushCmd.MarkFlagRequired(pusher.MnemonicFileFlag)
}

func runPush(cmd *cobra.Command, args []string) {
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

	logger := PusherLogger(chainRpcUrl, contractAddress)

	mnemonic, err := os.ReadFile(mnemonicFile)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to read mnemonic file")
	}

	interactor, err := NewContractInteractor(chainRpcUrl, chainWsUrl, contractAddress, mnemonic, verifyPublishers, logger, gasLimit)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize contract interactor")
	}

	pusher := pusher.NewPusher(storkWsEndpoint, storkAuth, chainRpcUrl, contractAddress, assetConfigFile, batchingWindow, pollingPeriod, interactor, &logger)
	pusher.Run(context.Background())
}
