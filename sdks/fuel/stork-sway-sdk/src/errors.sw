library;

pub enum StorkError {
    // Insufficient fee is paid to the method.
    InsufficientFee: (),
    // There is no fresh update, whereas expected fresh updates.
    NoFreshUpdate: (),
    // Not found.
    FeedNotFound: (),
    // Requested value is stale.
    StaleValue: (),
    // Signature is invalid.
    InvalidSignature: (),
    // Invalid stork public key.
    InvalidStorkPublicKey: (),
}
