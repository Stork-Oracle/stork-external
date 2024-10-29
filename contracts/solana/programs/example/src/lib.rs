use anchor_lang::prelude::*;
use stork_sdk::{
    pda::STORK_FEED_SEED,
    temporal_numeric_value::TemporalNumericValueFeed,
};

declare_id!("FGpoDwQC8gYadJAsB9vrsgPN38qkqDgSk3qQcaRiyPra");

#[program]
pub mod example {
    use super::*;

    // This instruction reads the latest price from a Stork feed
    pub fn read_price(ctx: Context<ReadPrice>, feed_id: [u8; 32]) -> Result<()> {
        let feed = &ctx.accounts.feed;
        
        // Get the latest timestamp and value from the feed
        let timestamp = feed.latest_value.timestamp_ns;
        let value = feed.latest_value.quantized_value;
        
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
        seeds::program = stork_sdk::ID
    )]
    pub feed: Account<'info, TemporalNumericValueFeed>,
}
