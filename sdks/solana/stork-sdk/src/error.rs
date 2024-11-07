use anchor_lang::error_code;

#[error_code]
#[derive(PartialEq)]
pub enum GetTemporalNumericValueError {
    #[msg("The feed id is invalid")]
    InvalidFeedId,
    #[msg("Error deserializing AccountInfo")]
    DeserializationError,
}
