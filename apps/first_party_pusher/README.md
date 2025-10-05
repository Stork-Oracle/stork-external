# First Party Chain Pusher

The First Party Chain Pusher is a standalone application that receives price update messages from publisher agents via WebSocket and pushes them to First Party Stork contracts based on configurable cadence and delta tolerance settings.

**Note: Unless otherwise specified, all commands in this README are in the context of the current directory.**

```bash
# from the root of this repo:
cd apps/first_party_pusher
```

## Configuration

Create an `asset-config.yaml` file. This file should be structured as follows:

```yaml
assets:
  BTCUSD:
    # The asset's symbol, used to set which incoming assets to expect from the websocket server
    asset_id: BTCUSD
    # The publisher's public key that signs the price updates
    public_key: 0x99e295e85cb07C16B7BB62A44dF532A7F2620237
    # If the data feed is not updated within this period the asset should be added to the batched updates
    fallback_period_sec: 60
    # If the data feed changes by more than this percentage, the asset should be added to the batched updates
    percent_change_threshold: 1.0
    # Optional: Store historical data on-chain (default: false)
    historic: false
```

See [configs/example/pusher-asset-config.yaml](configs/example/pusher-asset-config.yaml) for an example.

### Rust

Please ensure you've run `make rust` in the root of this repo before running the pusher, as portions of the pusher rely on calls to libraries built with rust and linked to the pusher via cgo. This is not necessary if you are running via docker.

## EVM Chain Setup

### Wallet Setup

Create a `.env` file containing the private key of your wallet. This is needed to pay gas/transaction fees.

See [.env.example](.env.example) for an example. This example also includes variables used to configure the accompanying publisher agent. For more information on the publisher agent see [Publisher Agent README](../publisher_agent/README.md).

### Running the EVM Pusher

For full explanation of the flags, run:

```bash
go run main.go evm --help
```

Basic usage:

```bash
go run main.go evm \
    -c <chain-rpc-url> \
    -x <contract-address> \
    -f <asset-config-file> 
```

### Running with Docker

For local development with a full stack: Data Provider -> Publisher Agent -> First Party Pusher -> First Party Stork Contract:

```bash
# Start local stack
docker compose --profile local up
```

This is the recommended method.

For local development with Publisher Agent -> First Party Pusher:

```bash
docker compose --profile production up
```
