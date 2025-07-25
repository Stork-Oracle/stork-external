use std::ffi::{CStr, CString};
use std::os::raw::c_char;
use std::str::FromStr;
use std::sync::Arc;
use tokio::runtime::Runtime;

use fuels::{
    prelude::*,
    crypto::SecretKey,
    types::{AssetId, ContractId, Bits256},
    programs::calls::Execution,
};
use serde::{Deserialize, Serialize};

// Generate the contract bindings from ABI
abigen!(
    Contract(
        name = "StorkContract", 
        abi = "stork_abi.json"
    ),
    Contract(
        name = "ProxyContract",
        abi = "proxy_abi.json"
    )
);

// FFI-compatible structures for JSON serialization
#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct FuelTemporalNumericValue {
    pub timestamp_ns: u64,
    pub quantized_value: String, // I128 as string for JSON serialization
}

#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct FuelTemporalNumericValueInput {
    pub temporal_numeric_value: FuelTemporalNumericValue,
    pub id: String,               // b256 as hex string
    pub publisher_merkle_root: String, // b256 as hex string  
    pub value_compute_alg_hash: String, // b256 as hex string
    pub r: String,               // b256 as hex string
    pub s: String,               // b256 as hex string
    pub v: u8,
}

#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct FuelConfig {
    pub rpc_url: String,
    pub contract_address: String,
    pub private_key: String,
    pub gas_asset_id: String,
}

pub struct FuelClient {
    wallet: WalletUnlocked,
    proxy_contract: ProxyContract<WalletUnlocked>,
    contract: StorkContract<WalletUnlocked>,
    rt: Arc<Runtime>,
    gas_asset_id: AssetId,
}

impl FuelClient {
    pub async fn new(config: FuelConfig) -> std::result::Result<Self, Box<dyn std::error::Error + Send + Sync>> {
        let provider = Provider::connect(&config.rpc_url).await?;
        
        let secret_key = SecretKey::from_str(&config.private_key)?;
        let wallet = WalletUnlocked::new_from_private_key(secret_key, Some(provider.clone()));
        
        // Parse proxy contract address as hex string to ContractId
        let proxy_contract_id = ContractId::from_str(&config.contract_address)?;
        println!("Proxy Contract ID: {}", proxy_contract_id);
        
        let rt = Arc::new(
            tokio::runtime::Builder::new_multi_thread()
                .enable_all()
                .build()?
        );

        // Create proxy contract instance
        let proxy_contract = ProxyContract::new(proxy_contract_id, wallet.clone());
        
        // Get the implementation contract ID from proxy
        let implementation_result = proxy_contract
            .methods()
            .proxy_target()
            .simulate(Execution::StateReadOnly)
            .await?;
            
        let implementation_contract_id = implementation_result.value.1;
        println!("Implementation Contract ID: {}", implementation_contract_id);

        // Create implementation contract instance
        let contract = StorkContract::new(implementation_contract_id, wallet.clone());

        // Parse gas asset ID
        let gas_asset_id = AssetId::from_str(&config.gas_asset_id)?;

        Ok(FuelClient {
            wallet,
            proxy_contract,
            contract,
            rt,
            gas_asset_id,
        })
    }

    pub async fn get_latest_temporal_numeric_value(&self, id: [u8; 32]) -> std::result::Result<Option<FuelTemporalNumericValue>, Box<dyn std::error::Error + Send + Sync>> {
        // Convert [u8; 32] to Bits256
        let id_hex = hex::encode(id);
        let id_bits256 = Bits256::from_hex_str(&format!("0x{}", id_hex))?;

        // Call the implementation contract method
        let result = self.proxy_contract
            .methods()
            .get_temporal_numeric_value_unchecked_v1(id_bits256)
            .with_contracts(&[&self.contract])
            .simulate(Execution::StateReadOnly)
            .await;

        match result {
            Ok(response) => {
                let contract_tnv = response.value;
                let tnv = FuelTemporalNumericValue {
                    timestamp_ns: contract_tnv.timestamp_ns,
                    quantized_value: format!("{}", contract_tnv.quantized_value.underlying as i128),
                };
                Ok(Some(tnv))
            },
            Err(e) => {
                // Log error but don't fail - this might be a feed not found error which is normal
                eprintln!("Error getting temporal numeric value: {}", e);
                Ok(None)
            }
        }
    }

    pub async fn update_temporal_numeric_values(&self, inputs: Vec<FuelTemporalNumericValueInput>) -> std::result::Result<String, Box<dyn std::error::Error + Send + Sync>> {
        // Convert inputs to the generated contract types
        let mut contract_inputs = Vec::new();

        for input in inputs {
            // Parse hex strings to Bits256
            let id = Bits256::from_hex_str(&input.id)?;
            let publisher_merkle_root = Bits256::from_hex_str(&input.publisher_merkle_root)?;
            let value_compute_alg_hash = Bits256::from_hex_str(&input.value_compute_alg_hash)?;
            let r = Bits256::from_hex_str(&input.r)?;
            let s = Bits256::from_hex_str(&input.s)?;

            // Parse quantized value as I128 with proper encoding (matching TypeScript logic)
            let quantized_value_num: i128 = input.temporal_numeric_value.quantized_value.parse()
                .map_err(|e| format!("Failed to parse quantized value: {}", e))?;

            // Create the generated I128 type with 2^127 offset (matching TypeScript stringToI128Input)
            use abigen_bindings::stork_contract_mod::sway_libs::signed_integers::i128::I128;

            // Add 2^127 offset for proper signed integer representation
            let offset = 1u128 << 127; // 2^127
            let quantized_with_offset = (quantized_value_num as u128).wrapping_add(offset);

            let i128_value = I128 {
                underlying: quantized_with_offset,
            };

            // Create the contract input using generated types
            let contract_input = TemporalNumericValueInput {
                temporal_numeric_value: TemporalNumericValue {
                    timestamp_ns: input.temporal_numeric_value.timestamp_ns,
                    quantized_value: i128_value,
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
        let fee_result = self.proxy_contract
            .methods()
            .get_update_fee_v1(contract_inputs.clone())
            .with_contracts(&[&self.contract])
            .simulate(Execution::StateReadOnly)
            .await
            .map_err(|e| {
                eprintln!("Failed to get update fee: {}", e);
                format!("Failed to get update fee: {}", e)
            })?;

        let fee = fee_result.value;

        // Call through proxy contract with payment
        let tx_response = self.proxy_contract
            .methods()
            .update_temporal_numeric_values_v1(contract_inputs.clone())
            .with_contracts(&[&self.contract])
            .call_params(CallParameters::new(fee, self.gas_asset_id, 1_000_000))
            .map_err(|e| {
                eprintln!("Failed to set call parameters (fee: {}, gas_limit: 1000000): {}", fee, e);
                format!("Failed to set call parameters (fee: {}, gas_limit: 1000000): {}", fee, e)
            })?
            .call()
            .await
            .map_err(|e| {
                eprintln!("Contract call failed with {} inputs", contract_inputs.len());
                for (i, input) in contract_inputs.iter().enumerate() {
                    eprintln!("Input {}: id={:?}, timestamp={}, r={:?}, s={:?}, v={}", 
                        i, input.id, input.temporal_numeric_value.timestamp_ns, input.r, input.s, input.v);
                }
                eprintln!("Failed to call update_temporal_numeric_values_v1: {}", e);
                format!("Failed to call update_temporal_numeric_values_v1: {}", e)
            })?;

        Ok(format!("0x{}", tx_response.tx_id.unwrap_or_default()))
    }

    pub async fn get_wallet_balance(&self) -> std::result::Result<u64, Box<dyn std::error::Error + Send + Sync>> {
        let balance = self.wallet.get_asset_balance(&self.gas_asset_id).await?;
        Ok(balance)
    }
}

// FFI functions for Go interop

#[no_mangle]
pub extern "C" fn fuel_client_new(config_json: *const c_char) -> *mut FuelClient {
    if config_json.is_null() {
        return std::ptr::null_mut();
    }

    let config_str = unsafe {
        match CStr::from_ptr(config_json).to_str() {
            Ok(s) => s,
            Err(_) => return std::ptr::null_mut(),
        }
    };

    let config: FuelConfig = match serde_json::from_str(config_str) {
        Ok(c) => c,
        Err(_) => return std::ptr::null_mut(),
    };

    let rt = match tokio::runtime::Builder::new_multi_thread().enable_all().build() {
        Ok(rt) => rt,
        Err(_) => return std::ptr::null_mut(),
    };

    let client = match rt.block_on(FuelClient::new(config)) {
        Ok(c) => c,
        Err(_) => return std::ptr::null_mut(),
    };

    Box::into_raw(Box::new(client))
}

#[no_mangle]
pub extern "C" fn fuel_client_free(client: *mut FuelClient) {
    if !client.is_null() {
        unsafe {
            let _ = Box::from_raw(client);
        }
    }
}

#[no_mangle]
pub extern "C" fn fuel_get_latest_value(
    client: *mut FuelClient,
    id_ptr: *const u8,
) -> *mut c_char {
    if client.is_null() || id_ptr.is_null() {
        return std::ptr::null_mut();
    }

    let client = unsafe { &mut *client };
    
    let id: [u8; 32] = unsafe {
        let mut arr = [0u8; 32];
        std::ptr::copy_nonoverlapping(id_ptr, arr.as_mut_ptr(), 32);
        arr
    };

    match client.rt.block_on(client.get_latest_temporal_numeric_value(id)) {
        Ok(Some(value)) => {
            match serde_json::to_string(&value) {
                Ok(json_str) => {
                    match CString::new(json_str) {
                        Ok(c_str) => c_str.into_raw(),
                        Err(_) => std::ptr::null_mut(),
                    }
                }
                Err(_) => std::ptr::null_mut(),
            }
        }
        _ => std::ptr::null_mut(),
    }
}

static mut LAST_ERROR: Option<String> = None;

#[no_mangle]
pub extern "C" fn fuel_get_last_error() -> *mut c_char {
    unsafe {
        match &LAST_ERROR {
            Some(error) => {
                match CString::new(error.clone()) {
                    Ok(c_str) => c_str.into_raw(),
                    Err(_) => std::ptr::null_mut(),
                }
            }
            None => std::ptr::null_mut(),
        }
    }
}

#[no_mangle]
pub extern "C" fn fuel_update_values(
    client: *mut FuelClient,
    inputs_json: *const c_char,
) -> *mut c_char {
    if client.is_null() || inputs_json.is_null() {
        return std::ptr::null_mut();
    }

    let client = unsafe { &mut *client };
    
    let inputs_str = unsafe {
        match CStr::from_ptr(inputs_json).to_str() {
            Ok(s) => s,
            Err(_) => return std::ptr::null_mut(),
        }
    };

    let inputs: Vec<FuelTemporalNumericValueInput> = match serde_json::from_str(inputs_str) {
        Ok(i) => i,
        Err(_) => return std::ptr::null_mut(),
    };

    // Use catch_unwind to handle panics from the fuels SDK
    let result = std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| {
        client.rt.block_on(client.update_temporal_numeric_values(inputs))
    }));
    
    match result {
        Ok(Ok(tx_hash)) => {
            match CString::new(tx_hash) {
                Ok(c_str) => c_str.into_raw(),
                Err(_) => {
                    eprintln!("Failed to convert transaction hash to C string");
                    std::ptr::null_mut()
                }
            }
        }
        Ok(Err(e)) => {
            let error_msg = format!("{}", e);
            unsafe {
                LAST_ERROR = Some(error_msg.clone());
            }
            if error_msg.contains("UTXO input") && error_msg.contains("was already spent") {
                eprintln!("UTXO_SPENT_ERROR: {}", e);
            } else if error_msg.contains("insufficient funds") || error_msg.contains("InsufficientBalance") {
                eprintln!("Transaction failed due to insufficient balance: {}", e);
            } else if error_msg.contains("InvalidTransaction") || error_msg.contains("invalid") {
                eprintln!("Transaction failed due to invalid parameters: {}", e);
            } else if error_msg.contains("network") || error_msg.contains("connection") || error_msg.contains("timeout") {
                eprintln!("Transaction failed due to network issues: {}", e);
            } else if error_msg.contains("gas") || error_msg.contains("limit") {
                eprintln!("Transaction failed due to gas limit exceeded: {}", e);
            } else {
                eprintln!("Transaction failed with unknown error: {}", e);
            }
            std::ptr::null_mut()
        }
        Err(_) => {
            eprintln!("Panic caught in fuel_update_values - transaction failed due to SDK panic (likely network timeout or contract error)");
            std::ptr::null_mut()
        }
    }
}

#[no_mangle]
pub extern "C" fn fuel_get_wallet_balance(client: *mut FuelClient) -> u64 {
    if client.is_null() {
        eprintln!("fuel_get_wallet_balance called with null client");
        return 0;
    }

    let client = unsafe { &mut *client };

    client.rt.block_on(client.get_wallet_balance()).unwrap_or_else(|e| {
        eprintln!("Failed to get wallet balance: {}", e);
        0
    })
}

#[no_mangle]
pub extern "C" fn fuel_free_string(s: *mut c_char) {
    if !s.is_null() {
        unsafe {
            let _ = CString::from_raw(s);
        }
    }
}