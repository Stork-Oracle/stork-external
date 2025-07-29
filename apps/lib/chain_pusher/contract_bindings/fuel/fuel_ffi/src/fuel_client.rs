use anyhow::Result;
use std::str::FromStr;
use std::sync::Arc;
use tokio::runtime::Runtime;

use crate::error::FuelClientError;
use fuels::{
    accounts::signers::private_key::PrivateKeySigner,
    crypto::SecretKey,
    prelude::*,
    programs::calls::Execution,
    types::{AssetId, Bits256, ContractId},
};
use serde::{Deserialize, Serialize};

// Generate the contract bindings from ABI
abigen!(Contract(name = "StorkContract", abi = "stork_abi.json"),);

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
        let provider = Provider::connect(&config.rpc_url).await.map_err(|e| {
            FuelClientError::NetworkError(format!("Failed to connect to RPC: {}", e))
        })?;

        let secret_key = SecretKey::from_str(&config.private_key)
            .map_err(|e| FuelClientError::InvalidConfig(format!("Invalid private key: {}", e)))?;

        let proxy_contract_id = ContractId::from_str(&config.contract_address).map_err(|e| {
            FuelClientError::InvalidConfig(format!("Invalid contract address: {}", e))
        })?;

        let rt = Arc::new(
            tokio::runtime::Builder::new_multi_thread()
                .enable_all()
                .build()
                .map_err(|e| {
                    FuelClientError::SystemError(format!("Failed to create async runtime: {}", e))
                })?,
        );

        let base_asset_id = provider
            .consensus_parameters()
            .await
            .map_err(|e| {
                FuelClientError::NetworkError(format!("Failed to get consensus parameters: {}", e))
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
                    "Failed to determine missing contracts: {}",
                    e
                ))
            })?
            .simulate(Execution::state_read_only())
            .await
            .map_err(|e| FuelClientError::ContractCallFailed(e.to_string()))?;

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
            let id = Bits256::from_hex_str(&input.id)
                .map_err(|e| FuelClientError::InvalidTransaction(format!("Invalid id: {}", e)))?;
            let publisher_merkle_root = Bits256::from_hex_str(&input.publisher_merkle_root)
                .map_err(|e| {
                    FuelClientError::InvalidTransaction(format!(
                        "Invalid publisher merkle root: {}",
                        e
                    ))
                })?;
            let value_compute_alg_hash = Bits256::from_hex_str(&input.value_compute_alg_hash)
                .map_err(|e| {
                    FuelClientError::InvalidTransaction(format!(
                        "Invalid value compute alg hash: {}",
                        e
                    ))
                })?;
            let r = Bits256::from_hex_str(&input.r)
                .map_err(|e| FuelClientError::InvalidTransaction(format!("Invalid r: {}", e)))?;
            let s = Bits256::from_hex_str(&input.s)
                .map_err(|e| FuelClientError::InvalidTransaction(format!("Invalid s: {}", e)))?;

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
                    "Failed to determine missing contracts: {}",
                    e
                ))
            })?
            .simulate(Execution::state_read_only())
            .await
            .map_err(|e| {
                FuelClientError::ContractCallFailed(format!("Failed to get update fee: {}", e))
            })?;

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
                    "Failed to determine missing contracts: {}",
                    e
                ))
            })?
            .call_params(call_params)
            .map_err(|e| {
                FuelClientError::ContractCallFailed(format!("Failed to set call parameters: {}", e))
            })?
            .call()
            .await
            .map_err(|e| {
                FuelClientError::ContractCallFailed(format!(
                    "Failed to call update_temporal_numeric_values_v1: {}",
                    e
                ))
            })?;

        match tx_response.tx_id {
            Some(tx_id) => Ok(format!("0x{}", tx_id)),
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
                FuelClientError::FuelSdkError(format!("Failed to get wallet balance: {}", e))
            })?;
        Ok(balance as u64)
    }
}

// utility functions

// converts i128 to I128 using 2^127 offset
impl From<i128> for I128 {
    fn from(value: i128) -> Self {
        let offset = 1u128 << 127; // 2^127
        let quantized_with_offset = (value as u128).wrapping_add(offset); // TODO: wrapping add feels scary here
        I128 {
            underlying: quantized_with_offset,
        }
    }
}

// undoes the 2^127 offset to get the original i128 value
impl From<I128> for i128 {
    fn from(value: I128) -> Self {
        let offset = 1u128 << 127; // 2^127
        let quantized_with_offset = (value.underlying as u128).wrapping_sub(offset); // TODO: wrapping sub feels scary here
        quantized_with_offset as i128
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use tokio;

    // Test setup function - fill in your actual values
    async fn setup_test_client() -> Result<FuelClient, FuelClientError> {
        let config = FuelConfig {
            rpc_url: "https://testnet.fuel.network/graphql".to_string(),
            contract_address: "0x1d1360ce59331a1d2774f1cfe86af2885189bb84974caeef0ce780da3e68c502"
                .to_string(),
            private_key: "10a98e108053e466f98d66b246e34c18217f3749b42941fe1c4d20e744142165"
                .to_string(),
        };

        FuelClient::new(config).await
    }

    #[tokio::test]
    async fn test_client_creation() {
        let client = setup_test_client().await;
        match client {
            Ok(_) => println!("✅ Client created successfully"),
            Err(e) => println!("❌ Failed to create client: {}", e),
        }
    }

    #[tokio::test]
    async fn test_get_latest_value_existing() {
        let client = setup_test_client().await.expect("Failed to create client");

        // Test with your BTCUSD asset ID
        let asset_id_hex = "7404e3d104ea7841c3d9e6fd20adfe99b4ad586bc08d8f3bd3afef894cf184de";
        let asset_id_bytes = hex::decode(asset_id_hex).expect("Invalid hex");
        let mut asset_id = [0u8; 32];
        asset_id.copy_from_slice(&asset_id_bytes);

        println!(
            "🔍 Testing get_latest_temporal_numeric_value with asset_id: {}",
            asset_id_hex
        );

        let result = client.get_latest_temporal_numeric_value(asset_id).await;
        match result {
            Ok(Some(value)) => {
                println!(
                    "✅ Got value: timestamp_ns={}, quantized_value={}",
                    value.timestamp_ns, value.quantized_value
                );
            }
            Ok(None) => {
                println!("ℹ️  No value found (returned None)");
            }
            Err(e) => {
                println!("❌ Error getting value: {}", e);
                // Print the full error for debugging
                println!("🔍 Full error: {:?}", e);
            }
        }
    }

    #[tokio::test]
    async fn test_get_latest_value_nonexistent() {
        let client = setup_test_client().await.expect("Failed to create client");

        // Test with a definitely non-existent asset ID (all zeros)
        let asset_id = [0u8; 32];

        println!("🔍 Testing get_latest_temporal_numeric_value with non-existent asset_id");

        let result = client.get_latest_temporal_numeric_value(asset_id).await;
        match result {
            Ok(Some(value)) => {
                println!(
                    "✅ Got value: timestamp_ns={}, quantized_value={}",
                    value.timestamp_ns, value.quantized_value
                );
            }
            Ok(None) => {
                println!("ℹ️  No value found (returned None)");
            }
            Err(e) => {
                println!("❌ Error getting value: {}", e);
                println!("🔍 Full error: {:?}", e);
            }
        }
    }

    #[tokio::test]
    async fn test_wallet_balance() {
        let client = setup_test_client().await.expect("Failed to create client");

        println!("🔍 Testing get_wallet_balance");

        let result = client.get_wallet_balance().await;
        match result {
            Ok(balance) => println!("✅ Wallet balance: {}", balance),
            Err(e) => {
                println!("❌ Error getting wallet balance: {}", e);
                println!("🔍 Full error: {:?}", e);
            }
        }
    }

    #[tokio::test]
    async fn test_contract_call_direct() {
        let client = setup_test_client().await.expect("Failed to create client");

        // Test the contract call directly to see what happens
        let asset_id_hex = "7404e3d104ea7841c3d9e6fd20adfe99b4ad586bc08d8f3bd3afef894cf184de";
        let asset_id_bytes = hex::decode(asset_id_hex).expect("Invalid hex");
        let mut asset_id_array = [0u8; 32];
        asset_id_array.copy_from_slice(&asset_id_bytes);
        let id_bits256 = Bits256(asset_id_array);

        println!("🔍 Testing direct contract call");

        // This is where the panic should occur
        let result = client
            .proxy_contract
            .methods()
            .get_temporal_numeric_value_unchecked_v1(id_bits256)
            .simulate(Execution::state_read_only())
            .await;

        match result {
            Ok(response) => {
                println!("✅ Contract call succeeded");
                println!("📊 Response value: {:?}", response.value);
            }
            Err(e) => {
                println!("❌ Contract call failed: {}", e);
                println!("🔍 Error type: {:?}", std::any::type_name_of_val(&e));
                println!("🔍 Full error: {:?}", e);
            }
        }
    }
}
