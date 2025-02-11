//! Errors for the Stork CosmWasm Contract.
use sylvia::cw_std::StdError;
use thiserror::Error;

#[derive(Error, Debug, PartialEq)]
pub enum StorkError {
    #[error("{0}")]
    Std(#[from] StdError),
    #[error("Invalid signature: {0}")]
    InvalidSignature(String),
    #[error("Insufficient funds")]
    InsufficientFunds,
    #[error("Feed not found")]
    FeedNotFound,
    #[error("Not Authorized")]
    NotAuthorized,
}
