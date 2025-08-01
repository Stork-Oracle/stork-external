library;

use ::interface::TemporalNumericValueInput;

#[error_type]
pub enum StorkError {
    // Incorrect fee asset. Includes the provided asset id.
    #[error(m = "Incorrect fee asset.")]
    IncorrectFeeAsset: AssetId,
    // Insufficient fee for number of updates. Includes the provided fee.
    #[error(m = "Insufficient fee for updates.")]
    InsufficientFee: u64,
    // There is no fresh update.
    #[error(m = "No fresh update.")]
    NoFreshUpdate: (),
    // Feed not found. Includes the id of the feed that was not found.
    #[error(m = "Feed not found.")]
    FeedNotFound: b256,
    // Signature is invalid. Includes the id of the feed that was invalid.
    #[error(m = "Invalid signature.")]
    InvalidSignature: TemporalNumericValueInput,
}
