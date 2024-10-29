use anchor_lang::error_code;

#[error_code]
#[derive(PartialEq)]
pub enum GetTemporalNumericValueError {
    #[msg("The feed id is invalid")]
    InvalidFeedId,
}

#[macro_export]
macro_rules! check {
    ($cond:expr, $err:expr) => {
        if !$cond {
            return Err($err);
        }
    };
}
