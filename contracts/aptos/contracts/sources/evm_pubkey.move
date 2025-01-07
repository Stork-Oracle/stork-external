module stork::evm_pubkey {


    // === Imports ===

    use std::vector;

    // === Errors ===

    /// The EVM public key length is invalid
    const E_INVALID_LENGTH: u64 = 0;

    // === Constants ===

    /// The length of an EVM public key
    const EVM_PUBKEY_LENGTH: u64 = 20;

    // === Structs ===

    /// The EVM public key struct 
    struct EvmPubKey has copy, drop, store {
        /// Byte array of the EVM public key
        bytes: vector<u8>,  
    }

    // === Functions ===

    /// Creates a new EVM public key
    public fun from_bytes(bytes: vector<u8>): EvmPubKey {
        assert!(bytes.length() == EVM_PUBKEY_LENGTH, E_INVALID_LENGTH);
        EvmPubKey { bytes }
    }

    /// Gets the bytes of the EVM public key
    public fun get_bytes(self: &EvmPubKey): vector<u8> {
        self.bytes
    }

    // === Tests ===

    #[test]
    fun test_from_bytes() {
        let bytes = create_zeroed_byte_vector(EVM_PUBKEY_LENGTH); 
        let evm_pubkey = from_bytes(bytes);
        assert!(evm_pubkey.bytes == bytes);
    }

    #[test]
    #[expected_failure(abort_code = E_INVALID_LENGTH)]
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
        let i = 0;
        let bytes = vector::empty<u8>();
        while (i < length) {
            bytes.push_back(zero);
            i = i + 1;
        };
        bytes
    }

    #[test_only]
    package fun create_zeroed_evm_pubkey(): EvmPubKey {
        let bytes = create_zeroed_byte_vector(EVM_PUBKEY_LENGTH);
        from_bytes(bytes)
    }
}