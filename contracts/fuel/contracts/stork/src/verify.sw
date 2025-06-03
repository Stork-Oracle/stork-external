library;

use std::bytes::Bytes;
use std::codec::encode;
use std::hash::*;
use std::ecr::ec_recover;
use std::b512::B512;
use std::option::Option;
use std::u128::U128;

use sway_libs::signed_integers::i128::I128;

pub fn verify_stork_signature(
    storkPubkey: Bytes,
    id: b256,
    recvTime: u64,
    quantized_value: I128,
    publisher_merkle_root: b256,
    value_compute_alg_hash: b256,
    r: b256,
    s: b256,
) -> bool {
    let quantized_value = i128_to_be_bytes(quantized_value);

    let msg_hash = get_stork_message_hash(
        storkPubkey,
        id,
        recvTime,
        quantized_value,
        publisher_merkle_root,
        value_compute_alg_hash,
    );

    let signed_message_hash = get_eth_signed_message_hash(msg_hash);

    return verify_ecdsa_signature(storkPubkey, signed_message_hash, r, s);    
}

fn get_stork_message_hash(
    stork_public_key: Bytes,
    id: b256,
    recv_time: u64,
    quantized_value: [u8; 24],
    publisher_merkle_root: b256,
    value_compute_alg_hash: b256,
) -> b256 {
    let mut data = Bytes::new();
    data.append(Bytes::from(stork_public_key));
    data.append(Bytes::from(encode(id)));
    // left pad with 24 0 bytes
    let mut pad_24 = Bytes::new();
    pad_24.resize(24, 0u8);
    data.append(pad_24);
    data.append(Bytes::from(encode(recv_time)));
    // left pad with 16 0 bytes
    let mut pad_16 = Bytes::new();
    pad_16.resize(16, 0u8);
    data.append(pad_16);
    data.append(Bytes::from(encode(quantized_value)));
    data.append(Bytes::from(encode(publisher_merkle_root)));
    data.append(Bytes::from(encode(value_compute_alg_hash)));

    let hash = keccak256(data);
    return hash;
}

fn get_eth_signed_message_hash(msg_hash: b256) -> b256 {
    let eip_191_prefix = "\x19Ethereum Signed Message:\n32";
    let mut data = Bytes::new();
    data.append(Bytes::from(encode(eip_191_prefix)));
    data.append(Bytes::from(encode(msg_hash)));
    let hash = keccak256(data);
    return hash;
}

fn verify_ecdsa_signature(stork_public_key: Bytes, signed_message_hash: b256, r: b256, s: b256) -> bool {
    let signature = try_get_rs_signature_from_parts(r, s);
    if signature.is_none() {
        return false;
    }
    // safe unwrap
    let signature = signature.unwrap();
    let recovered_pubkey = ec_recover(signature, signed_message_hash);
    if recovered_pubkey.is_err() {
        return false;
    }
    // safe unwrap
    let recovered_pubkey = recovered_pubkey.unwrap();
    let recovered_address = get_eth_address_from_pubkey(recovered_pubkey.into());
    stork_public_key == recovered_address
}

fn try_get_rs_signature_from_parts(r: b256, s: b256) -> Option<B512> {
    let mut signature_bytes = Bytes::new();
    signature_bytes.append(Bytes::from(encode(r)));
    signature_bytes.append(Bytes::from(encode(s)));
    return B512::try_from(signature_bytes);
}

fn get_eth_address_from_pubkey(pubkey: Bytes) -> Bytes {
    let hashed = keccak256(pubkey);
    let mut eth_address = Bytes::from(hashed);
    let (_, address) = eth_address.split_at(12);
    return address;
}

// helper function to convert I128 to [u8; 24]
fn i128_to_be_bytes(value: I128) -> [u8; 24] {
    let mut bytes = [0u8; 24];
    
    // Get the underlying U128 value
    let mut u128_value = value.underlying();
    
    // If the value is greater than indent (positive number)
    // subtract indent to get the actual positive value
    if u128_value > I128::indent() {
        u128_value = u128_value - I128::indent();
    } else if u128_value < I128::indent() {
        // For negative numbers, calculate two's complement
        // First get the absolute value (distance from indent)
        u128_value = I128::indent() - u128_value;
        // Then convert to two's complement
        u128_value = !u128_value + U128::from(1u64);
        // Set the sign bit
        bytes[8] = 0x80;
    }
    
    // Convert U128 to bytes, filling the last 16 bytes
    let mut i = 23;
    while i >= 8 {
        let bytes_val = (u128_value.binary_and(U128::from((0, 255)))).as_u64().unwrap().try_as_u8();
        // safe unwrap
        bytes[i] = bytes_val.unwrap();
        u128_value = u128_value >> 8;
        i -= 1;
    }
    
    bytes
}
