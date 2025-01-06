module stork::temporal_numeric_value_evm_updats {

    // === Imports ===

    use stork::temporal_numeric_value::TemporalNumericValue;
    use stork::encoded_asset_id::EncodedAssetId;

    // === Structs ===

    /// Struct representing the input for updating a temporal numeric value
    struct TemporalNumericValueEVMUpdate {
        /// The asset id
        id: EncodedAssetId,
        /// The temporal numeric value
        temporal_numeric_value: TemporalNumericValue,
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
    }

    // === Functions ===

    /// Creates a new update temporal numeric value EVM input   
    public fun new(
        /// The asset id
        id: EncodedAssetId,
        /// The temporal numeric value
        temporal_numeric_value: TemporalNumericValue,
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
    ): TemporalNumericValueEVMUpdate { 
        TemporalNumericValueEVMUpdate {
            id,
            temporal_numeric_value,
            publisher_merkle_root,
            value_compute_alg_hash,
            r,
            s,
            v,
        }
    }

    // === Public Functions ===

    /// Returns the asset id
    public fun get_id(input: &TemporalNumericValueEVMUpdate): EncodedAssetId {
        input.id
    }

    /// Returns the temporal numeric value
    public fun get_temporal_numeric_value(input: &TemporalNumericValueEVMUpdate): TemporalNumericValue {
        input.temporal_numeric_value
    }

    /// Returns the publisher's merkle root
    public fun get_publisher_merkle_root(input: &TemporalNumericValueEVMUpdate): vector<u8> {
        input.publisher_merkle_root
    }

    /// Returns the value compute algorithm hash
    public fun get_value_compute_alg_hash(input: &TemporalNumericValueEVMUpdate): vector<u8> {
        input.value_compute_alg_hash
    }

    /// Returns the signature r
    public fun get_r(input: &TemporalNumericValueEVMUpdate): vector<u8> {
        input.r
    }

    /// Returns the signature s
    public fun get_s(input: &TemporalNumericValueEVMUpdate): vector<u8> {
        input.s
    }

    /// Returns the signature v
    public fun get_v(input: &TemporalNumericValueEVMUpdate): u8 {
        input.v
    }
}