/*! [![stork logo](https://raw.githubusercontent.com/Stork-Oracle/stork-external/refs/heads/main/public/stork-logo-slim.png)](https://stork.network)

This is a Rust SDK to build Solana programs that consume Stork price feeds. This crate is maintained by [Stork Labs](https://stork.network).

A top level overview, as well as a small code sample, can be found in [the libraries entry in crates.io](https://crates.io/crates/stork-solana-sdk).

# Pull Model

The Stork Solana SDK allows users to consume Stork price updates on a pull basis. This puts the responsibility of submitting the price updates on-chain to the user whenever they want to interact with an app that consumes Stork price feeds. Stork Labs maintains a [Chain Pusher](https://github.com/stork-oracle/stork-external/tree/main/apps/chain_pusher/README.md) in order to do this.

# Stork Price Feed Accounts

On Solana, Stork price feeds exist as on-chain accounts. These accounts are instances of the [TemporalNumericValueFeed](temporal_numeric_value::TemporalNumericValueFeed) account struct, and are created and owned by the Stork Oracle contract. These account have an ID which associates them with a specific asset, and a `latest_value` field which stores the latest price update. The ID of a TemporalNumericValueFeed account is determined by taking the keccak256 hash of the asset ID. The `latest_value` field is an instance of the [TemporalNumericValue](temporal_numeric_value::TemporalNumericValue) struct, which holds the quantized price, and a unix timestamp indicating when the price was created.

# Anchor

The Stork Solana SDK is built on top of [Anchor](https://github.com/coral-xyz/anchor), a framework for building Solana programs.

*/
use anchor_lang::{declare_id, prelude::Pubkey, pubkey};

pub mod error;
pub mod pda;
pub mod program;
pub mod temporal_numeric_value;

/// The ID of the Stork Oracle program.
pub const PROGRAM_ID: Pubkey = pubkey!("stork1JUZMKYgjNagHiK2KdMmb42iTnYe9bYUCDUk8n");
declare_id!(PROGRAM_ID);
