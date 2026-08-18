//! Provides [`verify_stork_evm_signature`], implementing the signature verification logic for
//! Stork updates.
//!
//! Stork signs updates with a secp256k1 key over an EIP-191 prefixed keccak256 digest, so this
//! module reproduces the EVM contract's `getStorkMessageHashV1` byte-for-byte and then defers to
//! the `secp256k1` syscalls exposed by the Starknet corelib.

use core::integer::u128_byte_reverse;
use core::keccak::compute_keccak_byte_array;
use starknet::EthAddress;
use starknet::eth_signature::is_eth_signature_valid;
use starknet::secp256_trait::Signature;
use stork_starknet_sdk::temporal_numeric_value::{EncodedAssetId, twos_complement};

/// Sixteen 0xFF bytes, used to sign-extend a negative quantized value to 32 bytes.
const SIGN_EXTENSION: felt252 = 0xffffffffffffffffffffffffffffffff;

/// Verifies the EVM signature for a Stork update from the provided update parameters.
pub fn verify_stork_evm_signature(
    stork_evm_public_key: EthAddress,
    id: EncodedAssetId,
    recv_time: u64,
    quantized_value: i128,
    publisher_merkle_root: u256,
    value_compute_alg_hash: u256,
    r: u256,
    s: u256,
    v: u8,
) -> bool {
    // The EVM contract only accepts the canonical recovery ids.
    let y_parity = if v == 28 {
        true
    } else if v == 27 {
        false
    } else {
        return false;
    };

    let message = get_stork_message_hash(
        stork_evm_public_key,
        id,
        recv_time,
        quantized_value,
        publisher_merkle_root,
        value_compute_alg_hash,
    );
    let signed_message = get_recoverable_message_hash(message);

    is_eth_signature_valid(signed_message, Signature { r, s, y_parity }, stork_evm_public_key)
        .is_ok()
}

/// Reproduces `keccak256(abi.encodePacked(storkPubKey, id, recvTime, quantizedValue,
/// publisherMerkleRoot, valueComputeAlgHash))` from `StorkVerify.sol`.
///
/// Solidity packs `uint256 recvTime` and `int256 quantizedValue` as 32 bytes each, so the 64-bit
/// timestamp is left padded with 24 zero bytes and the 128-bit value is sign-extended to 32 bytes.
pub fn get_stork_message_hash(
    stork_evm_public_key: EthAddress,
    id: EncodedAssetId,
    recv_time: u64,
    quantized_value: i128,
    publisher_merkle_root: u256,
    value_compute_alg_hash: u256,
) -> u256 {
    let mut data: ByteArray = Default::default();

    data.append_word(stork_evm_public_key.into(), 20);
    append_u256(ref data, id);

    data.append_word(0, 24);
    data.append_word(recv_time.into(), 8);

    if quantized_value < 0 {
        data.append_word(SIGN_EXTENSION, 16);
    } else {
        data.append_word(0, 16);
    }
    data.append_word(twos_complement(quantized_value), 16);

    append_u256(ref data, publisher_merkle_root);
    append_u256(ref data, value_compute_alg_hash);

    keccak256(@data)
}

/// Applies the EIP-191 personal-sign prefix to a 32 byte message, matching
/// OpenZeppelin's `MessageHashUtils.toEthSignedMessageHash`.
pub fn get_recoverable_message_hash(message: u256) -> u256 {
    let mut data: ByteArray = "\x19Ethereum Signed Message:\n32";
    append_u256(ref data, message);

    keccak256(@data)
}

/// Ethereum-compatible keccak256.
///
/// The corelib returns the digest with its bytes reversed relative to Ethereum's convention, so
/// the result is byte-reversed back before being compared against EVM-produced hashes.
pub fn keccak256(data: @ByteArray) -> u256 {
    let hash = compute_keccak_byte_array(data);

    u256 { low: u128_byte_reverse(hash.high), high: u128_byte_reverse(hash.low) }
}

/// Appends a `u256` to `out` as 32 big-endian bytes.
fn append_u256(ref out: ByteArray, value: u256) {
    out.append_word(value.high.into(), 16);
    out.append_word(value.low.into(), 16);
}
