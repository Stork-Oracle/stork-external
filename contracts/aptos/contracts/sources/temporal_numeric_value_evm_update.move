module stork::temporal_numeric_value_evm_update {

    // === Imports ===

    use stork::temporal_numeric_value::{Self, TemporalNumericValue};
    use stork::encoded_asset_id::{Self, EncodedAssetId};
    use stork::i128;
    use aptos_std::vector;

    // === Errors ===

    const E_INVALID_LENGTHS: u64 = 0;
    const E_NO_UPDATES: u64 = 1;

    // === Structs ===

    /// Struct representing the input for updating a temporal numeric value
    struct TemporalNumericValueEVMUpdate has copy, drop{
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

    /// Creates a vector of temporal numeric value EVM updates from vectors of the individual fields
    public fun from_vectors(
        /// The asset ids
        ids: vector<vector<u8>>,
        /// The timestamps in nanoseconds
        timestamps_ns: vector<u64>,
        /// The quantized values
        quantized_values: vector<u128>,
        /// The publisher's merkle roots
        publisher_merkle_roots: vector<vector<u8>>,
        /// The value compute algorithm hashes
        value_compute_alg_hashes: vector<vector<u8>>,
        /// The signatures r
        rs: vector<vector<u8>>,
        /// The signatures s
        ss: vector<vector<u8>>,
        /// The signatures v
        vs: vector<u8>,
    ): vector<TemporalNumericValueEVMUpdate> {
        let num_updates = ids.length();

        assert!(num_updates > 0, E_NO_UPDATES);
        assert!(timestamps_ns.length() == num_updates, E_INVALID_LENGTHS);
        assert!(quantized_values.length() == num_updates, E_INVALID_LENGTHS);
        assert!(publisher_merkle_roots.length() == num_updates, E_INVALID_LENGTHS);
        assert!(value_compute_alg_hashes.length() == num_updates, E_INVALID_LENGTHS);
        assert!(rs.length() == num_updates, E_INVALID_LENGTHS);
        assert!(ss.length() == num_updates, E_INVALID_LENGTHS);
        assert!(vs.length() == num_updates, E_INVALID_LENGTHS);

        let updates = vector::empty<TemporalNumericValueEVMUpdate>();
        while (ids.length() > 0) {
            let id = ids.pop_back();
            let timestamp_ns = timestamps_ns.pop_back();
            let quantized_value = quantized_values.pop_back();
            let publisher_merkle_root = publisher_merkle_roots.pop_back();
            let value_compute_alg_hash = value_compute_alg_hashes.pop_back();
            let r = rs.pop_back();
            let s = ss.pop_back();
            let v = vs.pop_back();

            let encoded_asset_id = encoded_asset_id::from_bytes(id);
            let temporal_numeric_value = temporal_numeric_value::new(timestamp_ns, i128::from_u128(quantized_value));
            let update = new(encoded_asset_id, temporal_numeric_value, publisher_merkle_root, value_compute_alg_hash, r, s, v);
            updates.push_back(update);
        };

        updates
    }

    /// Returns the asset id
    public fun get_id(self: &TemporalNumericValueEVMUpdate): EncodedAssetId {
        self.id
    }

    /// Returns the temporal numeric value
    public fun get_temporal_numeric_value(self: &TemporalNumericValueEVMUpdate): TemporalNumericValue {
        self.temporal_numeric_value
    }

    /// Returns the publisher's merkle root
    public fun get_publisher_merkle_root(self: &TemporalNumericValueEVMUpdate): vector<u8> {
        self.publisher_merkle_root
    }

    /// Returns the value compute algorithm hash
    public fun get_value_compute_alg_hash(self: &TemporalNumericValueEVMUpdate): vector<u8> {
        self.value_compute_alg_hash
    }

    /// Returns the signature r
    public fun get_r(self: &TemporalNumericValueEVMUpdate): vector<u8> {
        self.r
    }

    /// Returns the signature s
    public fun get_s(self: &TemporalNumericValueEVMUpdate): vector<u8> {
        self.s
    }

    /// Returns the signature v
    public fun get_v(self: &TemporalNumericValueEVMUpdate): u8 {
        self.v
    }
}