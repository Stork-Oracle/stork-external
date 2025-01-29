use anchor_lang::prelude::*;
use stork_solana_sdk::{
    pda::STORK_FEED_SEED,
    temporal_numeric_value::TemporalNumericValueFeed,
};

// Change this to the program ID you're using
declare_id!("GzkgPe7VSGeqC6QsUMJjL9FWSqfyqbqcJxe5FW2Xjm61");

#[program]
pub mod example {
    use super::*;

    // This instruction reads the latest price from a Stork feed
    pub fn read_price(ctx: Context<ReadPrice>, feed_id: [u8; 32]) -> Result<()> {
        let feed = &ctx.accounts.feed;

        let latest_value = feed.get_latest_canonical_temporal_numeric_value_unchecked(&feed_id)?;
        
        // Get the latest timestamp and value from the feed
        let timestamp = latest_value.timestamp_ns;
        let value = latest_value.quantized_value;
        
        msg!(
            "Feed {} - Latest value: {}, Timestamp: {}", 
            hex::encode(feed_id),
            value,
            timestamp
        );

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
