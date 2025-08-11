module stork::encoded_asset_id {

    // === Errors ===

    /// The encoded asset id length is invalid
    const E_INVALID_LENGTH: u64 = 0;

    // === Constants ===

    /// The length of an encoded asset id
    const ASSET_ID_LENGTH: u64 = 32;

    // === Structs ===

    /// The encoded asset id struct
    struct EncodedAssetId has copy, drop, store {
        bytes: vector<u8>,
    }

    // === Functions ===

    /// Creates a new encoded asset id from bytes
    public fun from_bytes(bytes: vector<u8>): EncodedAssetId {
        assert!(bytes.length() == ASSET_ID_LENGTH, E_INVALID_LENGTH);
        EncodedAssetId {
            bytes,
        }
    }

    /// Gets the bytes of an encoded asset id
    public fun get_bytes(self: &EncodedAssetId): vector<u8> {
        self.bytes
    }

    // === Test Imports ===

    #[test_only]
    use std::vector;

    // === Tests ===

    #[test]
    fun test_from_bytes() {
        let bytes = create_zeroed_byte_vector(ASSET_ID_LENGTH);
        let encoded_asset_id = from_bytes(bytes);
        assert!(encoded_asset_id.bytes == bytes);
    }

    #[test]
    #[expected_failure(abort_code = E_INVALID_LENGTH)]
    fun test_from_bytes_invalid_length() {
        let bytes = create_zeroed_byte_vector(ASSET_ID_LENGTH + 1);
        from_bytes(bytes);
    }

    #[test]
    fun test_get_bytes() {
        let bytes = create_zeroed_byte_vector(ASSET_ID_LENGTH);
        let encoded_asset_id = from_bytes(bytes);
        assert!(get_bytes(&encoded_asset_id) == bytes);
    }

    // === Test Helpers ===

    #[test_only]
    fun create_zeroed_byte_vector(length: u64): vector<u8> {
        let zero = 0u8;
        let bytes = vector::empty<u8>();
        let i = 0;
        while (i < length) {
            bytes.push_back(zero);
            i = i + 1;
        };
        bytes
    }

    #[test_only]
    package fun create_zeroed_asset_id(): EncodedAssetId {
        from_bytes(create_zeroed_byte_vector(ASSET_ID_LENGTH))
    }
}
