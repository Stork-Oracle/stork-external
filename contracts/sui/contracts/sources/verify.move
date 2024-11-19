module stork::verify {

    // === Imports ===

    use stork::evm_pubkey::{Self, EvmPubkey};
    use stork::i128::{Self, I128};
    use sui::hash;
    use sui::ecdsa_k1;
    use sui::bcs;

    // === Public Functions ===

    public fun verify_stork_evm_signature(
        stork_evm_public_key: &EvmPubkey,
        id: vector<u8>,
        recv_time: u64,
        quantized_value: I128,
        publisher_merkle_root: vector<u8>,
        value_compute_alg_hash: vector<u8>,
        r: vector<u8>,
        s: vector<u8>,
        v: u8,
    ): bool {
        let message = get_stork_message_hash(
            stork_evm_public_key,
            id,
            recv_time,
            quantized_value,
            publisher_merkle_root,
            value_compute_alg_hash,
        );
        
        let signature = get_rsv_signature_from_parts(&r, &s, v);
        let recoverable_message = get_recoverable_message_hash(&message);
        
        verify_ecdsa_signature(stork_evm_public_key, &recoverable_message, &signature)
    }

    // === Private Functions ===

    fun get_stork_message_hash(
        stork_evm_public_key: &EvmPubkey,
        id: vector<u8>,
        recv_time: u64,
        quantized_value: I128,
        publisher_merkle_root: vector<u8>,
        value_compute_alg_hash: vector<u8>,
    ): vector<u8> {
        let mut data = vector::empty<u8>();
        vector::append(&mut data, evm_pubkey::get_bytes(stork_evm_public_key));
        vector::append(&mut data, id);
        
        // Left pad with 24 zero bytes
        vector::append(&mut data, vector[0u8, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0]);
        
        let recv_time_bytes = bcs::to_bytes(&recv_time);
        vector::append(&mut data, recv_time_bytes);
        
        // Left pad with 16 zero bytes
        vector::append(&mut data, vector[0u8, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0]);
        
        let value_bytes = i128::to_bytes(quantized_value);
        vector::append(&mut data, value_bytes);
        vector::append(&mut data, publisher_merkle_root);
        vector::append(&mut data, value_compute_alg_hash);
        
        hash::keccak256(&data)
    }

    fun get_recoverable_message_hash(message: &vector<u8>): vector<u8> {
        let prefix = b"\x19Ethereum Signed Message:\n32";
        let mut data = vector::empty<u8>();
        vector::append(&mut data, prefix);
        vector::append(&mut data, *message);
        hash::keccak256(&data)
    }

    fun get_rsv_signature_from_parts(r: &vector<u8>, s: &vector<u8>, v: u8): vector<u8> {
        let mut signature = vector::empty<u8>();
        vector::append(&mut signature, *r);
        vector::append(&mut signature, *s);
        vector::push_back(&mut signature, v);
        signature
    }

    fun verify_ecdsa_signature(pubkey: &EvmPubkey, message: &vector<u8>, signature: &vector<u8>): bool {
        if (signature.length() != 65) {
            return false
        };
        let v = signature[64];
        let recovery_id = match (v) {
            27 => 0,
            28 => 1,
            _ => return false,
        };
        let recovered_pubkey = ecdsa_k1::secp256k1_ecrecover(signature, message, recovery_id);
        
        // Get last 20 bytes of keccak hash of recovered pubkey
        let hashed = hash::keccak256(&recovered_pubkey);
        let mut eth_address = vector::empty<u8>();
        let mut i = 12;
        while (i < 32) {
            vector::push_back(&mut eth_address, *vector::borrow(&hashed, i));
            i = i + 1;
        };
        
        eth_address == evm_pubkey::get_bytes(pubkey)
    }

    // === Tests ===

    #[test]
    fun test_verify_ecdsa_signature() {
        let pubkey = evm_pubkey::from_bytes(x"0a803F9b1CCe32e2773e0d2e98b37E0775cA5d44");
        let message = x"3102baf2e5ad5188e24d56f239915bed3a9a7b51754007dcbf3a65f81bae3084";
        let signature = x"b9b3c9f80a355bd0cd6f609fff4a4b15fa4e3b4632adabb74c020f5bcd24074116fab526529ac795108d201832cff8c2d2b1c710da6711fe9f7ab288a71497581c";
        
        assert!(verify_ecdsa_signature(&pubkey, &message, &signature));
    }

    #[test]
    fun test_get_rsv_signature_from_parts() {
        let r = x"b9b3c9f80a355bd0cd6f609fff4a4b15fa4e3b4632adabb74c020f5bcd240741";
        let s = x"16fab526529ac795108d201832cff8c2d2b1c710da6711fe9f7ab288a7149758";
        let v = 28; // 0x1c
        
        let signature = get_rsv_signature_from_parts(&r, &s, v);
        assert!(signature == x"b9b3c9f80a355bd0cd6f609fff4a4b15fa4e3b4632adabb74c020f5bcd24074116fab526529ac795108d201832cff8c2d2b1c710da6711fe9f7ab288a71497581c");
    }

    #[test]
    fun test_get_stork_message_hash() {
        let stork_public_key = evm_pubkey::from_bytes(x"0a803F9b1CCe32e2773e0d2e98b37E0775cA5d44");
        let id = x"7404e3d104ea7841c3d9e6fd20adfe99b4ad586bc08d8f3bd3afef894cf184de";
        let recv_time = 1722632569208762117;
        let quantized_value = i128::from_u128(62507457175499998000000);
        let publisher_merkle_root = x"e5ff773b0316059c04aa157898766731017610dcbeede7d7f169bfeaab7cc318";
        let value_compute_alg_hash = x"9be7e9f9ed459417d96112a7467bd0b27575a2c7847195c68f805b70ce1795ba";

        let message_hash = get_stork_message_hash(
            &stork_public_key,
            id,
            recv_time,
            quantized_value,
            publisher_merkle_root,
            value_compute_alg_hash,
        );

        assert!(message_hash == x"3102baf2e5ad5188e24d56f239915bed3a9a7b51754007dcbf3a65f81bae3084");
    }

    #[test]
    fun test_get_recoverable_message_hash() {
        let message = x"3102baf2e5ad5188e24d56f239915bed3a9a7b51754007dcbf3a65f81bae3084";
        let message_hash = get_recoverable_message_hash(&message);
        
        assert!(message_hash == x"bfaa04ab8f3947f4687a0cb441f673ac3c2233ec3170e37986ff07e09aa50272");
    }

    #[test]
    fun test_recover_and_verify_pubkey() {
        let message = x"bfaa04ab8f3947f4687a0cb441f673ac3c2233ec3170e37986ff07e09aa50272";
        let signature = x"b9b3c9f80a355bd0cd6f609fff4a4b15fa4e3b4632adabb74c020f5bcd24074116fab526529ac795108d201832cff8c2d2b1c710da6711fe9f7ab288a71497581c";
        
        let recovered_pubkey = ecdsa_k1::secp256k1_ecrecover(&message, &signature);
        let hashed = hash::keccak256(&recovered_pubkey);
        
        let eth_address = vector::empty<u8>();
        let i = 12;
        while (i < 32) {
            vector::push_back(&mut eth_address, *vector::borrow(&hashed, i));
            i = i + 1;
        };
        
        assert!(eth_address == x"0a803f9b1cce32e2773e0d2e98b37e0775ca5d44");
    }
}
