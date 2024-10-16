use std::ptr::hash;
use starknet_core::crypto::{pedersen_hash, Signature};
use starknet_core::types::{FieldElement, FromByteArrayError};

fn bytes_to_field_element(bytes_ptr: *const u8) -> Result<FieldElement, FromByteArrayError> {
        let slice = unsafe { std::slice::from_raw_parts(bytes_ptr, 32) };
        let size32 = <&[u8; 32]>::try_from(slice).unwrap();
        let fe = FieldElement::from_bytes_be(size32);
        return fe;
}

fn write_field_element_to_buffer(field_element: FieldElement, buf_ptr: *mut u8) {
        let output_slice = unsafe { std::slice::from_raw_parts_mut(buf_ptr, 32) };
        output_slice.copy_from_slice(&field_element.to_bytes_be());
}

#[no_mangle]
pub extern "C" fn hash_and_sign(x_ptr: *const u8, y_ptr: *const u8, pk_ptr: *const u8, pedersen_hash_ptr: *mut u8, sig_r_ptr: *mut u8, sig_s_ptr: *mut u8) -> i32 {
        let x_fe = bytes_to_field_element(x_ptr).unwrap_or_else(|_e| {panic!("Failed to convert x byte buffer to field element")});
        let y_fe = bytes_to_field_element(y_ptr).unwrap_or_else(|_e| {panic!("Failed to convert y byte buffer to field element")});
        let pk_fe = bytes_to_field_element(pk_ptr).unwrap_or_else(|_e| {panic!("Failed to convert pk byte buffer to field element")});

        let hashed = pedersen_hash(&x_fe, &y_fe);
        let signature = starknet_core::crypto::ecdsa_sign(&pk_fe, &hashed).unwrap_or_else(|_e| {panic!("Failed to sign pedersen hash")});

        write_field_element_to_buffer(hashed, pedersen_hash_ptr);
        write_field_element_to_buffer(signature.r, sig_r_ptr);
        write_field_element_to_buffer(signature.s, sig_s_ptr);

        return 0
}

#[no_mangle]
pub extern "C" fn validate_stark_signature(x_ptr: *const u8, y_ptr: *const u8, public_key_ptr: *const u8, sig_r_ptr: *const u8, sig_s_ptr: *const u8) -> i32 {
    let x_fe = match bytes_to_field_element(x_ptr) {
        Ok(val) => val,
        Err(_) => return 0,
    };

    let y_fe = match bytes_to_field_element(y_ptr) {
        Ok(val) => val,
        Err(_) => return 0,
    };

    let public_key_fe = match bytes_to_field_element(public_key_ptr) {
        Ok(val) => val,
        Err(_) => return 0,
    };

    let sig_r_fe = match bytes_to_field_element(sig_r_ptr) {
        Ok(val) => val,
        Err(_) => return 0,
    };

    let sig_s_fe = match bytes_to_field_element(sig_s_ptr) {
        Ok(val) => val,
        Err(_) => return 0,
    };

    let hashed = pedersen_hash(&x_fe, &y_fe);
    let signature = Signature { r: sig_r_fe, s: sig_s_fe };

    let result = match starknet_core::crypto::ecdsa_verify(&public_key_fe, &hashed, &signature) {
        Ok(val) => val,
        Err(_) => return 0,
    };

    if result {
        1
    } else {
        0
    }
}
