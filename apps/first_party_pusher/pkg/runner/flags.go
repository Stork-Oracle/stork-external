package runner

const (
	WebsocketPortFlag   = "websocket-port"
	ChainRpcUrlFlag     = "chain-rpc-url"
	ChainWsUrlFlag      = "chain-ws-url"
	ContractAddressFlag = "contract-address"
	AssetConfigFileFlag = "asset-config-file"
	PrivateKeyFileFlag  = "private-key-file"
	GasLimitFlag        = "gas-limit"
)

const (
	WebsocketPortDesc   = "WebSocket server port"
	ChainRpcUrlDesc     = "Chain RPC URL"
	ChainWsUrlDesc      = "Chain WebSocket URL"
	ContractAddressDesc = "First Party Stork contract address"
	AssetConfigFileDesc = "Asset configuration file with push settings"
	PrivateKeyFileDesc  = "Private key file for signing transactions"
	GasLimitDesc        = "Gas limit for transactions (0 to use estimate)"
)
