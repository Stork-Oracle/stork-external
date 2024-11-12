use anchor_lang::{declare_id, prelude::Pubkey, pubkey};

pub mod error;
pub mod pda;
pub mod temporal_numeric_value;
pub mod program;

pub const PROGRAM_ID: Pubkey = pubkey!("stork1JUZMKYgjNagHiK2KdMmb42iTnYe9bYUCDUk8n");
declare_id!(PROGRAM_ID);
