# Stork Cosmwasm Contract

This directory contains the Stork compatable contract in the form of a [Sylvia](https://github.com/CosmWasm/sylvia) project, as well as a CLI tool used to manage the Stork Cosmwasm compatible contract.

### Getting started

As there is no core cosmwasm chain, but rather a multitude of chains built on top of cosmwasm, the specifics of development will vary depending on the chain.

For specific chain development, please refer to the chain's documentation.

For general purpose development and testing, we recommend using the Osmosis testnet, though the contract is compatible with any cosmwasm chain.

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

The contract can be built with optimizations using the cosmwasm optimizer. This is recommended for production builds. The latest version of the optimizer can be found [here](https://github.com/CosmWasm/optimizer).

```bash
docker run --rm -v "$(pwd)":/code \
  --mount type=volume,source="$(basename "$(pwd)")_cache",target=/target \
  --mount type=volume,source=registry_cache,target=/usr/local/cargo/registry \
  cosmwasm/optimizer:0.16.0
```

### Note

Though this contract is built with Sylvia, it is compatible with any cosmwasm contract.