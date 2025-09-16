package self_serve_evm

import (
	"github.com/spf13/cobra"
	"github.com/Stork-Oracle/stork-external/apps/self_serve_chain_pusher/pkg/pusher"
)

var EvmSelfServeCmd = &cobra.Command{
	Use:   "evm",
	Short: "Start EVM self-serve chain pusher",
	Run: func(cmd *cobra.Command, args []string) {
		config, err := buildEvmConfigFromFlags(cmd)
		if err != nil {
			panic(err)
		}

		runner := NewEvmSelfServeRunner(config)
		runner.Run()
	},
}

func init() {
	addEvmFlags(EvmSelfServeCmd)
}

func addEvmFlags(cmd *cobra.Command) {
	cmd.Flags().String(pusher.WebsocketPortFlag, "8080", pusher.WebsocketPortDesc)
	cmd.Flags().String(pusher.ChainRpcUrlFlag, "", pusher.ChainRpcUrlDesc)
	cmd.Flags().String(pusher.ChainWsUrlFlag, "", pusher.ChainWsUrlDesc)
	cmd.Flags().String(pusher.ContractAddressFlag, "", pusher.ContractAddressDesc)
	cmd.Flags().String(pusher.AssetConfigFileFlag, "", pusher.AssetConfigFileDesc)
	cmd.Flags().String(pusher.PrivateKeyFileFlag, "", pusher.PrivateKeyFileDesc)
	cmd.Flags().Uint64(pusher.GasLimitFlag, 0, pusher.GasLimitDesc)

	cmd.MarkFlagRequired(pusher.ChainRpcUrlFlag)
	cmd.MarkFlagRequired(pusher.ContractAddressFlag)
	cmd.MarkFlagRequired(pusher.AssetConfigFileFlag)
	cmd.MarkFlagRequired(pusher.PrivateKeyFileFlag)
}

func buildEvmConfigFromFlags(cmd *cobra.Command) (*EvmSelfServeConfig, error) {
	websocketPort, _ := cmd.Flags().GetString(pusher.WebsocketPortFlag)
	chainRpcUrl, _ := cmd.Flags().GetString(pusher.ChainRpcUrlFlag)
	chainWsUrl, _ := cmd.Flags().GetString(pusher.ChainWsUrlFlag)
	contractAddress, _ := cmd.Flags().GetString(pusher.ContractAddressFlag)
	assetConfigFile, _ := cmd.Flags().GetString(pusher.AssetConfigFileFlag)
	privateKeyFile, _ := cmd.Flags().GetString(pusher.PrivateKeyFileFlag)
	gasLimit, _ := cmd.Flags().GetUint64(pusher.GasLimitFlag)

	assetConfig, err := LoadAssetConfig(assetConfigFile)
	if err != nil {
		return nil, err
	}

	privateKey, err := LoadPrivateKey(privateKeyFile)
	if err != nil {
		return nil, err
	}

	return &EvmSelfServeConfig{
		WebsocketPort:   websocketPort,
		ChainRpcUrl:     chainRpcUrl,
		ChainWsUrl:      chainWsUrl,
		ContractAddress: contractAddress,
		AssetConfig:     assetConfig,
		PrivateKey:      privateKey,
		GasLimit:        gasLimit,
	}, nil
}
