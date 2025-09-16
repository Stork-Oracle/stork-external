# Self-Serve Chain Pusher

The Self-Serve Chain Pusher is a standalone application that receives price update messages from publisher agents via WebSocket and pushes them to self-serve Stork contracts based on configurable cadence and delta tolerance settings.

## Features

- **WebSocket Server**: Receives price updates from publisher agents in the same format as the main Stork system
- **Configurable Push Logic**: Supports both time-based (cadence) and price-change-based (delta tolerance) push triggers
- **Chain Support**: Currently supports EVM-based chains with the self-serve Stork contract
- **Rate Limiting**: Built-in rate limiting for blockchain RPC calls
- **Robust Error Handling**: Retry logic and comprehensive error handling for contract interactions

## Usage

### Build and Run

```bash
# Build the application
go build -o self-serve-chain-pusher ./cmd

# Run with EVM chain
./self-serve-chain-pusher evm \
  --chain-rpc-url https://your-rpc-endpoint.com \
  --contract-address 0x1234567890abcdef1234567890abcdef12345678 \
  --asset-config-file sample.asset-config.yaml \
  --private-key-file private-key.txt \
  --websocket-port 8080
```

### Configuration Files

#### Asset Configuration (`sample.asset-config.yaml`)

```yaml
assets:
  BTCUSD:
    asset_id: BTCUSD
    encoded_asset_id: 0x7404e3d104ea7841c3d9e6fd20adfe99b4ad586bc08d8f3bd3afef894cf184de
    push_interval_sec: 60      # Push at least every 60 seconds
    percent_change_threshold: 1.0  # Push if price changes by 1% or more
  ETHUSD:
    asset_id: ETHUSD
    encoded_asset_id: 0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef
    push_interval_sec: 30
    percent_change_threshold: 0.5
```

#### Private Key File

Create a file containing your private key in hex format:

```
0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef12
```

Or without the `0x` prefix:

```
1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef12
```

### Command Line Options

#### EVM Command

- `--websocket-port`: WebSocket server port (default: 8080)
- `--chain-rpc-url`: Chain RPC URL (required)
- `--chain-ws-url`: Chain WebSocket URL (optional)
- `--contract-address`: Self-serve Stork contract address (required)
- `--asset-config-file`: Asset configuration file path (required)
- `--private-key-file`: Private key file path (required)
- `--gas-limit`: Gas limit for transactions (0 to use estimate)
- `--limit-per-second`: JSON RPC call limit per second (default: 10.0)
- `--burst-limit`: JSON RPC call burst limit (default: 20)

### WebSocket API

The application exposes a WebSocket endpoint at `/ws` that accepts price update messages in the same format as the publisher agent:

```json
{
  "type": "prices",
  "data": [
    {
      "t": 1640995200000000000,
      "a": "BTCUSD",
      "v": "50000.123456",
      "m": {}
    }
  ]
}
```

Where:
- `t`: Publish timestamp in nanoseconds
- `a`: Asset identifier
- `v`: Price value (can be string or number)
- `m`: Optional metadata object

### Push Logic

The application pushes values to the contract when either condition is met:

1. **Time Trigger**: The configured `push_interval_sec` has elapsed since the last push
2. **Delta Trigger**: The price has changed by more than the configured `percent_change_threshold`

### Health Check

A health check endpoint is available at `/health` that returns `200 OK`.

## Architecture

The application consists of several key components:

1. **WebSocket Server**: Handles incoming connections and price update messages
2. **Value Processor**: Processes incoming price updates and determines when to trigger pushes
3. **Contract Interactor**: Manages blockchain interactions and transaction submission
4. **Asset State Manager**: Tracks the current state of each configured asset

## Development

### Adding New Chain Support

To add support for additional chains:

1. Create a new command file (e.g., `solana_command.go`)
2. Implement the chain-specific contract interactor
3. Add the command to the main CLI in `cmd/main.go`

### Testing

The application includes comprehensive logging to help with debugging and monitoring. Set the `--verbose` flag for detailed logs.

## Security Considerations

- Store private keys securely and never commit them to version control
- Use appropriate gas limits to prevent excessive transaction costs
- Monitor the application logs for any unusual behavior
- Consider running the application behind a firewall or VPN for production use

### Generate contract bindings for EVM

From the root of the repo, run:

```bash
abigen --abi <(jq -r '.abi' ./chains/evm/contracts/self_serve_stork/artifacts/contracts/SelfServeStork.sol/SelfServeStork.json) --pkg contract_bindings_evm --type SelfServeStorkContract --out ./apps/self_serve_chain_pusher/lib/contract_bindings/evm/stork_evm_contract.go
```
