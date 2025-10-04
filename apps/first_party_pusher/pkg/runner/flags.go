package runner

const (
	DefaultBatchingWindow = 5
	DefaultPollingPeriod  = 3
)

const (
	WebsocketPortFlag   = "websocket-port"
	ChainRpcUrlFlag     = "chain-rpc-url"
	ChainWsUrlFlag      = "chain-ws-url"
	ContractAddressFlag = "contract-address"
	AssetConfigFileFlag = "asset-config-file"
	PrivateKeyFileFlag  = "private-key-file"
	BatchingWindowFlag  = "batching-window"
	PollingPeriodFlag   = "polling-period"
	GasLimitFlag        = "gas-limit"
)

const (
	WebsocketPortDesc   = "WebSocket server port"
	ChainRpcUrlDesc     = "Chain RPC URL"
	ChainWsUrlDesc      = "Chain WebSocket URL"
	ContractAddressDesc = "First Party Stork contract address"
	AssetConfigFileDesc = "Asset configuration file with push settings"
	PrivateKeyFileDesc  = "Private key file for signing transactions"
	BatchingWindowDesc  = "Batching window (seconds)"
	PollingPeriodDesc   = "Polling period (seconds)"
	GasLimitDesc        = "Gas limit for transactions (0 to use estimate)"
)
