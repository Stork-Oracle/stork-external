package cosmwasm

import (
	"context"
	"os"

	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/pusher"
	"github.com/spf13/cobra"
)

func NewPushCmd() *cobra.Command {
	pushCmd := &cobra.Command{
		Use:   "cosmwasm",
		Short: "Push WebSocket prices to Cosmwasm contract",
		Run:   runPush,
	}

	pushCmd.Flags().StringP(pusher.StorkWebsocketEndpointFlag, "w", "", pusher.StorkWebsocketEndpointDesc)
	pushCmd.Flags().StringP(pusher.StorkAuthCredentialsFlag, "a", "", pusher.StorkAuthCredentialsDesc)
	pushCmd.Flags().StringP(pusher.ChainRpcUrlFlag, "r", "", pusher.ChainRpcUrlDesc)
	pushCmd.Flags().StringP(pusher.ContractAddressFlag, "x", "", pusher.ContractAddressDesc)
	pushCmd.Flags().StringP(pusher.AssetConfigFileFlag, "f", "", pusher.AssetConfigFileDesc)
	pushCmd.Flags().StringP(pusher.MnemonicFileFlag, "m", "", pusher.MnemonicFileDesc)
	pushCmd.Flags().IntP(pusher.BatchingWindowFlag, "b", pusher.DefaultBatchingWindow, pusher.BatchingWindowDesc)
	pushCmd.Flags().String(pusher.BatchingWindowStrFlag, "", pusher.BatchingWindowStrDesc)
	pushCmd.Flags().IntP(pusher.PollingPeriodFlag, "p", pusher.DefaultPollingPeriod, pusher.PollingPeriodDesc)
	pushCmd.Flags().Float64P(pusher.GasPriceFlag, "g", 0.0, pusher.GasPriceDesc)
	pushCmd.Flags().Float64P(pusher.GasAdjustmentFlag, "j", 1.0, pusher.GasAdjustmentDesc)
	pushCmd.Flags().StringP(pusher.DenomFlag, "d", "", pusher.DenomDesc)
	pushCmd.Flags().StringP(pusher.ChainIDFlag, "i", "", pusher.ChainIDDesc)
	pushCmd.Flags().StringP(pusher.ChainPrefixFlag, "c", "", pusher.ChainPrefixDesc)

	pushCmd.MarkFlagsMutuallyExclusive(pusher.BatchingWindowFlag, pusher.BatchingWindowStrFlag)

	_ = pushCmd.MarkFlagRequired(pusher.StorkWebsocketEndpointFlag)
	_ = pushCmd.MarkFlagRequired(pusher.StorkAuthCredentialsFlag)
	_ = pushCmd.MarkFlagRequired(pusher.ChainRpcUrlFlag)
	_ = pushCmd.MarkFlagRequired(pusher.ContractAddressFlag)
	_ = pushCmd.MarkFlagRequired(pusher.AssetConfigFileFlag)
	_ = pushCmd.MarkFlagRequired(pusher.MnemonicFileFlag)

	return pushCmd
}

func runPush(cmd *cobra.Command, args []string) {
	storkWsEndpoint, _ := cmd.Flags().GetString(pusher.StorkWebsocketEndpointFlag)
	storkAuth, _ := cmd.Flags().GetString(pusher.StorkAuthCredentialsFlag)
	chainRpcUrl, _ := cmd.Flags().GetString(pusher.ChainRpcUrlFlag)
	contractAddress, _ := cmd.Flags().GetString(pusher.ContractAddressFlag)
	assetConfigFile, _ := cmd.Flags().GetString(pusher.AssetConfigFileFlag)
	mnemonicFile, _ := cmd.Flags().GetString(pusher.MnemonicFileFlag)
	batchingWindow, _ := cmd.Flags().GetInt(pusher.BatchingWindowFlag)
	batchingWindowStr, _ := cmd.Flags().GetString(pusher.BatchingWindowStrFlag)
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

	interactor, err := NewContractInteractor(
		contractAddress,
		mnemonic,
		pollingPeriod,
		logger,
		gasPrice,
		gasAdjustment,
		denom,
		chainID,
		chainPrefix,
	)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize contract interactor")
	}

	pusher := pusher.NewPusher(
		storkWsEndpoint,
		storkAuth,
		chainRpcUrl,
		"",
		contractAddress,
		assetConfigFile,
		batchingWindowStr,
		batchingWindow,
		pollingPeriod,
		interactor,
		&logger,
	)
	pusher.Run(context.Background())
}
