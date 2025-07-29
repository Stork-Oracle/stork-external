// error.rs
use thiserror::Error;

#[derive(Error, Debug)]
pub enum FuelClientError {
    #[error("Invalid configuration: {0}")]
    InvalidConfig(String),
    #[error("Contract call failed: {0}")]
    ContractCallFailed(String),
    #[error("Insufficient balance: {balance}, required: {required}")]
    InsufficientBalance { balance: u64, required: u64 },
    #[error("Network error: {0}")]
    NetworkError(String),
    #[error("UTXO already spent (concurrent transaction)")]
    UtxoSpent,
    #[error("Invalid transaction parameters: {0}")]
    InvalidTransaction(String),
    #[error("JSON serialization error: {0}")]
    JsonError(#[from] serde_json::Error),
    #[error("System/runtime error: {0}")]
    SystemError(String), // For tokio runtime, threading, OS issues
    #[error("Fuel SDK error: {0}")]
    FuelSdkError(String), // For SDK-specific errors that don't fit elsewhere
    #[error("Client is null")]
    NullClient,
    #[error("Invalid Stork error code: {0}")]
    UnknownContractError(String),
}

// type safe C error codes
#[repr(C)]
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum FuelClientErrorCode {
    Success = 0,
    InvalidConfig = 1,
    ContractCallFailed = 2,
    InsufficientBalance = 3,
    NetworkError = 4,
    UtxoSpent = 5,
    InvalidTransaction = 6,
    JsonError = 7,
    SystemError = 8,
    FuelSdkError = 9,
    NullClient = 10,
    UnknownContractError = 11,
}

impl From<FuelClientError> for FuelClientErrorCode {
    fn from(error: FuelClientError) -> Self {
        match error {
            FuelClientError::InvalidConfig(_) => Self::InvalidConfig,
            FuelClientError::ContractCallFailed(_) => Self::ContractCallFailed,
            FuelClientError::InsufficientBalance { .. } => Self::InsufficientBalance,
            FuelClientError::NetworkError(_) => Self::NetworkError,
            FuelClientError::UtxoSpent => Self::UtxoSpent,
            FuelClientError::InvalidTransaction(_) => Self::InvalidTransaction,
            FuelClientError::JsonError(_) => Self::JsonError,
            FuelClientError::SystemError(_) => Self::SystemError,
            FuelClientError::FuelSdkError(_) => Self::FuelSdkError,
            FuelClientError::NullClient => Self::NullClient,
            FuelClientError::UnknownContractError(_) => Self::UnknownContractError,
        }
    }
}


// StorkError enum matching the StorkError enum in the Stork Fuel SDK
#[repr(C)]
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum StorkError {
    InsufficientFee = 0,
    NoFreshUpdate = 1,
    FeedNotFound = 2,
    StaleValue = 3,
    InvalidSignature = 4,
}

impl TryFrom<u64> for StorkError {
    type Error = FuelClientError;
    fn try_from(value: u64) -> Result<Self, Self::Error> {
        match value {
            0 => Ok(StorkError::InsufficientFee),
            1 => Ok(StorkError::NoFreshUpdate),
            2 => Ok(StorkError::FeedNotFound),
            3 => Ok(StorkError::StaleValue),
            4 => Ok(StorkError::InvalidSignature),
            _ => Err(FuelClientError::UnknownContractError(value.to_string())),
        }
    }
}
