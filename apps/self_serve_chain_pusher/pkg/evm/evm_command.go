package self_serve_evm

import (
	"context"

	"github.com/Stork-Oracle/stork-external/apps/self_serve_chain_pusher/pkg/pusher"
	"github.com/spf13/cobra"
)

func EvmSelfServeCmd() *cobra.Command {
	evmSelfServeCmd := &cobra.Command{
		Use:   "evm",
		Short: "Start EVM self-serve chain pusher",
		Run:   runPush,
	}

	evmSelfServeCmd.Flags().String(pusher.WebsocketPortFlag, "8080", pusher.WebsocketPortDesc)
	evmSelfServeCmd.Flags().String(pusher.ChainRpcUrlFlag, "", pusher.ChainRpcUrlDesc)
	evmSelfServeCmd.Flags().String(pusher.ChainWsUrlFlag, "", pusher.ChainWsUrlDesc)
	evmSelfServeCmd.Flags().String(pusher.ContractAddressFlag, "", pusher.ContractAddressDesc)
	evmSelfServeCmd.Flags().String(pusher.AssetConfigFileFlag, "", pusher.AssetConfigFileDesc)
	evmSelfServeCmd.Flags().String(pusher.PrivateKeyFileFlag, "", pusher.PrivateKeyFileDesc)
	evmSelfServeCmd.Flags().Uint64(pusher.GasLimitFlag, 0, pusher.GasLimitDesc)

	_ = evmSelfServeCmd.MarkFlagRequired(pusher.ChainRpcUrlFlag)
	_ = evmSelfServeCmd.MarkFlagRequired(pusher.ContractAddressFlag)
	_ = evmSelfServeCmd.MarkFlagRequired(pusher.AssetConfigFileFlag)
	_ = evmSelfServeCmd.MarkFlagRequired(pusher.PrivateKeyFileFlag)

	return evmSelfServeCmd
}

func runPush(cmd *cobra.Command, args []string) {
	websocketPort, _ := cmd.Flags().GetString(pusher.WebsocketPortFlag)
	chainRpcUrl, _ := cmd.Flags().GetString(pusher.ChainRpcUrlFlag)
	chainWsUrl, _ := cmd.Flags().GetString(pusher.ChainWsUrlFlag)
	contractAddress, _ := cmd.Flags().GetString(pusher.ContractAddressFlag)
	assetConfigFile, _ := cmd.Flags().GetString(pusher.AssetConfigFileFlag)
	privateKeyFile, _ := cmd.Flags().GetString(pusher.PrivateKeyFileFlag)
	gasLimit, _ := cmd.Flags().GetUint64(pusher.GasLimitFlag)

	logger := pusher.PusherLogger("evm", chainRpcUrl, contractAddress)

	assetConfig, err := LoadAssetConfig(assetConfigFile)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to load asset config")
	}

	privateKey, err := LoadPrivateKey(privateKeyFile)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to load private key")
	}

	config := &EvmSelfServeConfig{
		WebsocketPort:   websocketPort,
		ChainRpcUrl:     chainRpcUrl,
		ChainWsUrl:      chainWsUrl,
		ContractAddress: contractAddress,
		AssetConfig:     assetConfig,
		PrivateKey:      privateKey,
		GasLimit:        gasLimit,
	}

	ctx, cancel := context.WithCancel(context.Background())
	runner := NewEvmSelfServeRunner(config, cancel, logger)
	runner.Run(ctx)
}
