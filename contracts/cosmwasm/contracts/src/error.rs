use sylvia::cw_std::StdError;
use thiserror::Error;

#[derive(Error, Debug, PartialEq)]
pub enum StorkError {
    #[error("{0}")]
    Std(#[from] StdError),
    #[error("Invalid signature")]
    InvalidSignature,
    #[error("Insufficient funds")]
    InsufficientFunds,
    #[error("Feed not found")]
    FeedNotFound,
    #[error("Not Authorized")]
    NotAuthorized,
}
