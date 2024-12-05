module stork::update_temporal_numeric_value_evm_input {

    // === Imports ===

    use stork::temporal_numeric_value::{Self, TemporalNumericValue};
    use stork::i128::Self;
    use stork::encoded_asset_id::{Self, EncodedAssetId};
    // === Structs ===

    public struct UpdateTemporalNumericValueEvmInput has copy, drop, store {
        // the id of the asset to update
        id: EncodedAssetId,
        // the temporal numeric value to update
        temporal_numeric_value: TemporalNumericValue,
        // the publisher merkle root
        publisher_merkle_root: vector<u8>,
        // the value compute alg hash
        value_compute_alg_hash: vector<u8>,
        // the r value
        r: vector<u8>,
        // the s value
        s: vector<u8>,
        // the v value
        v: u8,
    }

    // === Public Functions ===

    public fun new(
        id: vector<u8>,
        temporal_numeric_value_timestamp_ns: u64,
        temporal_numeric_value_magnitude: u128,
        temporal_numeric_value_negative: bool,
        publisher_merkle_root: vector<u8>,
        value_compute_alg_hash: vector<u8>,
        r: vector<u8>,
        s: vector<u8>,
        v: u8,
    ): UpdateTemporalNumericValueEvmInput {
        UpdateTemporalNumericValueEvmInput {
            id: encoded_asset_id::from_bytes(id),
            temporal_numeric_value: temporal_numeric_value::new(temporal_numeric_value_timestamp_ns, i128::new(temporal_numeric_value_magnitude, temporal_numeric_value_negative)),
            publisher_merkle_root,
            value_compute_alg_hash,
            r,
            s,
            v,
        }   
    }

    public fun get_id(self: &UpdateTemporalNumericValueEvmInput): EncodedAssetId {
        self.id
    }

    public fun get_temporal_numeric_value(self: &UpdateTemporalNumericValueEvmInput): TemporalNumericValue {
        self.temporal_numeric_value
    }

    public fun get_publisher_merkle_root(self: &UpdateTemporalNumericValueEvmInput): vector<u8> {
        self.publisher_merkle_root
    }

    public fun get_value_compute_alg_hash(self: &UpdateTemporalNumericValueEvmInput): vector<u8> {
        self.value_compute_alg_hash
    }

    public fun get_r(self: &UpdateTemporalNumericValueEvmInput): vector<u8> {
        self.r
    }   

    public fun get_s(self: &UpdateTemporalNumericValueEvmInput): vector<u8> {
        self.s
    }

    public fun get_v(self: &UpdateTemporalNumericValueEvmInput): u8 {
        self.v
    }   

    // === Tests ===

    #[test]
    fun test_temporal_numeric_value_evm_input() {
        let input = new(x"0000000000000000000000000000000000000000000000000000000000000000", 0, 0, false, x"", x"", x"", x"", 0);
        assert!(input.id.get_bytes() == x"0000000000000000000000000000000000000000000000000000000000000000");
        assert!(input.temporal_numeric_value.get_timestamp_ns() == 0);
        assert!(input.temporal_numeric_value.get_quantized_value() == i128::from_u128(0));
        assert!(input.publisher_merkle_root == x"");
        assert!(input.value_compute_alg_hash == x"");
        assert!(input.r == x"");
        assert!(input.s == x"");
        assert!(input.v == 0);
    }
}
