module stork::update_temporal_numeric_value_evm_input_vec {

    // === Imports ===

    use stork::update_temporal_numeric_value_evm_input::{Self, UpdateTemporalNumericValueEvmInput};

    // === Errors ===

    const EInvalidLengths: u64 = 0;
    const ENoUpdates: u64 = 1;
    // === Structs ===

    public struct UpdateTemporalNumericValueEvmInputVec has copy, drop, store {
        data: vector<UpdateTemporalNumericValueEvmInput>,
    }

    // === Public Functions ===

    // Creates a vec of UpdateTemporalNumericValueEvmInput structs
    // takes vecs of inpults, where equal indices in the vecs corresponds to the same asset
    public fun new(
        ids: vector<vector<u8>>,
        temporal_numeric_value_timestamp_nss: vector<u64>,
        temporal_numeric_value_magnitudes: vector<u128>,
        temporal_numeric_value_negatives: vector<bool>,
        publisher_merkle_roots: vector<vector<u8>>,
        value_compute_alg_hashes: vector<vector<u8>>,
        rs: vector<vector<u8>>,
        ss: vector<vector<u8>>,
        vs: vector<u8>,
    ): UpdateTemporalNumericValueEvmInputVec {

        assert!(ids.length() > 0, ENoUpdates);
        // validate the lengths of the vectors
        assert!(ids.length() == temporal_numeric_value_timestamp_nss.length(), EInvalidLengths);
        assert!(ids.length() == temporal_numeric_value_magnitudes.length(), EInvalidLengths);
        assert!(ids.length() == temporal_numeric_value_negatives.length(), EInvalidLengths);
        assert!(ids.length() == publisher_merkle_roots.length(), EInvalidLengths);
        assert!(ids.length() == value_compute_alg_hashes.length(), EInvalidLengths);
        assert!(ids.length() == rs.length(), EInvalidLengths);
        assert!(ids.length() == ss.length(), EInvalidLengths);
        assert!(ids.length() == vs.length(), EInvalidLengths);

        let mut data: vector<UpdateTemporalNumericValueEvmInput> = vector::empty();
        let mut i = 0;
        while (i < ids.length()) {
            let update_temporal_numeric_value_evm_input = update_temporal_numeric_value_evm_input::new(
                ids[i],
                temporal_numeric_value_timestamp_nss[i],
                temporal_numeric_value_magnitudes[i],
                temporal_numeric_value_negatives[i],
                publisher_merkle_roots[i],
                value_compute_alg_hashes[i],
                rs[i],
                ss[i],
                vs[i],
            );
            data.push_back(update_temporal_numeric_value_evm_input);
            i = i + 1;
        };

        UpdateTemporalNumericValueEvmInputVec {
            data,
        }
    }

    public fun get_data(self: &UpdateTemporalNumericValueEvmInputVec): vector<UpdateTemporalNumericValueEvmInput> {
        self.data
    }

    public fun length(self: &UpdateTemporalNumericValueEvmInputVec): u64 {
        vector::length(&self.data)
    }

    // === Test Imports ===

    #[test_only] use stork::i128;

    // === Tests ===

    #[test]
    #[expected_failure(abort_code = ENoUpdates)]
    fun test_new_no_updates() {
        let _ = new(vector::empty(), vector::empty(), vector::empty(), vector::empty(), vector::empty(), vector::empty(), vector::empty(), vector::empty(), vector::empty());
    }

    #[test]
    #[expected_failure(abort_code = EInvalidLengths)]
    fun test_new_invalid_lengths() {
        let mut ids = vector::empty();
        // add 4 elements to the vector
        ids.push_back(b"00000000000000000000000000000000");
        ids.push_back(b"00000000000000000000000000000001");
        ids.push_back(b"00000000000000000000000000000002");
        ids.push_back(b"00000000000000000000000000000003");

        let mut temporal_numeric_value_timestamp_nss = vector::empty();
        // add 3 elements to the vector
        temporal_numeric_value_timestamp_nss.push_back(0);
        temporal_numeric_value_timestamp_nss.push_back(0);
        temporal_numeric_value_timestamp_nss.push_back(0);

        let mut temporal_numeric_value_magnitudes = vector::empty();    
        // add 4 elements to the vector
        temporal_numeric_value_magnitudes.push_back(0);
        temporal_numeric_value_magnitudes.push_back(0);
        temporal_numeric_value_magnitudes.push_back(0);
        temporal_numeric_value_magnitudes.push_back(0);

        let mut temporal_numeric_value_negatives = vector::empty();
        // add 4 elements to the vector
        temporal_numeric_value_negatives.push_back(false);
        temporal_numeric_value_negatives.push_back(false);
        temporal_numeric_value_negatives.push_back(false);
        temporal_numeric_value_negatives.push_back(false);

        let mut publisher_merkle_roots = vector::empty();
        // add 4 elements to the vector
        publisher_merkle_roots.push_back(x"");
        publisher_merkle_roots.push_back(x"");
        publisher_merkle_roots.push_back(x"");
        publisher_merkle_roots.push_back(x"");

        let mut value_compute_alg_hashes = vector::empty();
        // add 4 elements to the vector
        value_compute_alg_hashes.push_back(x"");
        value_compute_alg_hashes.push_back(x"");
        value_compute_alg_hashes.push_back(x"");
        value_compute_alg_hashes.push_back(x"");

        let mut rs = vector::empty();
        // add 4 elements to the vector
        rs.push_back(x"");
        rs.push_back(x"");
        rs.push_back(x"");
        rs.push_back(x"");

        let mut ss = vector::empty();
        // add 4 elements to the vector
        ss.push_back(x"");
        ss.push_back(x"");
        ss.push_back(x"");
        ss.push_back(x"");

        let mut vs = vector::empty();
        // add 4 elements to the vector
        vs.push_back(0);
        vs.push_back(0);
        vs.push_back(0);
        vs.push_back(0);

        let _ = new(ids, temporal_numeric_value_timestamp_nss, temporal_numeric_value_magnitudes, temporal_numeric_value_negatives, publisher_merkle_roots, value_compute_alg_hashes, rs, ss, vs);
    }

    #[test]
    fun test_length() {
        let mut ids = vector::empty();
        ids.push_back(b"00000000000000000000000000000000");
        ids.push_back(b"00000000000000000000000000000001");
        ids.push_back(b"00000000000000000000000000000002");
        ids.push_back(b"00000000000000000000000000000003");

        let mut temporal_numeric_value_timestamp_nss = vector::empty();
        temporal_numeric_value_timestamp_nss.push_back(0);
        temporal_numeric_value_timestamp_nss.push_back(0);
        temporal_numeric_value_timestamp_nss.push_back(0);
        temporal_numeric_value_timestamp_nss.push_back(0);

        let mut temporal_numeric_value_magnitudes = vector::empty();
        temporal_numeric_value_magnitudes.push_back(0);
        temporal_numeric_value_magnitudes.push_back(0);
        temporal_numeric_value_magnitudes.push_back(0);
        temporal_numeric_value_magnitudes.push_back(0);

        let mut temporal_numeric_value_negatives = vector::empty();
        temporal_numeric_value_negatives.push_back(false);
        temporal_numeric_value_negatives.push_back(false);
        temporal_numeric_value_negatives.push_back(false);
        temporal_numeric_value_negatives.push_back(false);

        let mut publisher_merkle_roots = vector::empty();
        publisher_merkle_roots.push_back(x"");
        publisher_merkle_roots.push_back(x"");
        publisher_merkle_roots.push_back(x"");
        publisher_merkle_roots.push_back(x"");

        let mut value_compute_alg_hashes = vector::empty();
        value_compute_alg_hashes.push_back(x"");
        value_compute_alg_hashes.push_back(x"");
        value_compute_alg_hashes.push_back(x"");
        value_compute_alg_hashes.push_back(x"");

        let mut rs = vector::empty();
        rs.push_back(x"");
        rs.push_back(x"");
        rs.push_back(x"");
        rs.push_back(x"");

        let mut ss = vector::empty();
        ss.push_back(x"");
        ss.push_back(x"");
        ss.push_back(x"");
        ss.push_back(x"");

        let mut vs = vector::empty();
        vs.push_back(0);
        vs.push_back(0);
        vs.push_back(0);
        vs.push_back(0);

        let input_vec = new(ids, temporal_numeric_value_timestamp_nss, temporal_numeric_value_magnitudes, temporal_numeric_value_negatives, publisher_merkle_roots, value_compute_alg_hashes, rs, ss, vs);
        assert!(length(&input_vec) == 4);
    }


    #[test]
    fun test_get_data() {
        let mut ids = vector::empty();
        ids.push_back(b"00000000000000000000000000000000");
        ids.push_back(b"00000000000000000000000000000001");
        ids.push_back(b"00000000000000000000000000000002");
        ids.push_back(b"00000000000000000000000000000003");

        let mut temporal_numeric_value_timestamp_nss = vector::empty();
        temporal_numeric_value_timestamp_nss.push_back(0);
        temporal_numeric_value_timestamp_nss.push_back(0);
        temporal_numeric_value_timestamp_nss.push_back(0);
        temporal_numeric_value_timestamp_nss.push_back(0);

        let mut temporal_numeric_value_magnitudes = vector::empty();
        temporal_numeric_value_magnitudes.push_back(0);
        temporal_numeric_value_magnitudes.push_back(0);
        temporal_numeric_value_magnitudes.push_back(0);
        temporal_numeric_value_magnitudes.push_back(0);

        let mut temporal_numeric_value_negatives = vector::empty();
        temporal_numeric_value_negatives.push_back(false);
        temporal_numeric_value_negatives.push_back(false);
        temporal_numeric_value_negatives.push_back(false);
        temporal_numeric_value_negatives.push_back(false);

        let mut publisher_merkle_roots = vector::empty();
        publisher_merkle_roots.push_back(x"");
        publisher_merkle_roots.push_back(x"");
        publisher_merkle_roots.push_back(x"");
        publisher_merkle_roots.push_back(x"");

        let mut value_compute_alg_hashes = vector::empty();
        value_compute_alg_hashes.push_back(x"");
        value_compute_alg_hashes.push_back(x"");
        value_compute_alg_hashes.push_back(x"");
        value_compute_alg_hashes.push_back(x"");

        let mut rs = vector::empty();
        rs.push_back(x"");
        rs.push_back(x"");
        rs.push_back(x"");
        rs.push_back(x"");

        let mut ss = vector::empty();
        ss.push_back(x"");
        ss.push_back(x"");
        ss.push_back(x"");
        ss.push_back(x"");

        let mut vs = vector::empty();
        vs.push_back(0);
        vs.push_back(0);
        vs.push_back(0);
        vs.push_back(0);

        let input_vec = new(ids, temporal_numeric_value_timestamp_nss, temporal_numeric_value_magnitudes, temporal_numeric_value_negatives, publisher_merkle_roots, value_compute_alg_hashes, rs, ss, vs);
        let data = get_data(&input_vec);
        assert!(vector::length(&data) == 4);

        let mut i = 0;
        while (i < 4) {
            assert!(update_temporal_numeric_value_evm_input::get_id(&data[i]).get_bytes() == ids[i]);
            assert!(update_temporal_numeric_value_evm_input::get_temporal_numeric_value(&data[i]).get_timestamp_ns() == temporal_numeric_value_timestamp_nss[i]);
            assert!(update_temporal_numeric_value_evm_input::get_temporal_numeric_value(&data[i]).get_quantized_value() == i128::from_u128(temporal_numeric_value_magnitudes[i]));
            assert!(update_temporal_numeric_value_evm_input::get_temporal_numeric_value(&data[i]).get_quantized_value().is_negative() == temporal_numeric_value_negatives[i]);
            assert!(update_temporal_numeric_value_evm_input::get_publisher_merkle_root(&data[i]) == publisher_merkle_roots[i]);
            assert!(update_temporal_numeric_value_evm_input::get_value_compute_alg_hash(&data[i]) == value_compute_alg_hashes[i]);
            assert!(update_temporal_numeric_value_evm_input::get_r(&data[i]) == rs[i]);
            assert!(update_temporal_numeric_value_evm_input::get_s(&data[i]) == ss[i]);
            assert!(update_temporal_numeric_value_evm_input::get_v(&data[i]) == vs[i]);
            i = i + 1;
        };
    }
}
