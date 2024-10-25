package chain_pusher

const StorkWebsocketEndpointFlag = "stork-ws-endpoint"
const StorkAuthCredentialsFlag = "stork-auth-credentials"
const ChainRpcUrlFlag = "chain-rpc-url"
const ChainWsUrlFlag = "chain-ws-url"
const ContractAddressFlag = "contract-address"
const AssetConfigFileFlag = "asset-config-file"
const MnemonicFileFlag = "mnemonic-file"
const PrivateKeyFileFlag = "private-key-file"

const VerifyPublishersFlag = "verify-publishers"
const BatchingWindowFlag = "batching-window"
const PollingFrequencyFlag = "polling-frequency"

// Descriptions for the flags
const StorkWebsocketEndpointDesc = "Stork WebSocket endpoint"
const StorkAuthCredentialsDesc = "Stork auth credentials - base64(username:password)"
const ChainRpcUrlDesc = "Chain RPC URL"
const ChainWsUrlDesc = "Chain WebSocket URL"
const ContractAddressDesc = "Contract address"
const AssetConfigFileDesc = "Asset config file"
const MnemonicFileDesc = "Mnemonic file"
const PrivateKeyFileDesc = "Private key file"
const VerifyPublishersDesc = "Verify the publisher signed prices before pushing stork signed value to contract"
const BatchingWindowDesc = "Batching window (seconds)"
const PollingFrequencyDesc = "Asset Polling frequency (seconds)"
