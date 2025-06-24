library;

use std::bytes::Bytes;
use std::codec::encode;
use std::hash::*;
use std::crypto::secp256k1::Secp256k1;
use std::crypto::message::Message;
use std::u128::U128;
use std::vm::evm::evm_address::EvmAddress;

use sway_libs::signed_integers::i128::I128;

pub fn verify_stork_signature(
    stork_pubkey: EvmAddress,
    id: b256,
    recv_time: u64,
    quantized_value: I128,
    publisher_merkle_root: b256,
    value_compute_alg_hash: b256,
    r: b256,
    s: b256,
    v: u8,
) -> bool {
    let quantized_value = i128_to_be_bytes(quantized_value);

    let msg_hash = get_stork_message_hash(
        stork_pubkey,
        id,
        recv_time,
        quantized_value,
        publisher_merkle_root,
        value_compute_alg_hash,
    );

    let signed_message_hash = get_eth_signed_message_hash(msg_hash);

    let signature = match try_get_rsv_signature_from_parts(r, s, v) {
        Some(sig) => sig,
        None => return false,
    };

    match signature.verify_evm_address(stork_pubkey, signed_message_hash) {
        Ok(_) => true,
        Err(_) => false,
    }
}

fn get_stork_message_hash(
    stork_public_key: EvmAddress,
    id: b256,
    recv_time: u64,
    quantized_value: Bytes,
    publisher_merkle_root: b256,
    value_compute_alg_hash: b256,
) -> b256 {
    let mut data = Bytes::new();
    data.append(stork_public_key.into());
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
    data.append(quantized_value);
    data.append(Bytes::from(encode(publisher_merkle_root)));
    data.append(Bytes::from(encode(value_compute_alg_hash)));

    let hash = keccak256(data);
    return hash;
}

fn get_eth_signed_message_hash(msg_hash: b256) -> Message {
    // Construct prefix bytes
    let mut padded_prefix_bytes = Bytes::from(0x0000000019457468657265756D205369676E6564204D6573736167653A0A3332);
    let (_, prefix_bytes) = padded_prefix_bytes.split_at(4);
    let mut data = Bytes::new();
    data.append(prefix_bytes);
    data.append(Bytes::from(msg_hash));
    let hash = keccak256(data);
    Message::from(hash)
}

fn try_get_rsv_signature_from_parts(r: b256, s: b256, v: u8) -> Option<Secp256k1> {
    // make most significant bit of s 0 or 1 depending on v
    match v {
        27 => {
            // Create a b256 mask with all 1s except the highest bit
            let mask = b256::max() & !(b256::from(0x8000000000000000000000000000000000000000000000000000000000000000));
            let s_cleared = s & mask;
            Some(Secp256k1::from((r, s_cleared)))
        }
        28 => {
            // Create a b256 mask with only the highest bit set
            let mask = b256::from(0x8000000000000000000000000000000000000000000000000000000000000000);
            let s_set = s | mask;
            Some(Secp256k1::from((r, s_set)))
        }
        _ => None,
    }
}

// helper function to convert I128 to Bytes
fn i128_to_be_bytes(value: I128) -> Bytes {
    // fast track for zero
    if value == I128::zero() {
        let mut bytes = Bytes::new();
        bytes.resize(16, 0u8);
        return bytes;
    }

    let mut bytes = [0u8; 16]; // 16 bytes
    let mut u128_value = value.underlying();

    if u128_value > I128::indent() {
        // Positive number
        u128_value = u128_value - I128::indent();
    } else if u128_value < I128::indent() {
        // Negative number - use two's complement
        let magnitude = I128::indent() - u128_value;
        // Create a mask of all 1's using NOT of zero
        let mask = !U128::from(0u64);
        // XOR implementation using available operations: (a | b) & !(a & b)
        let all_ones = mask;
        let magnitude_minus_one = magnitude - U128::from(1u64);
        let or_result = magnitude_minus_one.binary_or(all_ones);
        let and_result = magnitude_minus_one.binary_and(all_ones);
        u128_value = or_result.binary_and(!and_result);
    }
    // Convert U128 to bytes, filling all 16 bytes
    let mut i = 16;
    while i > 0 {
        i -= 1;
        bytes[i] = (u128_value.binary_and(U128::from((0, 255)))).as_u64().unwrap().try_as_u8().unwrap();
        u128_value = u128_value >> 8;
    }
    return Bytes::from(encode(bytes));
}

// Tests

#[test]
fn test_verify_stork_signature() {
    // construct stork pubkey
    let mut padded_evm_address_bytes = Bytes::from(0x0000000000000000000000000a803F9b1CCe32e2773e0d2e98b37E0775cA5d44);
    let (_, evm_address_bytes) = padded_evm_address_bytes.split_at(12);
    let stork_pubkey = EvmAddress::try_from(evm_address_bytes).unwrap();

    // construct quantized value
    let quantized_value_u64 = 62507457175499998u64;
    let mut quantized_value_u128 = U128::from(quantized_value_u64);
    let quantized_value_u128 = quantized_value_u128 * U128::from(1000000u64);
    let quantized_value = I128::try_from(quantized_value_u128).unwrap();

    // other vars that don't need multi step construction
    let id = b256::from(0x7404e3d104ea7841c3d9e6fd20adfe99b4ad586bc08d8f3bd3afef894cf184de);
    let recv_time = 1722632569208762117;
    let publisher_merkle_root = b256::from(0xe5ff773b0316059c04aa157898766731017610dcbeede7d7f169bfeaab7cc318);
    let value_compute_alg_hash = b256::from(0x9be7e9f9ed459417d96112a7467bd0b27575a2c7847195c68f805b70ce1795ba);
    let r = b256::from(0xb9b3c9f80a355bd0cd6f609fff4a4b15fa4e3b4632adabb74c020f5bcd240741);
    let s = b256::from(0x16fab526529ac795108d201832cff8c2d2b1c710da6711fe9f7ab288a7149758);
    let v = 28;
    let result = verify_stork_signature(
        stork_pubkey,
        id,
        recv_time,
        quantized_value,
        publisher_merkle_root,
        value_compute_alg_hash,
        r,
        s,
        v,
    );
    assert(result == true);
}

#[test]
fn test_get_stork_message_hash() {
    let mut padded_evm_address_bytes = Bytes::from(0x0000000000000000000000000a803F9b1CCe32e2773e0d2e98b37E0775cA5d44);
    let (_, evm_address_bytes) = padded_evm_address_bytes.split_at(12);
    let stork_pubkey = EvmAddress::try_from(evm_address_bytes).unwrap();

    // construct quantized value
    let quantized_value_u64 = 62507457175499998u64;
    let mut quantized_value_u128 = U128::from(quantized_value_u64);
    let quantized_value_u128 = quantized_value_u128 * U128::from(1000000u64);
    let quantized_value = I128::try_from(quantized_value_u128).unwrap();
    let quantized_value_bytes = i128_to_be_bytes(quantized_value);

    // other vars that don't need multi step construction
    let id = b256::from(0x7404e3d104ea7841c3d9e6fd20adfe99b4ad586bc08d8f3bd3afef894cf184de);
    let recv_time = 1722632569208762117;
    let publisher_merkle_root = b256::from(0xe5ff773b0316059c04aa157898766731017610dcbeede7d7f169bfeaab7cc318);
    let value_compute_alg_hash = b256::from(0x9be7e9f9ed459417d96112a7467bd0b27575a2c7847195c68f805b70ce1795ba);

    let message_hash = get_stork_message_hash(
        stork_pubkey,
        id,
        recv_time,
        quantized_value_bytes,
        publisher_merkle_root,
        value_compute_alg_hash,
    );

    assert(
        message_hash == b256::from(0x3102baf2e5ad5188e24d56f239915bed3a9a7b51754007dcbf3a65f81bae3084),
    );
}

#[test]
fn test_get_eth_signed_message_hash() {
    let msg_hash = b256::from(0x3102baf2e5ad5188e24d56f239915bed3a9a7b51754007dcbf3a65f81bae3084);
    let signed_message_hash = get_eth_signed_message_hash(msg_hash);
    let expected_message = Message::from(b256::from(0xbfaa04ab8f3947f4687a0cb441f673ac3c2233ec3170e37986ff07e09aa50272));
    assert(signed_message_hash.bytes() == expected_message.bytes());
}

#[test]
fn test_i128_to_be_bytes_zero() {
    let value: I128 = I128::zero();
    let bytes = i128_to_be_bytes(value);
    let mut expected_bytes = Bytes::new();
    expected_bytes.resize(16, 0u8);

    assert(bytes == expected_bytes);
}

#[test]
fn test_i128_to_be_bytes_positive() {
    let value_u64 = 1u64;
    let value_u128 = U128::from(value_u64);
    let value = I128::try_from(value_u128).unwrap();
    let bytes = i128_to_be_bytes(value);
    let mut expected_bytes = Bytes::new();
    expected_bytes.resize(16, 0u8);
    expected_bytes.set(15, 1u8);
    assert(bytes == expected_bytes);

    let value = I128::from(I128::MAX);
    let bytes = i128_to_be_bytes(value);
    let mut max_expected_bytes = Bytes::new();
    max_expected_bytes.resize(16, 0xFFu8);
    max_expected_bytes.set(0, 0x7Fu8);
    assert(bytes == max_expected_bytes);
}

#[test]
fn test_try_get_rsv_signature_from_parts() {
    let r = b256::from(0xb9b3c9f80a355bd0cd6f609fff4a4b15fa4e3b4632adabb74c020f5bcd240741);
    let s = b256::from(0x16fab526529ac795108d201832cff8c2d2b1c710da6711fe9f7ab288a7149758);
    let v = 28;
    let signature = try_get_rsv_signature_from_parts(r, s, v);
    assert(signature.is_some());
    let sig = signature.unwrap();
    let sig_bytes: Bytes = sig.into();

    let mut expected_bytes = Bytes::new();
    expected_bytes.append(Bytes::from(r));
    expected_bytes.append(Bytes::from(
        s
        | b256::from(0x8000000000000000000000000000000000000000000000000000000000000000),
    ));

    assert(sig_bytes == expected_bytes);
}

#[test]
fn test_i128_to_be_bytes_negative() {
    let value = I128::from(I128::MIN);
    let bytes = i128_to_be_bytes(value);
    let mut min_expected_bytes = Bytes::new();
    min_expected_bytes.resize(16, 0x00u8);
    min_expected_bytes.set(0, 0x80u8);

    assert(bytes == min_expected_bytes);
}
