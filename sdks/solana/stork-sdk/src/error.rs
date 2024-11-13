//! Error codes for the Stork Solana SDK.
use anchor_lang::error_code;

#[error_code]
#[derive(PartialEq)]
pub enum GetTemporalNumericValueError {
    /// The feed id is invalid.
    #[msg("The feed id is invalid")]
    InvalidFeedId,
    /// Error deserializing AccountInfo.
    #[msg("Error deserializing AccountInfo")]
    DeserializationError,
}
