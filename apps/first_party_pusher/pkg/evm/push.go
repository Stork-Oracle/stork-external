package first_party_evm

import (
	"context"

	"github.com/Stork-Oracle/stork-external/apps/first_party_pusher/pkg/pusher"
	"github.com/Stork-Oracle/stork-external/apps/first_party_pusher/pkg/types"
	"github.com/spf13/cobra"
)

func NewPushCmd() *cobra.Command {
	pushCmd := &cobra.Command{
		Use:   "evm",
		Short: "Start EVM First Party Chain Pusher",
		Run:   runPush,
	}

	pushCmd.Flags().StringP(pusher.WebsocketPortFlag, "w", "8080", pusher.WebsocketPortDesc)
	pushCmd.Flags().StringP(pusher.ChainRpcUrlFlag, "c", "", pusher.ChainRpcUrlDesc)
	pushCmd.Flags().StringP(pusher.ChainWsUrlFlag, "u", "", pusher.ChainWsUrlDesc)
	pushCmd.Flags().StringP(pusher.ContractAddressFlag, "x", "", pusher.ContractAddressDesc)
	pushCmd.Flags().StringP(pusher.AssetConfigFileFlag, "f", "", pusher.AssetConfigFileDesc)
	pushCmd.Flags().StringP(pusher.PrivateKeyFileFlag, "k", "", pusher.PrivateKeyFileDesc)
	pushCmd.Flags().IntP(pusher.BatchingWindowFlag, "b", pusher.DefaultBatchingWindow, pusher.BatchingWindowDesc)
	pushCmd.Flags().IntP(pusher.PollingPeriodFlag, "p", pusher.DefaultPollingPeriod, pusher.PollingPeriodDesc)
	pushCmd.Flags().Uint64P(pusher.GasLimitFlag, "g", 0, pusher.GasLimitDesc)

	_ = pushCmd.MarkFlagRequired(pusher.ChainRpcUrlFlag)
	_ = pushCmd.MarkFlagRequired(pusher.ContractAddressFlag)
	_ = pushCmd.MarkFlagRequired(pusher.AssetConfigFileFlag)

	return pushCmd
}

func runPush(cmd *cobra.Command, args []string) {
	websocketPort, _ := cmd.Flags().GetString(pusher.WebsocketPortFlag)
	chainRpcUrl, _ := cmd.Flags().GetString(pusher.ChainRpcUrlFlag)
	chainWsUrl, _ := cmd.Flags().GetString(pusher.ChainWsUrlFlag)
	contractAddress, _ := cmd.Flags().GetString(pusher.ContractAddressFlag)
	assetConfigFile, _ := cmd.Flags().GetString(pusher.AssetConfigFileFlag)
	privateKeyFile, _ := cmd.Flags().GetString(pusher.PrivateKeyFileFlag)
	batchingWindow, _ := cmd.Flags().GetInt(pusher.BatchingWindowFlag)
	pollingPeriod, _ := cmd.Flags().GetInt(pusher.PollingPeriodFlag)
	gasLimit, _ := cmd.Flags().GetUint64(pusher.GasLimitFlag)

	logger := pusher.PusherLogger("evm", chainRpcUrl, contractAddress)

	assetConfig, err := pusher.LoadAssetConfig(assetConfigFile)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to load asset config")
	}

	privateKey, err := pusher.LoadPrivateKey(privateKeyFile)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to load private key")
	}

	interactor, err := NewContractInteractor(
		chainRpcUrl,
		chainWsUrl,
		contractAddress,
		privateKey,
		gasLimit,
		logger,
	)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize contract interactor")
	}
	defer interactor.Close()

	config := &types.FirstPartyConfig{
		WebsocketPort:   websocketPort,
		ChainRpcUrl:     chainRpcUrl,
		ChainWsUrl:      chainWsUrl,
		ContractAddress: contractAddress,
		AssetConfig:     assetConfig,
		GasLimit:        gasLimit,
	}

	ctx, cancel := context.WithCancel(context.Background())
	pusher := pusher.NewFirstPartyRunner(config, interactor, batchingWindow, pollingPeriod, cancel, logger)
	pusher.Run(ctx)
}
