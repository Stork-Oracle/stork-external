use std::ffi::{c_char, CStr};
use starknet_core::crypto::pedersen_hash;
use starknet_core::types::{FieldElement, FromByteArrayError};
use eth_encode_packed::ethabi::ethereum_types::U256;

fn bytes_to_field_element(bytes_ptr: *const u8) -> Result<FieldElement, FromByteArrayError> {
        let slice = unsafe { std::slice::from_raw_parts(bytes_ptr, 32) };
        let size32 = <&[u8; 32]>::try_from(slice).unwrap();
        let fe = FieldElement::from_bytes_be(size32);
        return fe;
}

fn bytes_to_u256(bytes_ptr: *const u8) -> U256 {
        let slice = unsafe { std::slice::from_raw_parts(bytes_ptr, 32) };
        U256::from_big_endian(slice)
}

fn write_field_element_to_buffer(field_element: FieldElement, buf_ptr: *mut u8) {
        let output_slice = unsafe { std::slice::from_raw_parts_mut(buf_ptr, 32) };
        output_slice.copy_from_slice(&field_element.to_bytes_be());
}

#[no_mangle]
pub extern "C" fn hash_and_sign(
        asset_hex_padded: *const c_char,
        quantized_price: *const c_char,
        timestamp_ns: &i64,
        oracle_name_int_ptr: *const u8,
        pk_ptr: *const u8,
        pedersen_hash_ptr: *mut u8,
        sig_r_ptr: *mut u8,
        sig_s_ptr: *mut u8
) -> i32 {

        let asset_hex_padded_str = unsafe {
                CStr::from_ptr(asset_hex_padded).to_str().unwrap_or("Invalid UTF-8")
        };
        let quantized_price_str = unsafe {
                CStr::from_ptr(quantized_price).to_str().unwrap_or("Invalid UTF-8")
        };
        let asset_u256 = U256::from_str_radix(asset_hex_padded_str, 16).unwrap_or_else(|e| {panic!("Failed to convert padded asset hex to u256")});

        // convert price and timestamp to numbers
        let price_u256 = U256::from_dec_str(quantized_price_str).unwrap_or_else(|e| {panic!("Failed to convert quantized price to U256")});
        let timestamp_u256 = U256::from(timestamp_ns/1_000_000_000);

        // combine (asset + oracle name), (price + timestamp) into x and y
        let x_u256 = (asset_u256 << 40) + bytes_to_u256(oracle_name_int_ptr);
        let y_u256 = (price_u256 << 32) + timestamp_u256;

        // convert our U256 numbers into FieldElement's for hashing and signing
        fn u256_to_field_element(u256_val: &U256) -> Result<FieldElement, FromByteArrayError> {
                let mut buf = [0u8; 32];
                u256_val.to_big_endian(&mut buf);
                return FieldElement::from_bytes_be(&buf);
        }
        let x_fe = u256_to_field_element(&x_u256).unwrap_or_else(|e| {panic!("Failed to convert x int to field element")});
        let y_fe = u256_to_field_element(&y_u256).unwrap_or_else(|e| {panic!("Failed to convert y int to field element")});

        let pk_fe = bytes_to_field_element(pk_ptr).unwrap_or_else(|e| {panic!("Failed to convert pk byte buffer to field element")});

        let hashed = pedersen_hash(&x_fe, &y_fe);
        let signature = starknet_core::crypto::ecdsa_sign(&pk_fe, &hashed).unwrap_or_else(|e| {panic!("Failed to sign pedersen hash")});

        write_field_element_to_buffer(hashed, pedersen_hash_ptr);
        write_field_element_to_buffer(signature.r, sig_r_ptr);
        write_field_element_to_buffer(signature.s, sig_s_ptr);

        return 0
}