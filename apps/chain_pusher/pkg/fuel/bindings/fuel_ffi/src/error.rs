// error.rs
use crate::fuel_client::StorkContractError;
use fuels::core::codec::LogDecoder;
use fuels::types::errors::transaction::Reason;
use fuels::types::errors::Error as FuelError;
use thiserror::Error;

#[derive(Error, Debug)]
pub enum FuelClientError {
    #[error("Invalid configuration: {0}")]
    InvalidConfig(String),
    #[error("Contract call failed: {0}")]
    ContractCallFailed(String),
    #[error("Network error: {0}")]
    NetworkError(String),
    #[error("Invalid transaction parameters: {0}")]
    InvalidTransactionParameters(String),
    #[error("JSON serialization error: {0}")]
    JsonError(#[from] serde_json::Error),
    #[error("System/runtime error: {0}")]
    SystemError(String),
    #[error("Error getting wallet balance: {0}")]
    WalletBalanceError(String),
    #[error("Null pointer passed: {0}")]
    NullPointer(String),
    // Contract errors, these correspond with custom contract errors defined in the stork-fuel-sdk crate
    #[error("Incorrect fee asset")]
    IncorrectFeeAsset,
    #[error("Insufficient fee")]
    InsufficientFee,
    #[error("No fresh update")]
    NoFreshUpdate,
    #[error("Feed not found")]
    FeedNotFound,
    #[error("Invalid signature")]
    InvalidSignature,
}

// type safe C error codes
#[repr(C)]
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum FuelClientStatus {
    Success = 0,
    InvalidConfig = 1,
    ContractCallFailed = 2,
    NetworkError = 3,
    InvalidTransactionParameters = 4,
    JsonError = 5,
    SystemError = 6,
    WalletBalanceError = 7,
    NullPointer = 8,
    IncorrectFeeAsset = 9,
    InsufficientFee = 10,
    NoFreshUpdate = 11,
    FeedNotFound = 12,
    InvalidSignature = 13,
}

impl From<FuelClientError> for FuelClientStatus {
    fn from(error: FuelClientError) -> Self {
        match error {
            FuelClientError::InvalidConfig(_) => Self::InvalidConfig,
            FuelClientError::ContractCallFailed(_) => Self::ContractCallFailed,
            FuelClientError::NetworkError(_) => Self::NetworkError,
            FuelClientError::InvalidTransactionParameters(_) => Self::InvalidTransactionParameters,
            FuelClientError::JsonError(_) => Self::JsonError,
            FuelClientError::SystemError(_) => Self::SystemError,
            FuelClientError::WalletBalanceError(_) => Self::WalletBalanceError,
            FuelClientError::NullPointer(_) => Self::NullPointer,
            FuelClientError::IncorrectFeeAsset => Self::IncorrectFeeAsset,
            FuelClientError::InsufficientFee => Self::InsufficientFee,
            FuelClientError::NoFreshUpdate => Self::NoFreshUpdate,
            FuelClientError::FeedNotFound => Self::FeedNotFound,
            FuelClientError::InvalidSignature => Self::InvalidSignature,
        }
    }
}

impl From<StorkContractError> for FuelClientError {
    fn from(error: StorkContractError) -> Self {
        match error {
            StorkContractError::IncorrectFeeAsset(_) => Self::IncorrectFeeAsset,
            StorkContractError::InsufficientFee(_) => Self::InsufficientFee,
            StorkContractError::NoFreshUpdate => Self::NoFreshUpdate,
            StorkContractError::FeedNotFound(_) => Self::FeedNotFound,
            StorkContractError::InvalidSignature(_) => Self::InvalidSignature,
        }
    }
}

// helper function to process errors coming from .call or .simulate, determining if the error is a StorkContractError
pub fn process_contract_error(e: FuelError, log_decoder: &LogDecoder) -> FuelClientError {
    match e {
        FuelError::Transaction(reason) => {
            match reason {
                Reason::Failure {
                    reason,
                    revert_id: _,
                    receipts,
                } => {
                    let decoded_logs =
                        log_decoder.decode_logs_with_type::<StorkContractError>(&receipts);
                    let errors = match decoded_logs {
                        Ok(val) => val,
                        Err(_) => return FuelClientError::ContractCallFailed(reason),
                    };
                    if errors.is_empty() {
                        return FuelClientError::ContractCallFailed(reason);
                    }
                    // return the first error found in the case of multiple
                    errors[0].clone().into()
                }
                _ => FuelClientError::ContractCallFailed(reason.to_string()),
            }
        }
        _ => FuelClientError::ContractCallFailed(e.to_string()),
    }
}
