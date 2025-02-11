# Stork Solana SDK

This is a Rust SDK to build Solana programs that consume Stork price feeds. This crate is maintained by [Stork Labs](https://stork.network).

It is available on [crates.io](https://crates.io/crates/stork-solana-sdk).

## Pull Model

The Stork Solana SDK allows users to consume Stork price updates on a pull basis. This puts the responsibility of submitting the price updates on-chain to the user whenever they want to interact with an app that consumes Stork price feeds. Stork Labs maintains a [Chain Pusher](https://github.com/stork-oracle/stork-external/apps/docs/chain_pusher) in order to do this.

## Stork Price Feed Accounts

On Solana, Stork price feeds exist as on-chain accounts. These accounts are instances of the [`TemporalNumericValueFeed` account](./src/temporal_numeric_value.rs), and are created and owned by the Stork Oracle contract. These account have an ID which associates them with a specific asset, and a `latest_value` field which stores the latest price update.

## Example

The following snippet is an example of how to use this sdk to consume Stork price feed on chain. A full example is available [here](https://github.com/Stork-Oracle/stork-external/tree/main/examples/solana).

```rust 
    // This instruction reads the latest price from a Stork feed
    pub fn read_price(ctx: Context<ReadPrice>, feed_id: [u8; 32]) -> Result<()> {
        let feed = &ctx.accounts.feed;

        let latest_value = feed.get_latest_canonical_temporal_numeric_value_unchecked(&feed_id)?;
        
        // Get the latest timestamp and value from the feed
        let timestamp = latest_value.timestamp_ns;
        let value = latest_value.quantized_value;

        // You can do something with the price here...
        
        Ok(())
    }
}

#[derive(Accounts)]
#[instruction(feed_id: [u8; 32])]
pub struct ReadPrice<'info> {
    // This account holds the price feed data
    #[account(
        seeds = [STORK_FEED_SEED.as_ref(), feed_id.as_ref()],
        bump,
        seeds::program = stork_solana_sdk::ID
    )]
    pub feed: Account<'info, TemporalNumericValueFeed>,
}
```
