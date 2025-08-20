# Stork Sway SDK

This is a sway sdk to build Fuel contracts that consume Stork price feeds. This package is maintained by [Stork Labs](https://stork.network).

It is available on [forc.pub](https://forc.pub/package/stork_sway_sdk).

## Pull Model 

The Stork Sway SDK allows users to consume Stork price updates on a pull basis. This puts the responsibility of submitting the price updates on-chain to the user whenever they want to interact with an app that consumes Stork price feeds. Stork Labs maintains a [Chain Pusher](https://github.com/stork-oracle/stork-external/tree/main/apps/docs/chain_pusher.md) in order to do this.

## Details

The Stork Sway SDK provides a set of useful features for building Fuel contracts that consume Stork price feeds. Primarily, a consuming contract will be using:

- The `Stork` abi interface, in `stork_sway_sdk::interface`
- The `get_temporal_numeric_value_unchecked_v1` function on the `Stork` interface to get the latest price update
- The `TemporalNumericValue` struct, in `stork_sway_sdk::temporal_numeric_value`, which is the struct that represents a price update and is returned by the `get_temporal_numeric_value_unchecked_v1` function

## Example 

The following snippet is an example of how to use this sdk to consume Stork price feeds on chain. A full example is available [here](https://github.com/Stork-Oracle/stork-external/tree/main/examples/fuel).

```rust
    // This function reads the latest price from a Stork feed
    fn read_price(feed_id: b256, stork_contract_address: b256) {
        // Get the stork contract
        let stork_contract = abi(Stork, stork_contract_address);

        // Get the price
        let price = stork_contract.get_temporal_numeric_value_unchecked_v1(feed_id);

        // You can do something with the price here...
    }
```
