package cosmwasm

import (
	"context"
	"os"

	pusher "github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/pusher"
	"github.com/spf13/cobra"
)

var CosmwasmPushCmd = &cobra.Command{
	Use:   "cosmwasm",
	Short: "Push WebSocket prices to Cosmwasm contract",
	Run:   runCosmwasmPush,
}

func init() {
	CosmwasmPushCmd.Flags().StringP(pusher.StorkWebsocketEndpointFlag, "w", "", pusher.StorkWebsocketEndpointDesc)
	CosmwasmPushCmd.Flags().StringP(pusher.StorkAuthCredentialsFlag, "a", "", pusher.StorkAuthCredentialsDesc)
	CosmwasmPushCmd.Flags().StringP(pusher.ChainRpcUrlFlag, "r", "", pusher.ChainRpcUrlDesc)
	CosmwasmPushCmd.Flags().StringP(pusher.ContractAddressFlag, "x", "", pusher.ContractAddressDesc)
	CosmwasmPushCmd.Flags().StringP(pusher.AssetConfigFileFlag, "f", "", pusher.AssetConfigFileDesc)
	CosmwasmPushCmd.Flags().StringP(pusher.MnemonicFileFlag, "m", "", pusher.MnemonicFileDesc)
	CosmwasmPushCmd.Flags().IntP(pusher.BatchingWindowFlag, "b", 5, pusher.BatchingWindowDesc)
	CosmwasmPushCmd.Flags().IntP(pusher.PollingPeriodFlag, "p", 3, pusher.PollingPeriodDesc)
	CosmwasmPushCmd.Flags().Float64P(pusher.GasPriceFlag, "g", 0.0, pusher.GasPriceDesc)
	CosmwasmPushCmd.Flags().Float64P(pusher.GasAdjustmentFlag, "j", 1.0, pusher.GasAdjustmentDesc)
	CosmwasmPushCmd.Flags().StringP(pusher.DenomFlag, "d", "", pusher.DenomDesc)
	CosmwasmPushCmd.Flags().StringP(pusher.ChainIDFlag, "i", "", pusher.ChainIDDesc)
	CosmwasmPushCmd.Flags().StringP(pusher.ChainPrefixFlag, "c", "", pusher.ChainPrefixDesc)
	CosmwasmPushCmd.MarkFlagRequired(pusher.StorkWebsocketEndpointFlag)
	CosmwasmPushCmd.MarkFlagRequired(pusher.StorkAuthCredentialsFlag)
	CosmwasmPushCmd.MarkFlagRequired(pusher.ChainRpcUrlFlag)
	CosmwasmPushCmd.MarkFlagRequired(pusher.ContractAddressFlag)
	CosmwasmPushCmd.MarkFlagRequired(pusher.AssetConfigFileFlag)
	CosmwasmPushCmd.MarkFlagRequired(pusher.MnemonicFileFlag)
}

func runCosmwasmPush(cmd *cobra.Command, args []string) {
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
	logger := pusher.CosmwasmPusherLogger(chainRpcUrl, contractAddress)

	mnemonic, err := os.ReadFile(mnemonicFile)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to read mnemonic file")
	}

	cosmwasmInteractor, err := NewCosmwasmContractInteractor(chainRpcUrl, contractAddress, mnemonic, batchingWindow, pollingPeriod, logger, gasPrice, gasAdjustment, denom, chainID, chainPrefix)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create cosmwasm interactor")
	}

	cosmwasmPusher := pusher.NewPusher(storkWsEndpoint, storkAuth, chainRpcUrl, contractAddress, assetConfigFile, batchingWindow, pollingPeriod, cosmwasmInteractor, &logger)
	cosmwasmPusher.Run(context.Background())
}
