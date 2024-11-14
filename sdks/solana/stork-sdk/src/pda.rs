//! Provides functions to get Program Derived Addresses (PDAs) for the Stork Solana SDK.
use {
    crate::PROGRAM_ID, anchor_lang::prelude::*
};

/// The seed for the Stork config PDA.
pub const STORK_CONFIG_SEED: &[u8] = b"stork_config";
/// The seed for the Stork feed PDA.
pub const STORK_FEED_SEED: &[u8] = b"stork_feed";
/// The seed for the Stork treasury PDA.
pub const STORK_TREASURY_SEED: &[u8] = b"stork_treasury";

/// Gets the address of the Stork config PDA.
/// There is only one Stork config account, and only the owner of the Stork Oracle program can interact with it.
pub fn get_config_address() -> Pubkey {
    Pubkey::find_program_address(&[STORK_CONFIG_SEED.as_ref()], &PROGRAM_ID).0
}

/// Gets the address of the Stork treasury PDA.
/// There is one treasury for each u8 value to load balance the write load.
pub fn get_treasury_address(treasury_id: u8) -> Pubkey {
    Pubkey::find_program_address(&[STORK_TREASURY_SEED.as_ref(), &[treasury_id]], &PROGRAM_ID).0
}

/// Gets the address of the Stork feed PDA.
/// The ID of a Stork feed PDA is determinined by taking the keccak256 hash of the asset ID.
pub fn get_feed_address(feed_id: [u8; 32]) -> Pubkey {
    Pubkey::find_program_address(&[STORK_FEED_SEED.as_ref(), &feed_id], &PROGRAM_ID).0
}
