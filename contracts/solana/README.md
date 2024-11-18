# Stork Solana Contract

This directory contains an [Anchor](https://www.anchor-lang.com/) project used to manage and deploy the Stork Solana compatible contract.

This contract is used to write the latest values from the Stork network on-chain. For reading values on chain, see the [stork-sdk](../../sdks/solana/stork-sdk).

### TemporalNumericValueFeed Accounts

On Solana, Stork price feeds exist as on-chain accounts. These accounts are instances of the [`TemporalNumericValueFeed` account](../../sdks/solana/stork-sdk/src/temporal_numeric_value.rs), and are created and owned by the Stork Oracle contract. These account have an ID which associates them with a specific asset, and a `latest_value` field which stores the latest price update. The ID of a TemporalNumericValueFeed account is determined by taking the keccak256 hash of the asset ID.

#### Writing to a TemporalNumericValueFeed Account

A new value can be written to a feed account by calling the [`update_temporal_numeric_value_evm`](./programs/stork/src/lib.rs) instruction. This instruction takes a [`TemporalNumericValueEvmInput`](./programs/stork/src/lib.rs) struct as input. 

#### Signature Validation

In order for a new value to be accepted by the contract, the signature associated with the value must be validated against Stork's public key.

This signature is derived from

1. Stork's public key
2. Encoded Asset ID (keccak256 hash of the asset's symbol)
3. Quantized value
4. Timestamp in nanoseconds
5. Merkle root of the signed publisher message hashes
6. Value computation algorithm hash

### Treasury Accounts

Treasury accounts collect small fees from users submitting new values to the stork contract. Currently, these fees are not in use, and thus the update fee is set to the minimum of 1 lamport, or 0.000000001 SOL.

### StorkConfig Account

The [`StorkConfig` account](./programs/stork/src/lib.rs) stores the Stork Solana public key, the Stork EVM public key, the update fee, and the owner of the Stork contract. The owner fields is used to validate the caller ofadmin instructions

### Program Address

The program is currently deployed on Solana Mainnet and Devnet with address:
```
stork1JUZMKYgjNagHiK2KdMmb42iTnYe9bYUCDUk8n
```

### Getting started

```
solana-keygen new --outfile ~/.config/solana/id.json
export COPYFILE_DISABLE=1 # for macos
yarn install
anchor test
```

### Local Development

#### Run local node

This deploys all contracts in the Anchor workspace to the localnet cluster

```
anchor localnet
```

#### Deploy

```
anchor deploy
```

#### Upgrade

```
anchor upgrade
```

#### Verify

```
anchor verify <program-id>
```

#### Test

```
anchor test
```

#### Generate IDL

```
anchor build
```
#### Initializing contract

```
ANCHOR_PROVIDER_URL=http://localhost:8899 ANCHOR_WALLET=~/.config/solana/id.json npx ts-node ./app/admin.ts initialize
```

#### Deploying contract on-chain

```
anchor deploy --provider.cluster mainnet
```
