# Stork Cosmwasm Contract

This is the Stork CosmWasm compatible contract. This crate is maintained by [Stork Labs](https://stork.network).

It is available on [crates.io](https://crates.io/crates/stork-cw).

This crate can be used as an SDK to build contracts that interact with the Stork Contract by including it as a depency and enabling the `library` feature.

## Pull Model

The Stork Cosmwasm Contract allows users to consume Stork price updates on a pull basis. This puts the responsibility of submitting the price updates on-chain to the user whenever they want to interact with an app that consumes Stork price feeds. Stork Labs maintains a [Chain Pusher](https://github.com/stork-oracle/stork-external/apps/docs/chain_pusher) in order to do this.

## Stork Feeds

On CosmWasm, Stork feeds exist inside a table stored on-chain. This table associates a given encoded asset id (keccak256 of the plaintext asset id) with a [`TemporalNumericValue`](./src/temporal_numeric_value.rs) instance.

## Sylvia

This contract is built using the [Sylvia Framework](https://github.com/CosmWasm/sylvia). This heavily reduces the amount of boilerplate needed to create a CosmWasm contract while remaining fully compatible with the CosmWasm SDK. This generates the `sv` and `entry_points` modules, in the `contract` module.

Examples of using this crate as an SDK in both Sylvia and Non-Sylvia contracts to consume Stork data can be found in the [stork-external github repo](https://github.com/stork-oracle/stork-external/tree/main/examples/cosmwasm).
