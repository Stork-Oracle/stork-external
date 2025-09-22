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

	pushCmd.Flags().String(runner.WebsocketPortFlag, "8080", runner.WebsocketPortDesc)
	pushCmd.Flags().String(runner.ChainRpcUrlFlag, "", runner.ChainRpcUrlDesc)
	pushCmd.Flags().String(runner.ChainWsUrlFlag, "", runner.ChainWsUrlDesc)
	pushCmd.Flags().String(runner.ContractAddressFlag, "", runner.ContractAddressDesc)
	pushCmd.Flags().String(runner.AssetConfigFileFlag, "", runner.AssetConfigFileDesc)
	pushCmd.Flags().String(runner.PrivateKeyFileFlag, "", runner.PrivateKeyFileDesc)
	pushCmd.Flags().Uint64(runner.GasLimitFlag, 0, runner.GasLimitDesc)

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
