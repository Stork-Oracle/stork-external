module stork::evm_pubkey {

    // === Errors ===

    const EInvalidLength: u64 = 0;

    // === Constants ===

    const EVM_PUBKEY_LENGTH: u64 = 20;

    // === Structs ===
    public struct EvmPubkey has copy, drop, store {
        bytes: vector<u8>,
    }

    // === Functions ===

    public fun from_bytes(bytes: vector<u8>): EvmPubkey {
        assert!(bytes.length() == EVM_PUBKEY_LENGTH, EInvalidLength);
        EvmPubkey { bytes }
    }

    public fun get_bytes(evm_pubkey: &EvmPubkey): vector<u8> {
        evm_pubkey.bytes
    }

    // === Tests ===

    #[test]
    fun test_from_bytes() {
        let bytes = create_zeroed_byte_vector(EVM_PUBKEY_LENGTH); 
        let evm_pubkey = from_bytes(bytes);
        assert!(evm_pubkey.bytes == bytes);
    }

    #[test]
    #[expected_failure(abort_code = EInvalidLength)]
    fun test_from_bytes_invalid_length() {
        let bytes = create_zeroed_byte_vector(EVM_PUBKEY_LENGTH + 1);
        from_bytes(bytes);
    }

    #[test]
    fun test_get_bytes() {
        let bytes = create_zeroed_byte_vector(EVM_PUBKEY_LENGTH);
        let evm_pubkey = from_bytes(bytes);
        assert!(get_bytes(&evm_pubkey) == bytes);
    }

    #[test_only]
    fun create_zeroed_byte_vector(length: u64): vector<u8> {
        let zero = 0u8;
        let mut i = 0;
        let mut bytes = vector::empty<u8>();
        while (i < length) {
            bytes.push_back(zero);
            i = i + 1;
        };
        bytes
    }
}