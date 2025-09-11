use anyhow::Result;
use std::str::FromStr;
use std::sync::Arc;
use tokio::runtime::Runtime;

use crate::error::{process_contract_error, FuelClientError};
use fuels::{
    accounts::signers::private_key::PrivateKeySigner,
    crypto::SecretKey,
    prelude::*,
    programs::calls::Execution,
    types::{AssetId, Bits256, ContractId},
};
use serde::{Deserialize, Serialize};

// Generate the contract bindings from ABI (path is relative to the workspace root Cargo.toml)
abigen!(Contract(
    name = "StorkContract",
    abi = "./apps/chain_pusher/pkg/fuel/bindings/fuel_ffi/stork-abi.json"
),);

// re-export generated StorkError
pub use StorkError as StorkContractError;
// FFI-compatible structures for JSON serialization
#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct FuelTemporalNumericValue {
    pub timestamp_ns: u64,
    pub quantized_value: i128,
}

#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct FuelTemporalNumericValueInput {
    pub temporal_numeric_value: FuelTemporalNumericValue,
    pub id: String,                     // b256 as hex string
    pub publisher_merkle_root: String,  // b256 as hex string
    pub value_compute_alg_hash: String, // b256 as hex string
    pub r: String,                      // b256 as hex string
    pub s: String,                      // b256 as hex string
    pub v: u8,
}

#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct FuelConfig {
    pub rpc_url: String,
    pub contract_address: String,
    pub private_key: String,
}

pub struct FuelClient {
    wallet: Wallet,
    proxy_contract: StorkContract<Wallet>,
    pub rt: Arc<Runtime>,
    gas_asset_id: AssetId,
}

impl FuelClient {
    pub async fn new(config: FuelConfig) -> Result<Self, FuelClientError> {
        let provider = Provider::connect(&config.rpc_url)
            .await
            .map_err(|e| FuelClientError::NetworkError(format!("Failed to connect to RPC: {e}")))?;

        let secret_key = SecretKey::from_str(&config.private_key)
            .map_err(|e| FuelClientError::InvalidConfig(format!("Invalid private key: {e}")))?;

        let proxy_contract_id = ContractId::from_str(&config.contract_address).map_err(|e| {
            FuelClientError::InvalidConfig(format!("Invalid contract address: {e}"))
        })?;

        let rt = Arc::new(
            tokio::runtime::Builder::new_multi_thread()
                .enable_all()
                .build()
                .map_err(|e| {
                    FuelClientError::SystemError(format!("Failed to create async runtime: {e}"))
                })?,
        );

        let base_asset_id = provider
            .consensus_parameters()
            .await
            .map_err(|e| {
                FuelClientError::NetworkError(format!("Failed to get consensus parameters: {e}"))
            })?
            .base_asset_id()
            .to_owned();

        let signer = PrivateKeySigner::new(secret_key);

        let wallet = Wallet::new(signer, provider.clone());

        let proxy_contract = StorkContract::new(proxy_contract_id, wallet.clone());

        Ok(FuelClient {
            wallet,
            proxy_contract,
            rt,
            gas_asset_id: base_asset_id,
        })
    }

    pub async fn get_latest_temporal_numeric_value(
        &self,
        id: [u8; 32],
    ) -> Result<Option<FuelTemporalNumericValue>, FuelClientError> {
        let id_bits256 = Bits256(id);

        let response = self
            .proxy_contract
            .methods()
            .get_temporal_numeric_value_unchecked_v1(id_bits256)
            .determine_missing_contracts()
            .await
            .map_err(|e| {
                FuelClientError::ContractCallFailed(format!(
                    "Failed to determine missing contracts: {e}"
                ))
            })?
            .simulate(Execution::state_read_only())
            .await
            .map_err(|e| process_contract_error(e, &self.proxy_contract.log_decoder()))?;

        let contract_tnv = response.value;
        let tnv = FuelTemporalNumericValue {
            timestamp_ns: contract_tnv.timestamp_ns,
            quantized_value: contract_tnv.quantized_value.underlying as i128,
        };
        Ok(Some(tnv))
    }

    pub async fn update_temporal_numeric_values(
        &self,
        inputs: Vec<FuelTemporalNumericValueInput>,
    ) -> Result<String, FuelClientError> {
        // Convert inputs to the generated contract types
        let mut contract_inputs = Vec::new();

        for input in inputs {
            // Parse hex strings to Bits256
            let id = Bits256::from_hex_str(&input.id).map_err(|e| {
                FuelClientError::InvalidTransactionParameters(format!("Invalid id: {e}"))
            })?;
            let publisher_merkle_root = Bits256::from_hex_str(&input.publisher_merkle_root)
                .map_err(|e| {
                    FuelClientError::InvalidTransactionParameters(format!(
                        "Invalid publisher merkle root: {e}"
                    ))
                })?;
            let value_compute_alg_hash = Bits256::from_hex_str(&input.value_compute_alg_hash)
                .map_err(|e| {
                    FuelClientError::InvalidTransactionParameters(format!(
                        "Invalid value compute alg hash: {e}"
                    ))
                })?;
            let r = Bits256::from_hex_str(&input.r).map_err(|e| {
                FuelClientError::InvalidTransactionParameters(format!("Invalid r: {e}"))
            })?;
            let s = Bits256::from_hex_str(&input.s).map_err(|e| {
                FuelClientError::InvalidTransactionParameters(format!("Invalid s: {e}"))
            })?;

            // Create the contract input using generated types
            let contract_input = TemporalNumericValueInput {
                temporal_numeric_value: TemporalNumericValue {
                    timestamp_ns: input.temporal_numeric_value.timestamp_ns,
                    quantized_value: input.temporal_numeric_value.quantized_value.into(),
                },
                id,
                publisher_merkle_root,
                value_compute_alg_hash,
                r,
                s,
                v: input.v,
            };

            contract_inputs.push(contract_input);
        }

        // Get the update fee from implementation contract
        let fee_response = self
            .proxy_contract
            .methods()
            .get_update_fee_v1(contract_inputs.clone())
            .determine_missing_contracts()
            .await
            .map_err(|e| {
                FuelClientError::ContractCallFailed(format!(
                    "Failed to determine missing contracts: {e}"
                ))
            })?
            .simulate(Execution::state_read_only())
            .await
            .map_err(|e| process_contract_error(e, &self.proxy_contract.log_decoder()))?;

        let fee = fee_response.value;

        let call_params = CallParameters::default()
            .with_amount(fee)
            .with_asset_id(self.gas_asset_id);

        let tx_response = self
            .proxy_contract
            .methods()
            .update_temporal_numeric_values_v1(contract_inputs.clone())
            .determine_missing_contracts()
            .await
            .map_err(|e| {
                FuelClientError::ContractCallFailed(format!(
                    "Failed to determine missing contracts: {e}"
                ))
            })?
            .call_params(call_params)
            .map_err(|e| {
                FuelClientError::ContractCallFailed(format!("Failed to set call parameters: {e}"))
            })?
            .call()
            .await
            .map_err(|e| process_contract_error(e, &self.proxy_contract.log_decoder()))?;

        match tx_response.tx_id {
            Some(tx_id) => Ok(format!("0x{tx_id}")),
            None => Err(FuelClientError::ContractCallFailed(
                "No transaction ID".to_string(),
            )),
        }
    }

    pub async fn get_wallet_balance(&self) -> Result<u64, FuelClientError> {
        let balance = self
            .wallet
            .get_asset_balance(&self.gas_asset_id)
            .await
            .map_err(|e| {
                FuelClientError::WalletBalanceError(format!("Failed to get wallet balance: {e}"))
            })?;
        Ok(balance as u64)
    }
}

// utility functions

// converts i128 to I128 using 2^127 offset
impl From<i128> for I128 {
    fn from(value: i128) -> Self {
        let offset = 1u128 << 127; // 2^127
        let quantized_with_offset = (value as u128).wrapping_add(offset);
        I128 {
            underlying: quantized_with_offset,
        }
    }
}

// undoes the 2^127 offset to get the original i128 value
impl From<I128> for i128 {
    fn from(value: I128) -> Self {
        let offset = 1u128 << 127; // 2^127
        let quantized_with_offset = value.underlying.wrapping_sub(offset);
        quantized_with_offset as i128
    }
}
