use {
    crate::PROGRAM_ID, anchor_lang::prelude::*
};

pub const STORK_CONFIG_SEED: &str = "stork_config";
pub const STORK_FEED_SEED: &str = "stork_feed";
pub const STORK_TREASURY_SEED: &str = "stork_treasury";

pub fn get_config_address() -> Pubkey {
    Pubkey::find_program_address(&[STORK_CONFIG_SEED.as_ref()], &PROGRAM_ID).0
}

// There is one treasury for each u8 value to load balance the write load
pub fn get_treasury_address(treasury_id: u8) -> Pubkey {
    Pubkey::find_program_address(&[STORK_TREASURY_SEED.as_ref(), &[treasury_id]], &PROGRAM_ID).0
}

pub fn get_feed_address(feed_id: [u8; 32]) -> Pubkey {
    Pubkey::find_program_address(&[STORK_FEED_SEED.as_ref(), &feed_id], &PROGRAM_ID).0
}
