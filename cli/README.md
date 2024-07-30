# Stork CLI

A suite of tools to interact with Stork's services and on-chain contracts.

## evm-push

### Background

Stork signed data feeds are delivered off-chain from publishers to subscribers via Stork's aggregation network. In order for this data to be usable on-chain, it must be written to the Stork contract on any EVM compatible network. This tool is used to push signed data feeds to the Stork contract.

Because Stork does not write this data to the chain directly by default, any subscriber can choose to write the data to the chain if they so choose. This tool can be used to facilitate that process.

### Usage

1. Create an `asset-config.yaml` file. This file should be structured as follows:

```yaml
assets:
    BTCUSD:
        # The asset's symbol, used to subscribe to the asset on the Stork network
        asset_id: BTCUSD
        # The asset's encoded ID, used to write the asset's data to the Stork contract
        encoded_asset_id: 0x7404e3d104ea7841c3d9e6fd20adfe99b4ad586bc08d8f3bd3afef894cf184de
        # If the data feed is not updated within this period, the asset should be added to the batched updates
        fallback_period_sec: 60
        # If the data feed changes by more than this percentage, the asset should be added to the batched updates
        percent_change_threshold: 1
```

See `sample.asset-config.yaml` for an example.

2. Create a `private-key.secret` file. This file should contain the private key of the user's wallet. This is needed to pay gas/transaction fees.

3. Run the command:

For full explanation of the flags, run:
```
go run . evm-push --help
```

```
go run . evm-push \
    -w wss://api.jp.stork-oracle.network \
    -a <stork-api-key> \
    -c <chain-rpc-url> \
    -x <contract-id> \
    -f <asset-config-file> \
    -m <private-key-file>
```

## Development

Generate the contract bindings
```
abigen --abi ../contracts/evm/stork.abi --pkg main --type StorkContract --out stork_contract.go
```

