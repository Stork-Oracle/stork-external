package first_party_evm

import (
	"context"

	"github.com/Stork-Oracle/stork-external/apps/first_party_pusher/pkg/runner"
	"github.com/Stork-Oracle/stork-external/apps/first_party_pusher/pkg/types"
	"github.com/spf13/cobra"
)

func NewPushCmd() *cobra.Command {
	pushCmd := &cobra.Command{
		Use:   "evm",
		Short: "Start EVM First Party Chain Pusher",
		Run:   runPush,
	}

	pushCmd.Flags().StringP(runner.WebsocketPortFlag, "w", "8080", runner.WebsocketPortDesc)
	pushCmd.Flags().StringP(runner.ChainRpcUrlFlag, "c", "", runner.ChainRpcUrlDesc)
	pushCmd.Flags().StringP(runner.ChainWsUrlFlag, "u", "", runner.ChainWsUrlDesc)
	pushCmd.Flags().StringP(runner.ContractAddressFlag, "x", "", runner.ContractAddressDesc)
	pushCmd.Flags().StringP(runner.AssetConfigFileFlag, "f", "", runner.AssetConfigFileDesc)
	pushCmd.Flags().StringP(runner.PrivateKeyFileFlag, "k", "", runner.PrivateKeyFileDesc)
	pushCmd.Flags().Uint64P(runner.GasLimitFlag, "g", 0, runner.GasLimitDesc)

	_ = pushCmd.MarkFlagRequired(runner.ChainRpcUrlFlag)
	_ = pushCmd.MarkFlagRequired(runner.ContractAddressFlag)
	_ = pushCmd.MarkFlagRequired(runner.AssetConfigFileFlag)
	_ = pushCmd.MarkFlagRequired(runner.PrivateKeyFileFlag)

	return pushCmd
}

func runPush(cmd *cobra.Command, args []string) {
	websocketPort, _ := cmd.Flags().GetString(runner.WebsocketPortFlag)
	chainRpcUrl, _ := cmd.Flags().GetString(runner.ChainRpcUrlFlag)
	chainWsUrl, _ := cmd.Flags().GetString(runner.ChainWsUrlFlag)
	contractAddress, _ := cmd.Flags().GetString(runner.ContractAddressFlag)
	assetConfigFile, _ := cmd.Flags().GetString(runner.AssetConfigFileFlag)
	privateKeyFile, _ := cmd.Flags().GetString(runner.PrivateKeyFileFlag)
	gasLimit, _ := cmd.Flags().GetUint64(runner.GasLimitFlag)

	logger := runner.PusherLogger("evm", chainRpcUrl, contractAddress)

	assetConfig, err := runner.LoadAssetConfig(assetConfigFile)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to load asset config")
	}

	privateKey, err := runner.LoadPrivateKey(privateKeyFile)
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
		PrivateKey:      privateKey,
		GasLimit:        gasLimit,
	}

	ctx, cancel := context.WithCancel(context.Background())
	runner := runner.NewFirstPartyRunner(config, interactor, cancel, logger)
	runner.Run(ctx)
}
