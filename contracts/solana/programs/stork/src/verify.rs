use anchor_lang::solana_program::keccak::{hash, Hash};
use anchor_lang::solana_program::{
    secp256k1_recover::secp256k1_recover, secp256k1_recover::Secp256k1Pubkey,
};

pub type EvmPubkey = [u8; 20];

pub fn verify_stork_evm_signature(
    stork_public_key: &EvmPubkey,
    id: [u8; 32],
    recv_time: u64,
    quantized_value: i128,
    publisher_merkle_root: [u8; 32],
    value_compute_alg_hash: [u8; 32],
    r: [u8; 32],
    s: [u8; 32],
    v: u8,
) -> bool {
    let message = get_stork_message_hash(
        stork_public_key,
        id,
        recv_time,
        quantized_value,
        publisher_merkle_root,
        value_compute_alg_hash,
    );
    let signature = get_rsv_signature_from_parts(&r, &s, v);
    return verify_ecdsa_signature(stork_public_key, &message.as_ref(), &signature);
}

fn get_stork_message_hash(
    stork_public_key: &EvmPubkey,
    id: [u8; 32],
    recv_time: u64,
    quantized_value: i128,
    publisher_merkle_root: [u8; 32],
    value_compute_alg_hash: [u8; 32],
) -> Hash {
    let mut data: Vec<u8> = Vec::new();
    data.extend_from_slice(stork_public_key);
    data.extend_from_slice(&id);
    data.extend_from_slice(&[0u8; 24]); // Left pad with 24 zero bytes
    data.extend_from_slice(&recv_time.to_be_bytes());
    data.extend_from_slice(&[0u8; 16]); // Left pad with 16 zero bytes
    data.extend_from_slice(&quantized_value.to_be_bytes());
    data.extend_from_slice(&publisher_merkle_root);
    data.extend_from_slice(&value_compute_alg_hash);
    return hash(&data);
}

fn verify_ecdsa_signature(pubkey: &EvmPubkey, message: &[u8], signature: &[u8]) -> bool {
    if signature.len() != 65 {
        return false;
    }

    let message_hash = get_recoverable_message_hash(message);

    let recovered_pubkey = match recover_secp256k1_pubkey(&message_hash.as_ref(), signature) {
        Some(pk) => pk,
        None => return false,
    };

    let eth_pubkey = get_eth_pubkey(&recovered_pubkey);

    eth_pubkey == *pubkey
}

fn get_rsv_signature_from_parts(r: &[u8], s: &[u8], v: u8) -> [u8; 65] {
    let mut signature = [0; 65];
    signature[..32].copy_from_slice(r);
    signature[32..64].copy_from_slice(s);
    signature[64] = v;
    signature
}

fn get_recoverable_message_hash(message: &[u8]) -> Hash {
    let eip_191_prefix = format!("\x19Ethereum Signed Message:\n{}", message.len());

    let mut data: Vec<u8> = Vec::new();
    data.extend_from_slice(eip_191_prefix.as_bytes());
    data.extend_from_slice(message);

    hash(&data)
}

fn recover_secp256k1_pubkey(message: &[u8], signature: &[u8]) -> Option<Secp256k1Pubkey> {
    let rs = &signature[..64];
    let v = signature[64];
    let recovery_id = match v {
        27 => 0,
        28 => 1,
        _ => return None,
    };

    secp256k1_recover(message, recovery_id, rs).ok()
}

fn get_eth_pubkey(pubkey: &Secp256k1Pubkey) -> EvmPubkey {
    // The Ethereum public key is the last 20 bytes of keccak hashed Secp256k1Pubkey
    let hashed = hash(&pubkey.to_bytes());
    let mut eth_pubkey = [0; 20];
    eth_pubkey.copy_from_slice(&hashed.as_ref()[12..]);
    eth_pubkey
}

#[cfg(test)]
mod tests {
    use super::*;

    fn hex_to_string(hex: &[u8]) -> String {
        hex.iter().map(|b| format!("{:02x}", b)).collect::<String>()
    }

    fn hex_to_bytes(hex: &str) -> Vec<u8> {
        let hex = hex.trim_start_matches("0x");
        (0..hex.len())
            .step_by(2)
            .map(|i| u8::from_str_radix(&hex[i..i + 2], 16).unwrap())
            .collect()
    }

    #[test]
    fn test_verify_stork_evm_signature() {
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

        assert!(verify_stork_evm_signature(
            &stork_public_key,
            id,
            recv_time,
            quantized_value,
            publisher_merkle_root,
            value_compute_alg_hash,
            r,
            s,
            v,
        ));
    }

    #[test]
    fn test_verify_ecdsa_signature() {
        assert!(
            verify_ecdsa_signature(
                hex_to_bytes("0a803F9b1CCe32e2773e0d2e98b37E0775cA5d44")[..20]
                    .try_into()
                    .unwrap(),
                &hex_to_bytes("3102baf2e5ad5188e24d56f239915bed3a9a7b51754007dcbf3a65f81bae3084"),
                &hex_to_bytes("b9b3c9f80a355bd0cd6f609fff4a4b15fa4e3b4632adabb74c020f5bcd24074116fab526529ac795108d201832cff8c2d2b1c710da6711fe9f7ab288a71497581c"),
            )
        );
    }

    #[test]
    fn test_get_rsv_signature_from_parts() {
        let signature = get_rsv_signature_from_parts(
            &hex_to_bytes("b9b3c9f80a355bd0cd6f609fff4a4b15fa4e3b4632adabb74c020f5bcd240741"),
            &hex_to_bytes("16fab526529ac795108d201832cff8c2d2b1c710da6711fe9f7ab288a7149758"),
            hex_to_bytes("1c")[0],
        );
        assert_eq!(hex_to_string(&signature), "b9b3c9f80a355bd0cd6f609fff4a4b15fa4e3b4632adabb74c020f5bcd24074116fab526529ac795108d201832cff8c2d2b1c710da6711fe9f7ab288a71497581c");
    }

    #[test]
    fn test_get_stork_message_hash() {
        let stork_public_key = hex_to_bytes("0x0a803F9b1CCe32e2773e0d2e98b37E0775cA5d44")[..20]
            .try_into()
            .unwrap();

        let id = hex_to_bytes("0x7404e3d104ea7841c3d9e6fd20adfe99b4ad586bc08d8f3bd3afef894cf184de")
            [..32]
            .try_into()
            .unwrap();
        let recv_time = 1722632569208762117;
        let quantized_value = 62507457175499998000000;
        let publisher_merkle_root =
            hex_to_bytes("0xe5ff773b0316059c04aa157898766731017610dcbeede7d7f169bfeaab7cc318")
                [..32]
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

        let message_hash_hex = hex_to_string(&message_hash.as_ref());

        assert_eq!(
            message_hash_hex,
            "3102baf2e5ad5188e24d56f239915bed3a9a7b51754007dcbf3a65f81bae3084"
        );
    }

    #[test]
    fn test_get_recoverable_message_hash() {
        let message =
            hex_to_bytes("3102baf2e5ad5188e24d56f239915bed3a9a7b51754007dcbf3a65f81bae3084");
        let message_hash = get_recoverable_message_hash(&message);
        assert_eq!(
            hex_to_string(&message_hash.as_ref()),
            "bfaa04ab8f3947f4687a0cb441f673ac3c2233ec3170e37986ff07e09aa50272"
        );
    }

    #[test]
    fn test_recover_secp256k1_pubkey() {
        let message =
            hex_to_bytes("bfaa04ab8f3947f4687a0cb441f673ac3c2233ec3170e37986ff07e09aa50272");

        let signature = hex_to_bytes("b9b3c9f80a355bd0cd6f609fff4a4b15fa4e3b4632adabb74c020f5bcd24074116fab526529ac795108d201832cff8c2d2b1c710da6711fe9f7ab288a71497581c");
        let result = recover_secp256k1_pubkey(&message, &signature);
        assert!(result.is_some());

        let pubkey = result.unwrap();
        let eth_pubkey = get_eth_pubkey(&pubkey);

        let eth_pubkey_hex = hex_to_string(&eth_pubkey);

        assert_eq!(eth_pubkey_hex, "0a803f9b1cce32e2773e0d2e98b37e0775ca5d44");
    }
}
