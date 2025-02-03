//! Provides the [`verify_stork_evm_signature`] function implementing the signature verification logic for verifying updates.
use sylvia::cw_std::{Api, StdResult};
use tiny_keccak::{Hasher, Keccak};

/// The type for EVM public keys. This is an alias for a 20 byte array.
pub type EvmPubkey = [u8; 20];

/// Verifies the EVM signature for a Stork update from the provided update parameters.
pub fn verify_stork_evm_signature(
    api: &dyn Api,
    stork_evm_public_key: &EvmPubkey,
    id: [u8; 32],
    recv_time: u64,
    quantized_value: i128,
    publisher_merkle_root: [u8; 32],
    value_compute_alg_hash: [u8; 32],
    r: [u8; 32],
    s: [u8; 32],
    v: u8,
) -> StdResult<bool> {
    let message = get_stork_message_hash(
        stork_evm_public_key,
        id,
        recv_time,
        quantized_value,
        publisher_merkle_root,
        value_compute_alg_hash,
    );
    let signature = get_rsv_signature_from_parts(&r, &s, v);
    verify_ecdsa_signature(api, stork_evm_public_key, &message, &signature)
}

fn verify_ecdsa_signature(
    api: &dyn Api,
    pubkey: &EvmPubkey,
    message: &[u8],
    signature: &[u8],
) -> StdResult<bool> {
    if signature.len() != 65 {
        return Ok(false);
    }

    let message_hash = get_recoverable_message_hash(message);
    let recovered_pubkey = match recover_secp256k1_pubkey(api, &message_hash, signature)? {
        Some(pk) => pk,
        None => return Ok(false),
    };

    let eth_pubkey = get_eth_pubkey(&recovered_pubkey);
    println!("eth_pubkey: {:?}", eth_pubkey);
    println!("pubkey: {:?}", pubkey);
    Ok(eth_pubkey == *pubkey)
}

fn get_stork_message_hash(
    stork_evm_public_key: &EvmPubkey,
    id: [u8; 32],
    recv_time: u64,
    quantized_value: i128,
    publisher_merkle_root: [u8; 32],
    value_compute_alg_hash: [u8; 32],
) -> [u8; 32] {
    let mut data: Vec<u8> = Vec::new();
    data.extend_from_slice(stork_evm_public_key);
    data.extend_from_slice(&id);
    data.extend_from_slice(&[0u8; 24]); // Left pad with 24 zero bytes
    data.extend_from_slice(&recv_time.to_be_bytes());
    data.extend_from_slice(&[0u8; 16]); // Left pad with 16 zero bytes
    data.extend_from_slice(&quantized_value.to_be_bytes());
    data.extend_from_slice(&publisher_merkle_root);
    data.extend_from_slice(&value_compute_alg_hash);
    let mut hasher = Keccak::v256();
    hasher.update(&data);
    let mut hash_output = [0u8; 32];
    hasher.finalize(&mut hash_output);
    return hash_output;
}

fn get_recoverable_message_hash(message: &[u8]) -> [u8; 32] {
    let eip_191_prefix = format!("\x19Ethereum Signed Message:\n{}", message.len());

    let mut data: Vec<u8> = Vec::new();
    data.extend_from_slice(eip_191_prefix.as_bytes());
    data.extend_from_slice(message);

    let mut hasher = Keccak::v256();
    hasher.update(&data);
    let mut hash_output = [0u8; 32];
    hasher.finalize(&mut hash_output);
    return hash_output;
}

fn get_rsv_signature_from_parts(r: &[u8], s: &[u8], v: u8) -> [u8; 65] {
    let mut signature = [0; 65];
    signature[..32].copy_from_slice(r);
    signature[32..64].copy_from_slice(s);
    signature[64] = v;
    signature
}

fn recover_secp256k1_pubkey(
    api: &dyn Api,
    message: &[u8],
    signature: &[u8],
) -> StdResult<Option<Vec<u8>>> {
    let rs = &signature[..64];
    let v = signature[64];
    let recovery_id = match v {
        27 => 0,
        28 => 1,
        _ => return Ok(None),
    };

    match api.secp256k1_recover_pubkey(message, rs, recovery_id) {
        Ok(pk) => Ok(Some(pk)),
        Err(_) => Ok(None),
    }
}

fn get_eth_pubkey(pubkey: &[u8]) -> EvmPubkey {
    let mut hasher = Keccak::v256();
    hasher.update(&pubkey[1..]);
    let mut hash = [0u8; 32];
    hasher.finalize(&mut hash);

    // Take last 20 bytes to get ethereum address
    let mut eth_pubkey = [0u8; 20];
    eth_pubkey.copy_from_slice(&hash[12..32]);
    eth_pubkey
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::multitest::hex_to_bytes;
    use sylvia::cw_std::testing::mock_dependencies;

    #[test]
    fn test_verify_stork_evm_signature() {
        let deps = mock_dependencies();
        let api = deps.api;

        let stork_public_key = hex_to_bytes("0a803F9b1CCe32e2773e0d2e98b37E0775cA5d44")[..20]
            .try_into()
            .unwrap();
        let id = hex_to_bytes("7404e3d104ea7841c3d9e6fd20adfe99b4ad586bc08d8f3bd3afef894cf184de")
            [..32]
            .try_into()
            .unwrap();
        let recv_time = 1722632569208762117;
        let quantized_value = 62507457175499998000000;
        let publisher_merkle_root =
            hex_to_bytes("e5ff773b0316059c04aa157898766731017610dcbeede7d7f169bfeaab7cc318")[..32]
                .try_into()
                .unwrap();
        let value_compute_alg_hash =
            hex_to_bytes("9be7e9f9ed459417d96112a7467bd0b27575a2c7847195c68f805b70ce1795ba")[..32]
                .try_into()
                .unwrap();
        let r = hex_to_bytes("b9b3c9f80a355bd0cd6f609fff4a4b15fa4e3b4632adabb74c020f5bcd240741")
            [..32]
            .try_into()
            .unwrap();
        let s = hex_to_bytes("16fab526529ac795108d201832cff8c2d2b1c710da6711fe9f7ab288a7149758")
            [..32]
            .try_into()
            .unwrap();
        let v = 28;

        let result = verify_stork_evm_signature(
            &api,
            &stork_public_key,
            id,
            recv_time,
            quantized_value,
            publisher_merkle_root,
            value_compute_alg_hash,
            r,
            s,
            v,
        );

        assert!(result.unwrap());
    }

    #[test]
    fn test_verify_ecdsa_signature() {
        let deps = mock_dependencies();
        let api = deps.api;

        let result = verify_ecdsa_signature(
            &api,
            &hex_to_bytes("0a803F9b1CCe32e2773e0d2e98b37E0775cA5d44")[..20]
                .try_into()
                .unwrap(),
            &hex_to_bytes("3102baf2e5ad5188e24d56f239915bed3a9a7b51754007dcbf3a65f81bae3084"),
            &hex_to_bytes("b9b3c9f80a355bd0cd6f609fff4a4b15fa4e3b4632adabb74c020f5bcd24074116fab526529ac795108d201832cff8c2d2b1c710da6711fe9f7ab288a71497581c"),
        );

        assert!(result.unwrap());
    }

    #[test]
    fn test_get_stork_message_hash() {
        let stork_public_key = hex_to_bytes("0a803F9b1CCe32e2773e0d2e98b37E0775cA5d44")[..20]
            .try_into()
            .unwrap();
        let id = hex_to_bytes("7404e3d104ea7841c3d9e6fd20adfe99b4ad586bc08d8f3bd3afef894cf184de")
            [..32]
            .try_into()
            .unwrap();
        let recv_time = 1722632569208762117;
        let quantized_value = 62507457175499998000000;
        let publisher_merkle_root =
            hex_to_bytes("e5ff773b0316059c04aa157898766731017610dcbeede7d7f169bfeaab7cc318")[..32]
                .try_into()
                .unwrap();
        let value_compute_alg_hash =
            hex_to_bytes("9be7e9f9ed459417d96112a7467bd0b27575a2c7847195c68f805b70ce1795ba")[..32]
                .try_into()
                .unwrap();

        let message_hash = get_stork_message_hash(
            &stork_public_key,
            id,
            recv_time,
            quantized_value,
            publisher_merkle_root,
            value_compute_alg_hash,
        );

        let expected =
            hex_to_bytes("3102baf2e5ad5188e24d56f239915bed3a9a7b51754007dcbf3a65f81bae3084");
        assert_eq!(message_hash.to_vec(), expected);
    }

    #[test]
    fn test_get_recoverable_message_hash() {
        let message =
            hex_to_bytes("3102baf2e5ad5188e24d56f239915bed3a9a7b51754007dcbf3a65f81bae3084");
        let message_hash = get_recoverable_message_hash(&message);

        let expected =
            hex_to_bytes("bfaa04ab8f3947f4687a0cb441f673ac3c2233ec3170e37986ff07e09aa50272");
        assert_eq!(message_hash.to_vec(), expected);
    }

    #[test]
    fn test_get_rsv_signature_from_parts() {
        let r = hex_to_bytes("b9b3c9f80a355bd0cd6f609fff4a4b15fa4e3b4632adabb74c020f5bcd240741");
        let s = hex_to_bytes("16fab526529ac795108d201832cff8c2d2b1c710da6711fe9f7ab288a7149758");
        let v = 28;

        let signature = get_rsv_signature_from_parts(&r, &s, v);

        let mut expected = Vec::new();
        expected.extend_from_slice(&r);
        expected.extend_from_slice(&s);
        expected.push(v);
        assert_eq!(signature.to_vec(), expected);
    }

    #[test]
    fn test_verify_stork_evm_signature_invalid_signature() {
        let deps = mock_dependencies();
        let api = deps.api;

        // Using zeroed inputs which should cause signature recovery to fail
        let stork_public_key = [0u8; 20];
        let id = [0u8; 32];
        let recv_time = 0;
        let quantized_value = 0;
        let publisher_merkle_root = [0u8; 32];
        let value_compute_alg_hash = [0u8; 32];
        let r = [0u8; 32];
        let s = [0u8; 32];
        let v = 27;

        let result = verify_stork_evm_signature(
            &api,
            &stork_public_key,
            id,
            recv_time,
            quantized_value,
            publisher_merkle_root,
            value_compute_alg_hash,
            r,
            s,
            v,
        );

        assert!(!result.unwrap());
    }
}
