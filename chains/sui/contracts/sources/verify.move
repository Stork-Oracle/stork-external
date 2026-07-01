module stork::verify {

    // === Imports ===

    use stork::evm_pubkey::{Self, EvmPubkey};
    use stork::i128::{Self, I128};
    use sui::hash;
    use sui::ecdsa_k1;

    // === Constants ===

    const EInvalidLength: u64 = 0;

    // Length of bytes32 fields (publisher_merkle_root, value_compute_alg_hash)
    // that the signature binds to. Enforced so field boundaries match the
    // EVM-side payload and signatures cannot be reinterpreted across them.
    const BYTES32_LENGTH: u64 = 32;

    // secp256k1n / 2, big-endian. The upper bound for a canonical (low-s) ECDSA
    // signature; any s above this is the malleable counterpart and is rejected.
    // Matches the constant used by OpenZeppelin's ECDSA library on the EVM side.
    const SECP256K1_HALF_ORDER: vector<u8> = x"7fffffffffffffffffffffffffffffff5d576e7357a4501ddfe92f46681b20a0";

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
        assert!(publisher_merkle_root.length() == BYTES32_LENGTH, EInvalidLength);
        assert!(value_compute_alg_hash.length() == BYTES32_LENGTH, EInvalidLength);

        // Reject non-canonical (high-s) signatures to prevent signature malleability.
        // This mirrors the EVM contract, whose OpenZeppelin ECDSA.recover rejects any
        // s > secp256k1n/2. Sui's ecdsa_k1::secp256k1_ecrecover does not, so we enforce
        // it here. Returns false (not abort) to stay consistent with the other
        // signature-validity failures handled by verify_ecdsa_signature.
        if (!is_low_s(&s)) {
            return false
        };

        let message = get_stork_message_hash(
            stork_evm_public_key,
            id,
            recv_time,
            quantized_value,
            publisher_merkle_root,
            value_compute_alg_hash,
        );
        
        let signature = get_rsv_signature_from_parts(&r, &s, v);
        
        verify_ecdsa_signature(stork_evm_public_key, &message, &signature)
    }

    // === Private Functions ===

    // Returns true if s is a canonical (low) value, i.e. s <= secp256k1n/2.
    // s is a 32-byte big-endian scalar; comparison is a byte-wise lexicographic
    // compare against SECP256K1_HALF_ORDER, which is equivalent to numeric compare
    // for equal-length big-endian values. A non-32-byte s is treated as non-low.
    fun is_low_s(s: &vector<u8>): bool {
        if (s.length() != BYTES32_LENGTH) {
            return false
        };
        let half_order = SECP256K1_HALF_ORDER;
        let mut i = 0;
        while (i < BYTES32_LENGTH) {
            let s_byte = s[i];
            let half_byte = half_order[i];
            if (s_byte < half_byte) {
                return true
            } else if (s_byte > half_byte) {
                return false
            };
            i = i + 1;
        };
        // s == secp256k1n/2 exactly, which is still canonical.
        true
    }

    fun get_stork_message(
        stork_evm_public_key: &EvmPubkey,
        id: vector<u8>,
        recv_time: u64,
        quantized_value: I128,
        publisher_merkle_root: vector<u8>,
        value_compute_alg_hash: vector<u8>,
    ): vector<u8> {
        let mut data = vector<u8>[];
        vector::append(&mut data, evm_pubkey::get_bytes(stork_evm_public_key));
        vector::append(&mut data, id);
        
        // Left pad with 24 zero bytes
        vector::append(&mut data, vector[0u8, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0]);
        
        
        let mut recv_time_bytes = vector<u8>[];
        let mut i = 8;
        while (i > 0) {
            i = i - 1;
            vector::push_back(&mut recv_time_bytes, ((recv_time >> (i * 8)) & 0xFF as u8));
        };
        vector::append(&mut data, recv_time_bytes);
        
        // Left pad with 16 zero bytes
        let sign_extension_byte = if (i128::is_negative(&quantized_value)) { 0xFF } else { 0x00 };
        vector::append(&mut data, vector[sign_extension_byte, sign_extension_byte, sign_extension_byte, sign_extension_byte, sign_extension_byte, sign_extension_byte, sign_extension_byte, sign_extension_byte, sign_extension_byte, sign_extension_byte, sign_extension_byte, sign_extension_byte, sign_extension_byte, sign_extension_byte, sign_extension_byte, sign_extension_byte]);
        
        let value_bytes = i128::to_bytes(quantized_value);
        vector::append(&mut data, value_bytes);
        vector::append(&mut data, publisher_merkle_root);
        vector::append(&mut data, value_compute_alg_hash);
        data
    }

    fun get_stork_message_hash(
        stork_evm_public_key: &EvmPubkey,
        id: vector<u8>,
        recv_time: u64,
        quantized_value: I128,
        publisher_merkle_root: vector<u8>,
        value_compute_alg_hash: vector<u8>,
    ): vector<u8> {
        hash::keccak256(&get_stork_message(stork_evm_public_key, id, recv_time, quantized_value, publisher_merkle_root, value_compute_alg_hash))
    }

    fun get_recoverable_message(message: &vector<u8>): vector<u8> {
        // Create the prefix "\x19Ethereum Signed Message:\n32"
        let mut prefix = vector[
            0x19, // The byte 0x19, not the characters '\x19'
        ];
        vector::append(&mut prefix, b"Ethereum Signed Message:\n32");
        
        let mut data = vector<u8>[];
        vector::append(&mut data, prefix);
        vector::append(&mut data, *message);
        data
    }

    fun get_rsv_signature_from_parts(r: &vector<u8>, s: &vector<u8>, v: u8): vector<u8> {
        let mut signature = vector<u8>[];
        vector::append(&mut signature, *r);
        vector::append(&mut signature, *s);
        vector::push_back(&mut signature, v);
        signature
    }

    fun verify_ecdsa_signature(pubkey: &EvmPubkey, message: &vector<u8>, signature: &vector<u8>): bool {
        if (signature.length() != 65) {
            return false
        };
        
        let recoverable_message = get_recoverable_message(message);

        let mut recovered_pubkey_option = recover_secp256k1_pubkey(&recoverable_message, signature);
        if (recovered_pubkey_option == option::none()) {
            return false
        };
        
        let recovered_pubkey = recovered_pubkey_option.extract();

        let evm_pubkey = get_eth_pubkey(&recovered_pubkey);
        evm_pubkey == *pubkey
    }

    fun recover_secp256k1_pubkey(message: &vector<u8>, signature: &vector<u8>): Option<vector<u8>> {
        let v = signature[64];
        let recovery_id = match (v) {
            27 => 0,
            28 => 1,
            _ => return option::none(),
        };
        let mut signature_copy = *signature;
        signature_copy.pop_back();
        signature_copy.push_back(recovery_id);

        let recovered_pubkey = ecdsa_k1::secp256k1_ecrecover(&signature_copy, message, 0);
        let mut decompressed_pubkey = ecdsa_k1::decompress_pubkey(&recovered_pubkey);
        decompressed_pubkey.remove(0);
        option::some(decompressed_pubkey)
    }

    fun get_eth_pubkey(pubkey: &vector<u8>): EvmPubkey {
        let hashed = hash::keccak256(pubkey);
        let mut eth_address = vector<u8>[];
        let mut i = 12;
        while (i < 32) {
            vector::push_back(&mut eth_address, *vector::borrow(&hashed, i));
            i = i + 1;
        };
        evm_pubkey::from_bytes(eth_address)
    }

    // === Tests ===

    #[test]
    fun test_verify_stork_evm_signature() {
        let stork_public_key = evm_pubkey::from_bytes(x"0a803F9b1CCe32e2773e0d2e98b37E0775cA5d44");
        let id = x"7404e3d104ea7841c3d9e6fd20adfe99b4ad586bc08d8f3bd3afef894cf184de";
        let recv_time = 1722632569208762117;
        let quantized_value = i128::from_u128(62507457175499998000000);
        let publisher_merkle_root = x"e5ff773b0316059c04aa157898766731017610dcbeede7d7f169bfeaab7cc318";
        let value_compute_alg_hash = x"9be7e9f9ed459417d96112a7467bd0b27575a2c7847195c68f805b70ce1795ba";
        let r = x"b9b3c9f80a355bd0cd6f609fff4a4b15fa4e3b4632adabb74c020f5bcd240741";
        let s = x"16fab526529ac795108d201832cff8c2d2b1c710da6711fe9f7ab288a7149758";
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
    fun test_verify_stork_evm_signature_negative_value() {
        let stork_public_key = evm_pubkey::from_bytes(x"3db9E960ECfCcb11969509FAB000c0c96DC51830");
        let id = x"281a649a11eb25eca04f0025c15e99264a056229e722735c7d6c55fef649dfbf";
        let recv_time = 1750794968021348308;
        let quantized_value = i128::new(3020199000000, true);
        let publisher_merkle_root = x"5ea4136e8064520a3311961f3f7030dfbc0b96652f46a473e79f2a019b3cd878";
        let value_compute_alg_hash = x"9be7e9f9ed459417d96112a7467bd0b27575a2c7847195c68f805b70ce1795ba";
        let r = x"14c36cf7272689cec0335efdc5f82dc2d4b1aceb8d2320d3245e4593df32e696";
        let s = x"79ab437ecd56dc9fcf850f192328840f7f47d5df57cb939d99146b33014c39f0";
        let v = 27;

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
    fun test_get_recoverable_message() {
        let message = x"3102baf2e5ad5188e24d56f239915bed3a9a7b51754007dcbf3a65f81bae3084";
        let recoverable_message = get_recoverable_message(&message);
        assert!(recoverable_message == x"19457468657265756d205369676e6564204d6573736167653a0a33323102baf2e5ad5188e24d56f239915bed3a9a7b51754007dcbf3a65f81bae3084");
    }
    
    #[test]
    #[expected_failure(abort_code = EInvalidLength)]
    fun test_verify_stork_evm_signature_rejects_short_publisher_merkle_root() {
        let stork_public_key = evm_pubkey::from_bytes(x"0a803F9b1CCe32e2773e0d2e98b37E0775cA5d44");
        let id = x"7404e3d104ea7841c3d9e6fd20adfe99b4ad586bc08d8f3bd3afef894cf184de";
        let recv_time = 1722632569208762117;
        let quantized_value = i128::from_u128(62507457175499998000000);
        // 31 bytes instead of 32 — should abort
        let publisher_merkle_root = x"e5ff773b0316059c04aa157898766731017610dcbeede7d7f169bfeaab7cc3";
        let value_compute_alg_hash = x"9be7e9f9ed459417d96112a7467bd0b27575a2c7847195c68f805b70ce1795ba";
        let r = x"b9b3c9f80a355bd0cd6f609fff4a4b15fa4e3b4632adabb74c020f5bcd240741";
        let s = x"16fab526529ac795108d201832cff8c2d2b1c710da6711fe9f7ab288a7149758";
        let v = 28;

        verify_stork_evm_signature(
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
    }

    #[test]
    #[expected_failure(abort_code = EInvalidLength)]
    fun test_verify_stork_evm_signature_rejects_short_value_compute_alg_hash() {
        let stork_public_key = evm_pubkey::from_bytes(x"0a803F9b1CCe32e2773e0d2e98b37E0775cA5d44");
        let id = x"7404e3d104ea7841c3d9e6fd20adfe99b4ad586bc08d8f3bd3afef894cf184de";
        let recv_time = 1722632569208762117;
        let quantized_value = i128::from_u128(62507457175499998000000);
        let publisher_merkle_root = x"e5ff773b0316059c04aa157898766731017610dcbeede7d7f169bfeaab7cc318";
        // 31 bytes instead of 32 — should abort
        let value_compute_alg_hash = x"9be7e9f9ed459417d96112a7467bd0b27575a2c7847195c68f805b70ce17";
        let r = x"b9b3c9f80a355bd0cd6f609fff4a4b15fa4e3b4632adabb74c020f5bcd240741";
        let s = x"16fab526529ac795108d201832cff8c2d2b1c710da6711fe9f7ab288a7149758";
        let v = 28;

        verify_stork_evm_signature(
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
    }

    // A signature whose byte length is not 65 (here r is 31 bytes, so r||s||v == 64)
    // must be rejected — verify returns false rather than aborting.
    #[test]
    fun test_verify_stork_evm_signature_rejects_wrong_length_signature() {
        let stork_public_key = evm_pubkey::from_bytes(x"0a803F9b1CCe32e2773e0d2e98b37E0775cA5d44");
        let id = x"7404e3d104ea7841c3d9e6fd20adfe99b4ad586bc08d8f3bd3afef894cf184de";
        let recv_time = 1722632569208762117;
        let quantized_value = i128::from_u128(62507457175499998000000);
        let publisher_merkle_root = x"e5ff773b0316059c04aa157898766731017610dcbeede7d7f169bfeaab7cc318";
        let value_compute_alg_hash = x"9be7e9f9ed459417d96112a7467bd0b27575a2c7847195c68f805b70ce1795ba";
        // 31-byte r → total signature length 64, not 65
        let r = x"b9b3c9f80a355bd0cd6f609fff4a4b15fa4e3b4632adabb74c020f5bcd2407";
        let s = x"16fab526529ac795108d201832cff8c2d2b1c710da6711fe9f7ab288a7149758";
        let v = 28;

        assert!(!verify_stork_evm_signature(
            &stork_public_key, id, recv_time, quantized_value,
            publisher_merkle_root, value_compute_alg_hash, r, s, v,
        ));
    }

    // A recovery id (v) other than 27 or 28 must be rejected.
    #[test]
    fun test_verify_stork_evm_signature_rejects_invalid_v() {
        let stork_public_key = evm_pubkey::from_bytes(x"0a803F9b1CCe32e2773e0d2e98b37E0775cA5d44");
        let id = x"7404e3d104ea7841c3d9e6fd20adfe99b4ad586bc08d8f3bd3afef894cf184de";
        let recv_time = 1722632569208762117;
        let quantized_value = i128::from_u128(62507457175499998000000);
        let publisher_merkle_root = x"e5ff773b0316059c04aa157898766731017610dcbeede7d7f169bfeaab7cc318";
        let value_compute_alg_hash = x"9be7e9f9ed459417d96112a7467bd0b27575a2c7847195c68f805b70ce1795ba";
        let r = x"b9b3c9f80a355bd0cd6f609fff4a4b15fa4e3b4632adabb74c020f5bcd240741";
        let s = x"16fab526529ac795108d201832cff8c2d2b1c710da6711fe9f7ab288a7149758";
        // valid signature bytes but an out-of-range recovery id
        let v = 29;

        assert!(!verify_stork_evm_signature(
            &stork_public_key, id, recv_time, quantized_value,
            publisher_merkle_root, value_compute_alg_hash, r, s, v,
        ));
    }

    // A valid signature checked against a different public key must not verify.
    #[test]
    fun test_verify_stork_evm_signature_rejects_wrong_key() {
        // Not the key that produced the signature below.
        let wrong_public_key = evm_pubkey::from_bytes(x"3db9E960ECfCcb11969509FAB000c0c96DC51830");
        let id = x"7404e3d104ea7841c3d9e6fd20adfe99b4ad586bc08d8f3bd3afef894cf184de";
        let recv_time = 1722632569208762117;
        let quantized_value = i128::from_u128(62507457175499998000000);
        let publisher_merkle_root = x"e5ff773b0316059c04aa157898766731017610dcbeede7d7f169bfeaab7cc318";
        let value_compute_alg_hash = x"9be7e9f9ed459417d96112a7467bd0b27575a2c7847195c68f805b70ce1795ba";
        let r = x"b9b3c9f80a355bd0cd6f609fff4a4b15fa4e3b4632adabb74c020f5bcd240741";
        let s = x"16fab526529ac795108d201832cff8c2d2b1c710da6711fe9f7ab288a7149758";
        let v = 28;

        assert!(!verify_stork_evm_signature(
            &wrong_public_key, id, recv_time, quantized_value,
            publisher_merkle_root, value_compute_alg_hash, r, s, v,
        ));
    }

    // Signature malleability: the high-s counterpart of a valid low-s signature
    // (s' = n - s, with the recovery id flipped) recovers the same public key and
    // would otherwise be an accepted second form of an existing signature. The
    // low-s enforcement rejects it, matching the EVM contract's OpenZeppelin
    // ECDSA.recover, which reverts on any s > secp256k1n/2.
    #[test]
    fun test_verify_stork_evm_signature_rejects_high_s_malleable_signature() {
        let stork_public_key = evm_pubkey::from_bytes(x"0a803F9b1CCe32e2773e0d2e98b37E0775cA5d44");
        let id = x"7404e3d104ea7841c3d9e6fd20adfe99b4ad586bc08d8f3bd3afef894cf184de";
        let recv_time = 1722632569208762117;
        let quantized_value = i128::from_u128(62507457175499998000000);
        let publisher_merkle_root = x"e5ff773b0316059c04aa157898766731017610dcbeede7d7f169bfeaab7cc318";
        let value_compute_alg_hash = x"9be7e9f9ed459417d96112a7467bd0b27575a2c7847195c68f805b70ce1795ba";
        let r = x"b9b3c9f80a355bd0cd6f609fff4a4b15fa4e3b4632adabb74c020f5bcd240741";
        // s' = n - VALID_S, recovery id flipped from 28 to 27 — the malleable form.
        let high_s = x"e9054ad9ad65386aef72dfe7cd30073be7fd15d5d4e18e3d2057ac042921a9e9";
        let v = 27;

        assert!(!verify_stork_evm_signature(
            &stork_public_key, id, recv_time, quantized_value,
            publisher_merkle_root, value_compute_alg_hash, r, high_s, v,
        ));
    }

    // A signature at exactly secp256k1n/2 is still canonical (low-s) and must be
    // accepted by is_low_s — the boundary is inclusive, matching OZ (rejects s > n/2).
    #[test]
    fun test_is_low_s_boundary() {
        // s == secp256k1n/2 → low-s
        assert!(is_low_s(&x"7fffffffffffffffffffffffffffffff5d576e7357a4501ddfe92f46681b20a0"));
        // s == secp256k1n/2 + 1 → high-s
        assert!(!is_low_s(&x"7fffffffffffffffffffffffffffffff5d576e7357a4501ddfe92f46681b20a1"));
        // a small canonical value → low-s
        assert!(is_low_s(&x"0000000000000000000000000000000000000000000000000000000000000001"));
        // wrong length → treated as non-low
        assert!(!is_low_s(&x"01"));
    }

    #[test]
    fun test_recover_secp256k1_pubkey() {
        let message = x"3102baf2e5ad5188e24d56f239915bed3a9a7b51754007dcbf3a65f81bae3084";
        let recoverable_message = get_recoverable_message(&message);
        let signature = x"b9b3c9f80a355bd0cd6f609fff4a4b15fa4e3b4632adabb74c020f5bcd24074116fab526529ac795108d201832cff8c2d2b1c710da6711fe9f7ab288a71497581c";
        
        let mut recovered_pubkey_option = recover_secp256k1_pubkey(&recoverable_message, &signature);
        assert!(option::is_some(&recovered_pubkey_option));
        
        let recovered_pubkey = recovered_pubkey_option.extract();
        let eth_pubkey = get_eth_pubkey(&recovered_pubkey);
        
        assert!(eth_pubkey == evm_pubkey::from_bytes(x"0a803F9b1CCe32e2773e0d2e98b37E0775cA5d44"));
    }
}
