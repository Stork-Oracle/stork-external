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
        // asset id
        id: EncodedAssetId,
        // temporal numeric value
        temporal_numeric_value: TemporalNumericValue,
        // publisher's merkle root
        publisher_merkle_root: vector<u8>,
        // value compute algorithm hash
        value_compute_alg_hash: vector<u8>,
        // signature r
        r: vector<u8>,
        // signature s
        s: vector<u8>,
        // signature v
        v: u8,
    }

    // === Functions ===

    /// Creates a new update temporal numeric value EVM input   
    public fun new(
        // asset id
        id: EncodedAssetId,
        // temporal numeric value
        temporal_numeric_value: TemporalNumericValue,
        // publisher's merkle root
        publisher_merkle_root: vector<u8>,
        // value compute algorithm hash
        value_compute_alg_hash: vector<u8>,
        // signature r
        r: vector<u8>,
        // signature s
        s: vector<u8>,
        // signature v
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
        // asset ids
        ids: vector<vector<u8>>,
        // timestamps in nanoseconds
        timestamps_ns: vector<u64>,
        // temporal numeric value magnitudes
        magnitudes: vector<u128>,
        // temporal numeric value negatives
        negatives: vector<bool>,
        // publisher's merkle roots
        publisher_merkle_roots: vector<vector<u8>>,
        // value compute algorithm hashes
        value_compute_alg_hashes: vector<vector<u8>>,
        // signatures r
        rs: vector<vector<u8>>,
        // signatures s
        ss: vector<vector<u8>>,
        // signatures v
        vs: vector<u8>,
    ): vector<TemporalNumericValueEVMUpdate> {
        let num_updates = ids.length();

        assert!(num_updates > 0, E_NO_UPDATES);
        assert!(timestamps_ns.length() == num_updates, E_INVALID_LENGTHS);
        assert!(magnitudes.length() == num_updates, E_INVALID_LENGTHS);
        assert!(negatives.length() == num_updates, E_INVALID_LENGTHS);
        assert!(publisher_merkle_roots.length() == num_updates, E_INVALID_LENGTHS);
        assert!(value_compute_alg_hashes.length() == num_updates, E_INVALID_LENGTHS);
        assert!(rs.length() == num_updates, E_INVALID_LENGTHS);
        assert!(ss.length() == num_updates, E_INVALID_LENGTHS);
        assert!(vs.length() == num_updates, E_INVALID_LENGTHS);

        let updates = vector::empty<TemporalNumericValueEVMUpdate>();
        while (ids.length() > 0) {
            let id = ids.pop_back();
            let timestamp_ns = timestamps_ns.pop_back();
            let magnitude = magnitudes.pop_back();
            let negative = negatives.pop_back();
            let publisher_merkle_root = publisher_merkle_roots.pop_back();
            let value_compute_alg_hash = value_compute_alg_hashes.pop_back();
            let r = rs.pop_back();
            let s = ss.pop_back();
            let v = vs.pop_back();

            let encoded_asset_id = encoded_asset_id::from_bytes(id);
            let temporal_numeric_value = temporal_numeric_value::new(timestamp_ns, i128::new(magnitude, negative));
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

    // === Tests ===

    #[test]
    fun test_new() {
        let id = encoded_asset_id::create_zeroed_asset_id();
        let tnv = temporal_numeric_value::create_zeroed_temporal_numeric_value();
        let publisher_merkle_root = vector::empty();
        let value_compute_alg_hash = vector::empty();
        let r = vector::empty();
        let s = vector::empty();
        let v = 0;

        let update = new(
            id,
            tnv,
            publisher_merkle_root,
            value_compute_alg_hash,
            r,
            s,
            v
        );

        assert!(get_id(&update) == id, 0);
        assert!(get_temporal_numeric_value(&update) == tnv, 1);
        assert!(get_publisher_merkle_root(&update) == publisher_merkle_root, 2);
        assert!(get_value_compute_alg_hash(&update) == value_compute_alg_hash, 3);
        assert!(get_r(&update) == r, 4);
        assert!(get_s(&update) == s, 5);
        assert!(get_v(&update) == v, 6);
    }

    #[test]
    fun test_from_vectors() {
        let ids = vector::singleton(encoded_asset_id::get_bytes(&encoded_asset_id::create_zeroed_asset_id()));
        let timestamps_ns = vector::singleton(0u64);
        let magnitudes = vector::singleton(0u128);
        let negatives = vector::singleton(false);
        let publisher_merkle_roots = vector::singleton(vector::empty());
        let value_compute_alg_hashes = vector::singleton(vector::empty());
        let rs = vector::singleton(vector::empty());
        let ss = vector::singleton(vector::empty());
        let vs = vector::singleton(0u8);

        let updates = from_vectors(
            ids,
            timestamps_ns,
            magnitudes,
            negatives,
            publisher_merkle_roots,
            value_compute_alg_hashes,
            rs,
            ss,
            vs
        );

        assert!(vector::length(&updates) == 1, 0);
        
        let update = vector::borrow(&updates, 0);
        assert!(get_v(update) == 0, 1);
    }

    #[test]
    #[expected_failure(abort_code = E_NO_UPDATES)]
    fun test_from_vectors_empty() {
        let ids = vector::empty();
        let timestamps_ns = vector::empty();
        let magnitudes = vector::empty();
        let negatives = vector::empty();
        let publisher_merkle_roots = vector::empty();
        let value_compute_alg_hashes = vector::empty();
        let rs = vector::empty();
        let ss = vector::empty();
        let vs = vector::empty();

        from_vectors(
            ids,
            timestamps_ns,
            magnitudes,
            negatives,
            publisher_merkle_roots,
            value_compute_alg_hashes,
            rs,
            ss,
            vs
        );
    }

    #[test]
    #[expected_failure(abort_code = E_INVALID_LENGTHS)]
    fun test_from_vectors_mismatched_lengths() {
        let ids = vector::singleton(encoded_asset_id::get_bytes(&encoded_asset_id::create_zeroed_asset_id()));
        let timestamps_ns = vector::empty();  // Different length than ids
        let magnitudes = vector::singleton(0u128);
        let negatives = vector::singleton(false);
        let publisher_merkle_roots = vector::singleton(vector::empty());
        let value_compute_alg_hashes = vector::singleton(vector::empty());
        let rs = vector::singleton(vector::empty());
        let ss = vector::singleton(vector::empty());
        let vs = vector::singleton(0u8);

        from_vectors(
            ids,
            timestamps_ns,
            magnitudes,
            negatives,
            publisher_merkle_roots,
            value_compute_alg_hashes,
            rs,
            ss,
            vs
        );
    }
}