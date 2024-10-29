use anchor_lang::{declare_id, prelude::Pubkey, pubkey};

pub mod error;
pub mod pda;
pub mod temporal_numeric_value;
pub mod program;

pub const PROGRAM_ID: Pubkey = pubkey!("2TSL7JwuTu9co7yUizwuh8EVdd4d96vDo9JykQCw8SHi");
declare_id!(PROGRAM_ID);
