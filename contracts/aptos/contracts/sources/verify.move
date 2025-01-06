module stork::verify {

    // === Imports ===

    use stork::evm_pubkey::{self, EvmPubKey};
    use stork::i128::{Self, I128};
    use aptos_std::secp256k1::{Self, ECDSASignature, ECDSARawPubkey};
    use aptos_std::aptos_hash::keccak256;

    // === Public Functions ===

    /// Verifies an EVM signature of a stork signed update
    public fun verify_evm_signature(
        /// The EVM public key
        stork_evm_public_key: &EvmPubKey,
        /// The asset id
        asset_id: vector<u8>,
        /// The timestamp
        recv_time: u64,
        /// The quantized value
        quantized_value: I128,
        /// The publisher's merkle root
        publisher_merkle_root: vector<u8>,
        /// The value compute algorithm hash
        value_compute_alg_hash: vector<u8>,
        /// The signature r
        r: vector<u8>,
        /// The signature s
        s: vector<u8>,
        /// The signature v
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

        let signature = get_rsv_signature_from_parts(&r, &s, v);

        verify_ecdsa_signature(stork_evm_public_key, message, signature)
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
        vector::append(&mut data, evm_pubkey::get_bytes(&stork_evm_public_key));
        vector::append(&mut data, asset_id);

        // left pad with 24 0 bytes
        vector::append(&mut data, vector[0u8, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0]);

        let recv_time_bytes = vector::empty();
        let i = 8;
        while (i > 0) {
            i = i - 1;
            vector::push_back(&mut recv_time_bytes, (recv_time >> (i * 8)) & 0xFF);
        };
        vector::append(&mut data, recv_time_bytes);

        //left pad with 16 0 bytes
        vector::append(&mut data, vector[0u8, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0]);

        let value_bytes = i128::to_bytes(quantized_value);
        vector::append(&mut data, value_bytes);
        vector::append(&mut data, publisher_merkle_root);
        vector::append(&mut data, value_compute_alg_hash);
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
        vector::append(&mut prefix, b"Ethereum Signed Message:\n32");
        let data = vector::empty<u8>();
        vector::append(&mut data, prefix);
        vector::append(&mut data, message);
        data
    }

    fun get_rsv_signature_from_parts(
        r: &vector<u8>,
        s: &vector<u8>,
        v: u8,
    ): ECDSASignature {
        let signature_bytes = vector::append(&mut r, s);
        signature_bytes = vector::push_back(&mut signature_bytes, v);
        secp256k1::ecdsa_signature_from_bytes(signature_bytes)
    }

    fun verify_ecdsa_signature(
        pubkey: &EvmPubKey,
        message: vector<u8>,
        signature: ECDSASignature,
    ): bool {
        let signature_bytes = secp256k1::ecdsa_signature_to_bytes(signature);
        let v = signature_bytes[64];
        let recovery_id: u8 = v - 27;
        let recovered_pubkey_option = secp256k1::ecdsa_recover(message, recovery_id, &signature);

        if (recovered_pubkey_option == option::none()) {
            return false;
        };

        let recovered_pubkey = option::extract(recovered_pubkey_option);
        let evm_pubkey = get_evm_pubkey(recovered_pubkey);
        evm_pubkey == *pubkey
    }

    fun get_evm_pubkey(pubkey: ECDSARawPubkey): EvmPubKey {
        let hashed = keccak256(secp256k1::ecdsa_raw_pubkey_to_bytes(pubkey));
        let evm_address = vector::empty<u8>();
        let i = 12;
        while (i < 32) {
            vector::push_back(&mut evm_address, &hashed[i]);
            i = i + 1;
        };
        evm_pubkey::from_bytes(&evm_address)
    }
}
