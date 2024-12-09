module stork::encoded_asset_id {

    // === Errors ===

    const EInvalidLength: u64 = 0;

    // === Constants ===

    const ASSET_ID_LENGTH: u64 = 32;

    // === Structs ===

    public struct EncodedAssetId has copy, drop, store {
        bytes: vector<u8>,
    }

    // === Functions ===

    public fun from_bytes(bytes: vector<u8>): EncodedAssetId {
        assert!(bytes.length() == ASSET_ID_LENGTH, EInvalidLength);
        EncodedAssetId {
            bytes,
        }
    }

    public fun get_bytes(encoded_asset_id: &EncodedAssetId): vector<u8> {
        encoded_asset_id.bytes
    }

    // === Tests ===

    #[test]
    fun test_from_bytes() {
        let bytes = create_zeroed_byte_vector(ASSET_ID_LENGTH);
        let encoded_asset_id = from_bytes(bytes);
        assert!(encoded_asset_id.bytes == bytes);
    }

    #[test]
    #[expected_failure(abort_code = EInvalidLength)]
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
        let mut i = 0;
        let mut bytes = vector::empty<u8>();
        while (i < length) {
            bytes.push_back(zero);
            i = i + 1;
        };
        bytes
    }

    #[test_only]
    public(package) fun create_zeroed_asset_id(): EncodedAssetId {
        from_bytes(create_zeroed_byte_vector(ASSET_ID_LENGTH))
    }
}