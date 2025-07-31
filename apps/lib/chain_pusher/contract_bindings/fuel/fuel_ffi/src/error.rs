// error.rs
use thiserror::Error;
use fuels::types::errors::Error as FuelError;
use fuels::types::errors::transaction::Reason;
use fuels::core::codec::LogDecoder;
use crate::fuel_client::StorkContractError;

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

// helper function to process errors coming from .call or .simulate, determining if the error is a Str
pub fn process_contract_error(e: FuelError, log_decoder: &LogDecoder) -> FuelClientError {
    match e {
        FuelError::Transaction(reason) => {
            match reason {
                Reason::Failure {
                    reason: reason_str,
                    revert_id,
                    receipts,
                } => {
                    println!("reason: {:?}", reason_str);
                    println!("receipts: {:?}", receipts);
                    let logs = log_decoder.decode_logs(&receipts);
                    println!("logs: {:?}", logs);
                    FuelClientError::ContractCallFailed(reason_str)
                }
                _ => FuelClientError::ContractCallFailed(reason.to_string()),
            }

        }
        _ => FuelClientError::ContractCallFailed(e.to_string()),
    }
}

// impl TryFrom<FuelError> for StorkContractError {
//     type Error = FuelClientError;
//     fn try_from(value: FuelError) -> Result<Self, Self::Error> {
//         match value {
//             FuelError::Transaction(reason) => {
//                 match reason {
//                     Reason::Failure {
//                         reason: reason_str,
//                         revert_id,
//                         receipts,
//                     } => {
//                         // hardcoded from abi, hopefully there's a cleaner way to do this in the future (github issue: https://github.com/FuelLabs/fuels-rs/issues/1680)
//                         match revert_id {
//                             Some()

//                         }
//                     }
//                 }
//             }
//         }
//     }
// }
