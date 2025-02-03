# Stork CosmWasm Contract

This directory contains the Stork compatable contract in the form of a [Sylvia](https://github.com/CosmWasm/sylvia) project, as well as a CLI tool used to manage the Stork CosmWasm compatible contract.

This contract can be used as an SDK with the `library` feature, and is available on [crates.io](https://crates.io/crates/stork-cw).

### Getting started

As there is no core CosmWasm chain, but rather a multitude of chains built on top of CosmWasm, the specifics of development will vary depending on the chain.

For specific chain development, please refer to the chain's documentation.

For general purpose development and testing, we recommend using the Osmosis testnet, though the contract is compatible with any CosmWasm chain.

### Development

ensure you have the correct target installed for the chain you are developing on. This is typycially `wasm32-unknown-unknown`. 

```bash
rustup target add wasm32-unknown-unknown
```

#### Build

```bash
cargo wasm 
```

#### Test

```bash
cargo test
```

#### Optimized Build

The contract can be built with optimizations using the CosmWasm optimizer. This is recommended for production builds. The latest version of the optimizer can be found [here](https://github.com/CosmWasm/optimizer).

*The following command may not reflect the latest version of the optimizer.*

```bash
docker run --rm -v "$(pwd)":/code \
  --mount type=volume,source="$(basename "$(pwd)")_cache",target=/target \
  --mount type=volume,source=registry_cache,target=/usr/local/cargo/registry \
  cosmwasm/optimizer:0.16.0
```

#### Deploy

This will vary chain to chain, but typically looks something like this Osmosis Testnet example:

```bash
osmosisd tx wasm store artifacts/stork.wasm --from wallet --chain-id=osmo-test-5 --gas-prices=0.1uosmo --gas=auto --gas-adjustment 1.3 -y --output json -b sync 
```
#### Generate JSON Schema

```bash
cargo run schema
```

#### Generate Typescript Types

This step is only necessary if you update or add entrypoints in the contract and need to update the CLI tool.

```bash
npm install @cosmwasm/ts-codegen
npx @cosmwasm/ts-codegen generate --plugin client --schema ./schema --out ../cli/client/ --name Stork --no-bundle
```

### Note

Though this contract is built with Sylvia, it is compatible with any CosmWasm contract.
