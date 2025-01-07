module stork::verify {

    // === Imports ===

    use stork::evm_pubkey::{Self, EvmPubKey};
    use stork::i128::{Self, I128};
    use aptos_std::secp256k1::{Self, ECDSASignature, ECDSARawPublicKey};
    use aptos_std::aptos_hash::keccak256;
    use std::vector;
    use std::option;
    
    // === Public Functions ===

    /// Verifies an EVM signature of a stork signed update
    public fun verify_evm_signature(
        // The EVM public key
        stork_evm_public_key: &EvmPubKey,
        // The asset id
        asset_id: vector<u8>,
        // The timestamp
        recv_time: u64,
        // The quantized value
        quantized_value: I128,
        // The publisher's merkle root
        publisher_merkle_root: vector<u8>,
        // The value compute algorithm hash
        value_compute_alg_hash: vector<u8>,
        // The signature r
        r: vector<u8>,
        // The signature s
        s: vector<u8>,
        // The signature v
        v: u8,
    ): bool {
        let message = get_stork_message_hash(
            stork_evm_public_key,
            asset_id,
            recv_time,
            quantized_value,
            publisher_merkle_root,
            value_compute_alg_hash,
        );

        let signature = get_rs_signature_from_parts(r, s);
        let recovery_id = get_recovery_id(v);
        verify_ecdsa_signature(stork_evm_public_key, message, signature, recovery_id)
    }


    // === Private Functions ===

    fun get_stork_message_bytes(
        stork_evm_public_key: &EvmPubKey,
        asset_id: vector<u8>,
        recv_time: u64,
        quantized_value: I128,
        publisher_merkle_root: vector<u8>,
        value_compute_alg_hash: vector<u8>,
    ): vector<u8> {
        let data = vector::empty();
        data.append(evm_pubkey::get_bytes(stork_evm_public_key));
        data.append(asset_id);

        // left pad with 24 0 bytes
        data.append(vector[0u8, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0]);

        let recv_time_bytes = vector::empty<u8>();
        let i: u8 = 8;
        while (i > 0) {
            i = i - 1;
            recv_time_bytes.push_back(((recv_time >> (i * 8)) & 0xFF) as u8);
        };
        data.append(recv_time_bytes);

        //left pad with 16 0 bytes
        data.append(vector[0u8, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0]);

        let value_bytes = i128::to_bytes(quantized_value);
        data.append(value_bytes);
        data.append(publisher_merkle_root);
        data.append(value_compute_alg_hash);
        data
    }

    fun get_stork_message_hash(
        stork_evm_public_key: &EvmPubKey,
        asset_id: vector<u8>,
        recv_time: u64,
        quantized_value: I128,
        publisher_merkle_root: vector<u8>,
        value_compute_alg_hash: vector<u8>,
    ): vector<u8> {
        keccak256(get_stork_message_bytes(stork_evm_public_key, asset_id, recv_time, quantized_value, publisher_merkle_root, value_compute_alg_hash))
    }

    fun get_recoverable_message(message: vector<u8>): vector<u8> {
        // create the prefix "\x19Ethereum Signed Message:\n32"
        let prefix = vector[0x19];
        prefix.append(b"Ethereum Signed Message:\n32");
        let data = vector::empty<u8>();
        data.append(prefix);
        data.append(message);
        data
    }

    fun get_recoverable_message_hash(message: vector<u8>): vector<u8> {
        keccak256(get_recoverable_message(message))
    }

    fun get_rs_signature_from_parts(
        r: vector<u8>,
        s: vector<u8>,
    ): ECDSASignature {
        let signature_bytes = vector::empty();
        signature_bytes.append(r);
        signature_bytes.append(s);
        
        secp256k1::ecdsa_signature_from_bytes(signature_bytes)
    }

    fun verify_ecdsa_signature(
        pubkey: &EvmPubKey,
        message: vector<u8>,
        signature: ECDSASignature,
        recovery_id: u8,
    ): bool {

        let recoverable_message_hash = get_recoverable_message_hash(message);

        let recovered_pubkey_option = secp256k1::ecdsa_recover(recoverable_message_hash, recovery_id, &signature);

        if (recovered_pubkey_option == option::none()) {
            return false;
        };

        let recovered_pubkey = recovered_pubkey_option.extract();
        let evm_pubkey = get_evm_pubkey(recovered_pubkey);
        evm_pubkey == *pubkey
    }

    fun get_evm_pubkey(pubkey: ECDSARawPublicKey): EvmPubKey {
        let hashed = keccak256(secp256k1::ecdsa_raw_public_key_to_bytes(&pubkey));
        let evm_address = vector::empty<u8>();
        let i = 12;
        while (i < 32) {
            evm_address.push_back(hashed[i]);
            i = i + 1;
        };
        evm_pubkey::from_bytes(evm_address)
    }

    fun get_recovery_id(v: u8): u8 {
        v - 27
    }

// === Tests ===

    #[test]
    fun test_verify_evm_signature() {
        let stork_public_key = evm_pubkey::from_bytes(x"0a803F9b1CCe32e2773e0d2e98b37E0775cA5d44");
        let asset_id = x"7404e3d104ea7841c3d9e6fd20adfe99b4ad586bc08d8f3bd3afef894cf184de";
        let recv_time = 1722632569208762117;
        let quantized_value = i128::from_u128(62507457175499998000000);
        let publisher_merkle_root = x"e5ff773b0316059c04aa157898766731017610dcbeede7d7f169bfeaab7cc318";
        let value_compute_alg_hash = x"9be7e9f9ed459417d96112a7467bd0b27575a2c7847195c68f805b70ce1795ba";
        let r = x"b9b3c9f80a355bd0cd6f609fff4a4b15fa4e3b4632adabb74c020f5bcd240741";
        let s = x"16fab526529ac795108d201832cff8c2d2b1c710da6711fe9f7ab288a7149758";
        let v = 28;

        assert!(verify_evm_signature(
            &stork_public_key,
            asset_id,
            recv_time,
            quantized_value,
            publisher_merkle_root,
            value_compute_alg_hash,
            r,
            s,
            v,
        ), 0);
    }

    #[test]
    fun test_get_stork_message_hash() {
        let stork_public_key = evm_pubkey::from_bytes(x"0a803F9b1CCe32e2773e0d2e98b37E0775cA5d44");
        let asset_id = x"7404e3d104ea7841c3d9e6fd20adfe99b4ad586bc08d8f3bd3afef894cf184de";
        let recv_time = 1722632569208762117;
        let quantized_value = i128::from_u128(62507457175499998000000);
        let publisher_merkle_root = x"e5ff773b0316059c04aa157898766731017610dcbeede7d7f169bfeaab7cc318";
        let value_compute_alg_hash = x"9be7e9f9ed459417d96112a7467bd0b27575a2c7847195c68f805b70ce1795ba";

        let message_hash = get_stork_message_hash(
            &stork_public_key,
            asset_id,
            recv_time,
            quantized_value,
            publisher_merkle_root,
            value_compute_alg_hash,
        );

        assert!(message_hash == x"3102baf2e5ad5188e24d56f239915bed3a9a7b51754007dcbf3a65f81bae3084", 0);
    }

    #[test]
    fun test_get_recoverable_message() {
        let message = x"3102baf2e5ad5188e24d56f239915bed3a9a7b51754007dcbf3a65f81bae3084";
        let recoverable_message = get_recoverable_message(message);
        assert!(recoverable_message == x"19457468657265756d205369676e6564204d6573736167653a0a33323102baf2e5ad5188e24d56f239915bed3a9a7b51754007dcbf3a65f81bae3084", 0);
    }

    #[test]
    fun test_get_rsv_signature() {
        let r = x"b9b3c9f80a355bd0cd6f609fff4a4b15fa4e3b4632adabb74c020f5bcd240741";
        let s = x"16fab526529ac795108d201832cff8c2d2b1c710da6711fe9f7ab288a7149758";
        let signature = get_rs_signature_from_parts(r, s);
        let signature_bytes = secp256k1::ecdsa_signature_to_bytes(&signature);

        assert!(signature_bytes == x"b9b3c9f80a355bd0cd6f609fff4a4b15fa4e3b4632adabb74c020f5bcd24074116fab526529ac795108d201832cff8c2d2b1c710da6711fe9f7ab288a7149758", 0);

    }

    #[test]
    fun test_get_recovery_id() {
        let v = 28;
        let recovery_id = get_recovery_id(v);
        assert!(recovery_id == 1, 0);
    }

    #[test]
    fun test_get_evm_pubkey() {
        let evm_pubkey = evm_pubkey::from_bytes(x"E419C8cF64567DE0bd125e74bEAE3041BA2636B9");
        let ecdsa_pubkey = secp256k1::ecdsa_raw_public_key_from_64_bytes(x"b10aec244e3ed4b584084cc61750f4d4c0140765bc72d8600c9208fe86d5c4842deadf18ddae210f57c4d89244c567622e3350421c298bfc94fc65e6195af261");
        assert!(get_evm_pubkey(ecdsa_pubkey) == evm_pubkey, 0);
    }

    #[test]
    fun test_verify_ecdsa_signature() {
        let pubkey = evm_pubkey::from_bytes(x"0a803F9b1CCe32e2773e0d2e98b37E0775cA5d44");
        let message = x"3102baf2e5ad5188e24d56f239915bed3a9a7b51754007dcbf3a65f81bae3084";
        let r = x"b9b3c9f80a355bd0cd6f609fff4a4b15fa4e3b4632adabb74c020f5bcd240741";
        let s = x"16fab526529ac795108d201832cff8c2d2b1c710da6711fe9f7ab288a7149758";
        let v = 28;

        let signature = get_rs_signature_from_parts(r, s);
        let recovery_id = get_recovery_id(v);
        assert!(verify_ecdsa_signature(&pubkey, message, signature, recovery_id), 0);
    }

    #[test]
    #[expected_failure(abort_code = 0, location = stork::verify)]
    fun test_verify_evm_signature_fails_with_wrong_value() {
        let stork_public_key = evm_pubkey::from_bytes(x"0a803F9b1CCe32e2773e0d2e98b37E0775cA5d44");
        let asset_id = x"7404e3d104ea7841c3d9e6fd20adfe99b4ad586bc08d8f3bd3afef894cf184de";
        let recv_time = 1722632569208762117;
        let quantized_value = i128::from_u128(62507457175499998000000 + 1); // Increment value by 1
        let publisher_merkle_root = x"e5ff773b0316059c04aa157898766731017610dcbeede7d7f169bfeaab7cc318";
        let value_compute_alg_hash = x"9be7e9f9ed459417d96112a7467bd0b27575a2c7847195c68f805b70ce1795ba";
        let r = x"b9b3c9f80a355bd0cd6f609fff4a4b15fa4e3b4632adabb74c020f5bcd240741";
        let s = x"16fab526529ac795108d201832cff8c2d2b1c710da6711fe9f7ab288a7149758";
        let v = 28;

        assert!(verify_evm_signature(
            &stork_public_key,
            asset_id,
            recv_time,
            quantized_value,
            publisher_merkle_root,
            value_compute_alg_hash,
            r,
            s,
            v,
        ), 0);
    }

}