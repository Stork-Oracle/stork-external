package self_serve_chain_pusher

import (
	"github.com/spf13/cobra"
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
	cmd.Flags().String(WebsocketPortFlag, "8080", WebsocketPortDesc)
	cmd.Flags().String(ChainRpcUrlFlag, "", ChainRpcUrlDesc)
	cmd.Flags().String(ChainWsUrlFlag, "", ChainWsUrlDesc)
	cmd.Flags().String(ContractAddressFlag, "", ContractAddressDesc)
	cmd.Flags().String(AssetConfigFileFlag, "", AssetConfigFileDesc)
	cmd.Flags().String(PrivateKeyFileFlag, "", PrivateKeyFileDesc)
	cmd.Flags().Uint64(GasLimitFlag, 0, GasLimitDesc)

	cmd.MarkFlagRequired(ChainRpcUrlFlag)
	cmd.MarkFlagRequired(ContractAddressFlag)
	cmd.MarkFlagRequired(AssetConfigFileFlag)
	cmd.MarkFlagRequired(PrivateKeyFileFlag)
}

func buildEvmConfigFromFlags(cmd *cobra.Command) (*EvmSelfServeConfig, error) {
	websocketPort, _ := cmd.Flags().GetString(WebsocketPortFlag)
	chainRpcUrl, _ := cmd.Flags().GetString(ChainRpcUrlFlag)
	chainWsUrl, _ := cmd.Flags().GetString(ChainWsUrlFlag)
	contractAddress, _ := cmd.Flags().GetString(ContractAddressFlag)
	assetConfigFile, _ := cmd.Flags().GetString(AssetConfigFileFlag)
	privateKeyFile, _ := cmd.Flags().GetString(PrivateKeyFileFlag)
	gasLimit, _ := cmd.Flags().GetUint64(GasLimitFlag)

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
