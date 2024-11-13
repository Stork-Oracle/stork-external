//! Provides the Stork program struct for implementing the anchor ID trait.
use anchor_lang::prelude::*;

/// The Stork program struct.
pub struct Stork;

/// Implements the anchor ID trait for the Stork program.
impl Id for Stork {
    fn id() -> Pubkey {
        crate::ID
    }
}
