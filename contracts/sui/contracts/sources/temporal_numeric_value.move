module stork::temporal_numeric_value {
    // === Imports ===

    use stork::i128::I128;

    // === Structs ===

    public struct TemporalNumericValue has store, copy, drop{
        // the timestamp in nanoseconds
        timestamp_ns: u64,
        // the quantized value
        quantized_value: I128,
    }

    // === Functions ===

    public fun new(timestamp_ns: u64, quantized_value: I128): TemporalNumericValue {
        TemporalNumericValue {
            timestamp_ns,
            quantized_value,
        }
    }

    // returns the timestamp in nanoseconds
    public fun get_timestamp_ns(value: &TemporalNumericValue): u64 {
        value.timestamp_ns
    }

    // returns the quantized value
    public fun get_quantized_value(value: &TemporalNumericValue): I128 {
        value.quantized_value
    }

    // === Test Imports ===

    #[test_only] use stork::i128;

    // === Tests ===

    #[test]
    fun test_get_timestamp_ns() {
        let value = new(1000, i128::from_u128(1000));
        assert!(get_timestamp_ns(&value) == 1000);
    }

    #[test]
    fun test_get_quantized_value() {
        let value = new(1000, i128::from_u128(1000));
        assert!(value.get_quantized_value() == i128::from_u128(1000));
    }

    // === Test Helpers ===

    #[test_only]
    public(package) fun create_zeroed_temporal_numeric_value(): TemporalNumericValue {
        new(0, i128::from_u128(0))
    }
}