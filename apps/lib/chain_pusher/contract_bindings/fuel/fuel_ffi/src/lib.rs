pub mod error;
pub mod fuel_client;

use std::ffi::{CStr, CString};
use std::os::raw::c_char;

use error::{FuelClientError, FuelClientErrorCode};
use fuel_client::{FuelClient, FuelConfig, FuelTemporalNumericValueInput};

// Helper function to handle common FFI patterns
fn handle_ffi_result<T>(
    result: Result<T, FuelClientError>,
    success_handler: impl FnOnce(T) -> Result<(), FuelClientError>,
) -> FuelClientErrorCode {
    match result {
        Ok(value) => match success_handler(value) {
            Ok(()) => FuelClientErrorCode::Success,
            Err(e) => e.into(),
        },
        Err(e) => e.into(),
    }
}

// Helper to safely convert C string to Rust string
fn c_str_to_string(c_str: *const c_char) -> Result<String, FuelClientError> {
    if c_str.is_null() {
        return Err(FuelClientError::InvalidConfig(
            "Null pointer provided".to_string(),
        ));
    }

    unsafe {
        CStr::from_ptr(c_str)
            .to_str()
            .map(|s| s.to_string())
            .map_err(|_| FuelClientError::InvalidConfig("Invalid UTF-8 in input".to_string()))
    }
}

// Helper to create C string from Rust string
fn string_to_c_char(s: String) -> Result<*mut c_char, FuelClientError> {
    CString::new(s)
        .map(|c_str| c_str.into_raw())
        .map_err(|_| FuelClientError::SystemError("Failed to create C string".to_string()))
}

#[no_mangle]
pub extern "C" fn fuel_client_new(config_json: *const c_char) -> *mut FuelClient {
    let result = (|| -> Result<FuelClient, FuelClientError> {
        let config_str = c_str_to_string(config_json)?;
        let config: FuelConfig = serde_json::from_str(&config_str)?;

        // Create a temporary runtime just for client creation
        let rt = tokio::runtime::Builder::new_multi_thread()
            .enable_all()
            .build()
            .map_err(|e| {
                FuelClientError::SystemError(format!("Failed to create runtime: {}", e))
            })?;

        rt.block_on(FuelClient::new(config))
    })();

    match result {
        Ok(client) => Box::into_raw(Box::new(client)),
        Err(_) => std::ptr::null_mut(), // Error info lost here, but that's the nature of this pattern
    }
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
    out_value_json: *mut *mut c_char,
) -> FuelClientErrorCode {
    if out_value_json.is_null() {
        return FuelClientErrorCode::InvalidConfig;
    }

    // Initialize output to null
    unsafe {
        *out_value_json = std::ptr::null_mut();
    }

    let result = (|| -> Result<Option<String>, FuelClientError> {
        if client.is_null() || id_ptr.is_null() {
            return Err(FuelClientError::NullClient);
        }

        let client = unsafe { &mut *client };

        let id: [u8; 32] = unsafe {
            let mut arr = [0u8; 32];
            std::ptr::copy_nonoverlapping(id_ptr, arr.as_mut_ptr(), 32);
            arr
        };

        let value_opt = client
            .rt
            .block_on(client.get_latest_temporal_numeric_value(id))
            .map_err(|e| {
                println!("Error: {:?}", e);
                e
            })?;

        match value_opt {
            Some(value) => {
                let json_str = serde_json::to_string(&value)?;
                Ok(Some(json_str))
            }
            None => Ok(None), // No value found - this is success, not error
        }
    })();

    handle_ffi_result(result, |value_opt| {
        if let Some(json_str) = value_opt {
            let c_str = string_to_c_char(json_str)?;
            unsafe {
                *out_value_json = c_str;
            }
        }
        Ok(())
    })
}

#[no_mangle]
pub extern "C" fn fuel_update_values(
    client: *mut FuelClient,
    inputs_json: *const c_char,
    out_tx_hash: *mut *mut c_char,
) -> FuelClientErrorCode {
    if out_tx_hash.is_null() {
        return FuelClientErrorCode::InvalidConfig;
    }

    // Initialize output to null
    unsafe {
        *out_tx_hash = std::ptr::null_mut();
    }

    let result = (|| -> Result<String, FuelClientError> {
        if client.is_null() {
            return Err(FuelClientError::InvalidConfig(
                "Null client pointer".to_string(),
            ));
        }

        let client = unsafe { &mut *client };
        let inputs_str = c_str_to_string(inputs_json)?;
        let inputs: Vec<FuelTemporalNumericValueInput> = serde_json::from_str(&inputs_str)?;

        client
            .rt
            .block_on(client.update_temporal_numeric_values(inputs))
    })();

    handle_ffi_result(result, |tx_hash| {
        let c_str = string_to_c_char(tx_hash)?;
        unsafe {
            *out_tx_hash = c_str;
        }
        Ok(())
    })
}

#[no_mangle]
pub extern "C" fn fuel_get_wallet_balance(
    client: *mut FuelClient,
    out_balance: *mut u64,
) -> FuelClientErrorCode {
    if out_balance.is_null() {
        return FuelClientErrorCode::InvalidConfig;
    }

    // Initialize output to 0
    unsafe {
        *out_balance = 0;
    }

    let result = (|| -> Result<u64, FuelClientError> {
        if client.is_null() {
            return Err(FuelClientError::InvalidConfig(
                "Null client pointer".to_string(),
            ));
        }

        let client = unsafe { &mut *client };
        client.rt.block_on(client.get_wallet_balance())
    })();

    handle_ffi_result(result, |balance| {
        unsafe {
            *out_balance = balance;
        }
        Ok(())
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
