//! Tests for the signature verification logic.
//!
//! The hash vectors are the same ones asserted by the CosmWasm contract's `verify.rs` tests, so a
//! passing run here means the Cairo implementation reproduces the EVM message encoding exactly.

use stork::verify::{
    get_recoverable_message_hash, get_stork_message_hash, keccak256, verify_stork_evm_signature,
};
use crate::vectors::{
    all_vectors, negative_asset_1, other_public_key, positive_asset_1, stork_public_key,
};

fn verify(
    input: stork_starknet_sdk::interface::TemporalNumericValueInput, signer: starknet::EthAddress,
) -> bool {
    verify_stork_evm_signature(
        signer,
        input.id,
        input.temporal_numeric_value.timestamp_ns,
        input.temporal_numeric_value.quantized_value,
        input.publisher_merkle_root,
        input.value_compute_alg_hash,
        input.r,
        input.s,
        input.v,
    )
}

#[test]
fn test_keccak256_matches_ethereum() {
    let input: ByteArray = "hello";
    assert!(
        keccak256(@input) == 0x1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8,
    );
}

#[test]
fn test_get_stork_message_hash() {
    let hash = get_stork_message_hash(
        0x0a803F9b1CCe32e2773e0d2e98b37E0775cA5d44.try_into().unwrap(),
        0x7404e3d104ea7841c3d9e6fd20adfe99b4ad586bc08d8f3bd3afef894cf184de,
        1722632569208762117,
        62507457175499998000000,
        0xe5ff773b0316059c04aa157898766731017610dcbeede7d7f169bfeaab7cc318,
        0x9be7e9f9ed459417d96112a7467bd0b27575a2c7847195c68f805b70ce1795ba,
    );

    assert!(
        hash == 0x3102baf2e5ad5188e24d56f239915bed3a9a7b51754007dcbf3a65f81bae3084,
        "unexpected message hash: {}",
        hash,
    );
}

#[test]
fn test_get_stork_message_hash_negative_value() {
    // Sign extension of a negative value is the one place the packing can silently diverge from
    // Solidity's `abi.encodePacked(int256)`.
    let hash = get_stork_message_hash(
        0x3db9E960ECfCcb11969509FAB000c0c96DC51830.try_into().unwrap(),
        0x281a649a11eb25eca04f0025c15e99264a056229e722735c7d6c55fef649dfbf,
        1750794968021348308,
        -3020199000000,
        0x5ea4136e8064520a3311961f3f7030dfbc0b96652f46a473e79f2a019b3cd878,
        0x9be7e9f9ed459417d96112a7467bd0b27575a2c7847195c68f805b70ce1795ba,
    );
    let signed = get_recoverable_message_hash(hash);

    // Recovering the documented signer from this digest pins both the hash and the padding.
    assert!(
        verify_stork_evm_signature(
            0x3db9E960ECfCcb11969509FAB000c0c96DC51830.try_into().unwrap(),
            0x281a649a11eb25eca04f0025c15e99264a056229e722735c7d6c55fef649dfbf,
            1750794968021348308,
            -3020199000000,
            0x5ea4136e8064520a3311961f3f7030dfbc0b96652f46a473e79f2a019b3cd878,
            0x9be7e9f9ed459417d96112a7467bd0b27575a2c7847195c68f805b70ce1795ba,
            0x14c36cf7272689cec0335efdc5f82dc2d4b1aceb8d2320d3245e4593df32e696,
            0x79ab437ecd56dc9fcf850f192328840f7f47d5df57cb939d99146b33014c39f0,
            27,
        ),
        "negative value signature rejected, signed hash was {}",
        signed,
    );
}

#[test]
fn test_get_recoverable_message_hash() {
    let hash = get_recoverable_message_hash(
        0x3102baf2e5ad5188e24d56f239915bed3a9a7b51754007dcbf3a65f81bae3084,
    );

    assert!(
        hash == 0xbfaa04ab8f3947f4687a0cb441f673ac3c2233ec3170e37986ff07e09aa50272,
        "unexpected eip-191 hash: {}",
        hash,
    );
}

#[test]
fn test_verify_all_vectors() {
    let key = stork_public_key();
    for input in all_vectors() {
        assert!(verify(input, key), "vector with id {} rejected", input.id);
    }
}

#[test]
fn test_verify_rejects_wrong_signer() {
    assert!(!verify(positive_asset_1(), other_public_key()));
}

#[test]
fn test_verify_rejects_tampered_value() {
    let mut input = positive_asset_1();
    input.temporal_numeric_value.quantized_value += 1;

    assert!(!verify(input, stork_public_key()));
}

#[test]
fn test_verify_rejects_tampered_timestamp() {
    let mut input = positive_asset_1();
    input.temporal_numeric_value.timestamp_ns += 1;

    assert!(!verify(input, stork_public_key()));
}

#[test]
fn test_verify_rejects_tampered_merkle_root() {
    let mut input = negative_asset_1();
    input.publisher_merkle_root += 1;

    assert!(!verify(input, stork_public_key()));
}

#[test]
fn test_verify_rejects_swapped_recovery_id() {
    let mut input = positive_asset_1();
    input.v = 27;

    assert!(!verify(input, stork_public_key()));
}

#[test]
fn test_verify_rejects_non_canonical_recovery_id() {
    // The EVM contract only accepts v of 27 or 28; anything else is rejected outright rather than
    // being normalised, so that a signature cannot be replayed under a second encoding.
    let mut input = positive_asset_1();
    input.v = 0;
    assert!(!verify(input, stork_public_key()));

    input.v = 1;
    assert!(!verify(input, stork_public_key()));

    input.v = 29;
    assert!(!verify(input, stork_public_key()));
}
