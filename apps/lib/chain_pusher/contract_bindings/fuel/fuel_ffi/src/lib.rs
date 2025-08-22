pub mod error;
pub mod fuel_client;

use std::ffi::{CStr, CString};
use std::os::raw::c_char;

use error::{FuelClientError, FuelClientStatus};
use fuel_client::{FuelClient, FuelConfig, FuelTemporalNumericValueInput};

#[no_mangle]
pub extern "C" fn fuel_client_new(
    config_json: *const c_char,
    out_client: *mut *mut FuelClient,
    out_error: *mut *mut c_char,
) -> FuelClientStatus {
    if out_client.is_null() {
        return FuelClientError::NullPointer("out_client is null".to_string()).into();
    }
    if config_json.is_null() {
        return FuelClientError::NullPointer("config_json is null".to_string()).into();
    }

    // Initialize output to null
    unsafe {
        *out_client = std::ptr::null_mut();
    }

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

    handle_ffi_result(result, |client| {
        unsafe {
            *out_client = Box::into_raw(Box::new(client));
        }
        Ok(())
    }, out_error)
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
    out_error: *mut *mut c_char,
) -> FuelClientStatus {
    if out_value_json.is_null() {
        return FuelClientError::NullPointer("out_value_json is null".to_string()).into();
    }
    if client.is_null() {
        return FuelClientError::NullPointer("client is null".to_string()).into();
    }
    if id_ptr.is_null() {
        return FuelClientError::NullPointer("id_ptr is null".to_string()).into();
    }

    // Initialize output to null
    unsafe {
        *out_value_json = std::ptr::null_mut();
    }

    let result = (|| -> Result<Option<String>, FuelClientError> {
        let client = unsafe { &mut *client };

        let id: [u8; 32] = unsafe {
            let mut arr = [0u8; 32];
            std::ptr::copy_nonoverlapping(id_ptr, arr.as_mut_ptr(), 32);
            arr
        };

        let result= std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| {
            client.rt.block_on(client.get_latest_temporal_numeric_value(id))
        })).map_err(|e| FuelClientError::SystemError(format!("Panic in get_latest_temporal_numeric_value: {:?}", e)))?;

        let value_opt = result?;

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
    }, out_error)
}

#[no_mangle]
pub extern "C" fn fuel_update_values(
    client: *mut FuelClient,
    inputs_json: *const c_char,
    out_tx_hash: *mut *mut c_char,
    out_error: *mut *mut c_char,
) -> FuelClientStatus {
    if out_tx_hash.is_null() {
        return FuelClientError::NullPointer("out_tx_hash is null".to_string()).into();
    }
    if client.is_null() {
        return FuelClientError::NullPointer("client is null".to_string()).into();
    }
    if inputs_json.is_null() {
        return FuelClientError::NullPointer("inputs_json is null".to_string()).into();
    }

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

        std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| {
            client
                .rt
                .block_on(client.update_temporal_numeric_values(inputs))
        })).map_err(|e| FuelClientError::SystemError(format!("Panic in update_temporal_numeric_values: {:?}", e)))?
    })();

    handle_ffi_result(result, |tx_hash| {
        let c_str = string_to_c_char(tx_hash)?;
        unsafe {
            *out_tx_hash = c_str;
        }
        Ok(())
    }, out_error)
}

#[no_mangle]
pub extern "C" fn fuel_get_wallet_balance(
    client: *mut FuelClient,
    out_balance: *mut u64,
    out_error: *mut *mut c_char,
) -> FuelClientStatus {
    if out_balance.is_null() {
        return FuelClientError::NullPointer("out_balance is null".to_string()).into();
    }
    if client.is_null() {
        return FuelClientError::NullPointer("client is null".to_string()).into();
    }

    unsafe {
        *out_balance = 0;
    }

    let result = (|| -> Result<u64, FuelClientError> {
        let client = unsafe { &mut *client };
        
        std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| {
            client.rt.block_on(client.get_wallet_balance())
        })).map_err(|e| FuelClientError::SystemError(format!("Panic in get_wallet_balance: {:?}", e)))?
    })();

    handle_ffi_result(result, |balance| {
        unsafe {
            *out_balance = balance;
        }
        Ok(())
    }, out_error)
}

#[no_mangle]
pub extern "C" fn fuel_free_string(s: *mut c_char) {
    if !s.is_null() {
        unsafe {
            let _ = CString::from_raw(s);
        }
    }
}

// Helper functions

fn handle_ffi_result<T>(
    result: Result<T, FuelClientError>,
    success_handler: impl FnOnce(T) -> Result<(), FuelClientError>,
    out_error: *mut *mut c_char,
) -> FuelClientStatus {
    let final_result =match result {
        Ok(value) => success_handler(value),
        Err(e) => Err(e),
    };

    match final_result {
        Ok(()) => {
            if !out_error.is_null() {
                unsafe {
                    *out_error = std::ptr::null_mut();
                }
            }
            FuelClientStatus::Success
        }
        Err(e) => {
            if !out_error.is_null() {
                if let Ok(error_str) = string_to_c_char(e.to_string()) {
                    unsafe {
                        *out_error = error_str;
                    }
                }
            }
            e.into()
        }
    }
}

fn c_str_to_string(c_str: *const c_char) -> Result<String, FuelClientError> {
    if c_str.is_null() {
        return Err(FuelClientError::NullPointer("c_str is null".to_string()));
    }

    let c_str = unsafe { CStr::from_ptr(c_str) };
    c_str
        .to_str()
        .map(|s| s.to_owned())
        .map_err(|e| FuelClientError::SystemError(format!("Invalid UTF-8 in input: {}", e)))
}

fn string_to_c_char(s: String) -> Result<*mut c_char, FuelClientError> {
    CString::new(s)
        .map(|c_str| c_str.into_raw())
        .map_err(|e| FuelClientError::SystemError(format!("Failed to create C string: {}", e)))
}
