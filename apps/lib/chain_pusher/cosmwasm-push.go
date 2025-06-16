package chain_pusher

import (
	"context"
	"os"

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
	CosmwasmPushCmd.Flags().StringP(ChainRpcUrlFlag, "r", "", ChainRpcUrlDesc)
	CosmwasmPushCmd.Flags().StringP(ContractAddressFlag, "x", "", ContractAddressDesc)
	CosmwasmPushCmd.Flags().StringP(AssetConfigFileFlag, "f", "", AssetConfigFileDesc)
	CosmwasmPushCmd.Flags().StringP(MnemonicFileFlag, "m", "", MnemonicFileDesc)
	CosmwasmPushCmd.Flags().IntP(BatchingWindowFlag, "b", 5, BatchingWindowDesc)
	CosmwasmPushCmd.Flags().IntP(PollingPeriodFlag, "p", 3, PollingPeriodDesc)
	CosmwasmPushCmd.Flags().Float64P(GasPriceFlag, "g", 0.0, GasPriceDesc)
	CosmwasmPushCmd.Flags().Float64P(GasAdjustmentFlag, "j", 1.0, GasAdjustmentDesc)
	CosmwasmPushCmd.Flags().StringP(DenomFlag, "d", "", DenomDesc)
	CosmwasmPushCmd.Flags().StringP(ChainIDFlag, "i", "", ChainIDDesc)
	CosmwasmPushCmd.Flags().StringP(ChainPrefixFlag, "c", "", ChainPrefixDesc)
	CosmwasmPushCmd.MarkFlagRequired(StorkWebsocketEndpointFlag)
	CosmwasmPushCmd.MarkFlagRequired(StorkAuthCredentialsFlag)
	CosmwasmPushCmd.MarkFlagRequired(ChainRpcUrlFlag)
	CosmwasmPushCmd.MarkFlagRequired(ContractAddressFlag)
	CosmwasmPushCmd.MarkFlagRequired(AssetConfigFileFlag)
	CosmwasmPushCmd.MarkFlagRequired(MnemonicFileFlag)
}

func runCosmwasmPush(cmd *cobra.Command, args []string) {
	storkWsEndpoint, _ := cmd.Flags().GetString(StorkWebsocketEndpointFlag)
	storkAuth, _ := cmd.Flags().GetString(StorkAuthCredentialsFlag)
	chainRpcUrl, _ := cmd.Flags().GetString(ChainRpcUrlFlag)
	contractAddress, _ := cmd.Flags().GetString(ContractAddressFlag)
	assetConfigFile, _ := cmd.Flags().GetString(AssetConfigFileFlag)
	mnemonicFile, _ := cmd.Flags().GetString(MnemonicFileFlag)
	batchingWindow, _ := cmd.Flags().GetInt(BatchingWindowFlag)
	pollingPeriod, _ := cmd.Flags().GetInt(PollingPeriodFlag)

	gasPrice, _ := cmd.Flags().GetFloat64(GasPriceFlag)
	gasAdjustment, _ := cmd.Flags().GetFloat64(GasAdjustmentFlag)
	denom, _ := cmd.Flags().GetString(DenomFlag)
	chainID, _ := cmd.Flags().GetString(ChainIDFlag)
	chainPrefix, _ := cmd.Flags().GetString(ChainPrefixFlag)
	logger := CosmwasmPusherLogger(chainRpcUrl, contractAddress)

	mnemonic, err := os.ReadFile(mnemonicFile)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to read mnemonic file")
	}

	cosmwasmInteractor, err := NewCosmwasmContractInteractor(chainRpcUrl, contractAddress, mnemonic, batchingWindow, pollingPeriod, logger, gasPrice, gasAdjustment, denom, chainID, chainPrefix)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create cosmwasm interactor")
	}

	cosmwasmPusher := NewPusher(storkWsEndpoint, storkAuth, chainRpcUrl, contractAddress, assetConfigFile, batchingWindow, pollingPeriod, cosmwasmInteractor, &logger)
	cosmwasmPusher.Run(context.Background())
}
