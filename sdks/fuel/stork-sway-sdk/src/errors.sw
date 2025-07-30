library;

#[error_type]
pub enum StorkError {
    // Insufficient fee is paid to the method.
    #[error(m = "Insufficient fee")]
    InsufficientFee: (),
    // There is no fresh update, whereas expected fresh updates.
    #[error(m = "No fresh update")]
    NoFreshUpdate: (),
    // Not found.
    #[error(m = "Feed not found")]
    FeedNotFound: (),
    // Requested value is stale.
    #[error(m = "Stale value")]
    StaleValue: (),
    // Signature is invalid.
    #[error(m = "Invalid signature")]
    InvalidSignature: (),
}
