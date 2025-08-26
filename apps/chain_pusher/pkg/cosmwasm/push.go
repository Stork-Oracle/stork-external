package cosmwasm

import (
	"context"
	"os"

	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/pusher"
	"github.com/spf13/cobra"
)

var PushCmd = &cobra.Command{
	Use:   "cosmwasm",
	Short: "Push WebSocket prices to Cosmwasm contract",
	Run:   runPush,
}

func init() {
	PushCmd.Flags().StringP(pusher.StorkWebsocketEndpointFlag, "w", "", pusher.StorkWebsocketEndpointDesc)
	PushCmd.Flags().StringP(pusher.StorkAuthCredentialsFlag, "a", "", pusher.StorkAuthCredentialsDesc)
	PushCmd.Flags().StringP(pusher.ChainRpcUrlFlag, "r", "", pusher.ChainRpcUrlDesc)
	PushCmd.Flags().StringP(pusher.ContractAddressFlag, "x", "", pusher.ContractAddressDesc)
	PushCmd.Flags().StringP(pusher.AssetConfigFileFlag, "f", "", pusher.AssetConfigFileDesc)
	PushCmd.Flags().StringP(pusher.MnemonicFileFlag, "m", "", pusher.MnemonicFileDesc)
	PushCmd.Flags().IntP(pusher.BatchingWindowFlag, "b", 5, pusher.BatchingWindowDesc)
	PushCmd.Flags().IntP(pusher.PollingPeriodFlag, "p", 3, pusher.PollingPeriodDesc)
	PushCmd.Flags().Float64P(pusher.GasPriceFlag, "g", 0.0, pusher.GasPriceDesc)
	PushCmd.Flags().Float64P(pusher.GasAdjustmentFlag, "j", 1.0, pusher.GasAdjustmentDesc)
	PushCmd.Flags().StringP(pusher.DenomFlag, "d", "", pusher.DenomDesc)
	PushCmd.Flags().StringP(pusher.ChainIDFlag, "i", "", pusher.ChainIDDesc)
	PushCmd.Flags().StringP(pusher.ChainPrefixFlag, "c", "", pusher.ChainPrefixDesc)
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
	contractAddress, _ := cmd.Flags().GetString(pusher.ContractAddressFlag)
	assetConfigFile, _ := cmd.Flags().GetString(pusher.AssetConfigFileFlag)
	mnemonicFile, _ := cmd.Flags().GetString(pusher.MnemonicFileFlag)
	batchingWindow, _ := cmd.Flags().GetInt(pusher.BatchingWindowFlag)
	pollingPeriod, _ := cmd.Flags().GetInt(pusher.PollingPeriodFlag)

	gasPrice, _ := cmd.Flags().GetFloat64(pusher.GasPriceFlag)
	gasAdjustment, _ := cmd.Flags().GetFloat64(pusher.GasAdjustmentFlag)
	denom, _ := cmd.Flags().GetString(pusher.DenomFlag)
	chainID, _ := cmd.Flags().GetString(pusher.ChainIDFlag)
	chainPrefix, _ := cmd.Flags().GetString(pusher.ChainPrefixFlag)
	logger := PusherLogger(chainRpcUrl, contractAddress)

	mnemonic, err := os.ReadFile(mnemonicFile)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to read mnemonic file")
	}

	interactor, err := NewContractInteractor(chainRpcUrl, contractAddress, mnemonic, batchingWindow, pollingPeriod, logger, gasPrice, gasAdjustment, denom, chainID, chainPrefix)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize contract interactor")
	}

	pusher := pusher.NewPusher(storkWsEndpoint, storkAuth, chainRpcUrl, contractAddress, assetConfigFile, batchingWindow, pollingPeriod, interactor, &logger)
	pusher.Run(context.Background())
}
