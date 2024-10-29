use anchor_lang::prelude::*;

pub struct Stork;

impl Id for Stork {
    fn id() -> Pubkey {
        crate::ID
    }
}
