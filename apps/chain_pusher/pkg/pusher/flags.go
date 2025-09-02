package pusher

const (
	StorkWebsocketEndpointFlag = "stork-ws-endpoint"
	//nolint:gosec // This is not a credential
	StorkAuthCredentialsFlag = "stork-auth-credentials"
	ChainRpcUrlFlag          = "chain-rpc-url"
	ChainGrpcUrlFlag         = "chain-grpc-url"
	ChainWsUrlFlag           = "chain-ws-url"
	ContractAddressFlag      = "contract-address"
	AssetConfigFileFlag      = "asset-config-file"
	MnemonicFileFlag         = "mnemonic-file"
	PrivateKeyFileFlag       = "private-key-file"
)

const (
	VerifyPublishersFlag = "verify-publishers"
	BatchingWindowFlag   = "batching-window"
	PollingPeriodFlag    = "polling-period"
	LimitPerSecondFlag   = "limit-per-second"
	BurstLimitFlag       = "burst-limit"
	BatchSizeFlag        = "batch-size"
	GasLimitFlag         = "gas-limit"
)

// Cosmwasm flags.
const (
	GasPriceFlag      = "gas-price"
	GasAdjustmentFlag = "gas-adjustment"
	DenomFlag         = "denom"
	ChainIDFlag       = "chain-id"
	ChainPrefixFlag   = "chain-prefix"
)

// Descriptions for the flags.
const (
	StorkWebsocketEndpointDesc = "Stork WebSocket endpoint"
	//nolint:gosec // This is an example of credentials, not actual credentials
	StorkAuthCredentialsDesc = "Stork auth credentials - base64(username:password)"
	ChainRpcUrlDesc          = "Chain RPC URL"
	ChainGrpcUrlDesc         = "Chain gRPC URL"
	ChainWsUrlDesc           = "Chain WebSocket URL"
	ContractAddressDesc      = "Contract address"
	AssetConfigFileDesc      = "Asset config file"
	MnemonicFileDesc         = "Mnemonic file"
	PrivateKeyFileDesc       = "Private key file"
	VerifyPublishersDesc     = "Verify the publisher signed prices before pushing stork signed value to contract"
	BatchingWindowDesc       = "Batching window (seconds)"
	PollingPeriodDesc        = "Asset Polling Period (seconds)"
	LimitPerSecondDesc       = "JSON RPC call limit per second"
	BurstLimitDesc           = "JSON RPC call Burst limit"
	BatchSizeDesc            = "Batch size between 1 and 4"
	GasLimitDesc             = "Gas limit (0 to use estimate)"
)

// Cosmwasm descriptions.
const (
	GasPriceDesc      = "Gas price"
	GasAdjustmentDesc = "Gas adjustment"
	DenomDesc         = "Denom"
	ChainIDDesc       = "Chain ID"
	ChainPrefixDesc   = "Chain prefix"
)
