module stork::update_temporal_numeric_value_evm_input_vec {

    // === Imports ===

    use stork::update_temporal_numeric_value_evm_input::{Self, UpdateTemporalNumericValueEvmInput};

    // === Errors ===

    const EInvalidLengths: u64 = 0;

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
}