library;

use signed_int::i128::I128;

// Struct representing a temporal numeric value
pub struct TemporalNumericValue {
    // The timestamp in nanoseconds
    pub timestamp_ns: u64,
    // The quantized value
    pub quantized_value: I128,
}

impl TemporalNumericValue {
    // Creates a new temporal numeric value
    pub fn new(timestamp_ns: u64, quantized_value: I128) -> Self {
        Self {
            timestamp_ns,
            quantized_value,
        }
    }

    // Returns the timestamp in nanoseconds
    pub fn get_timestamp_ns(self) -> u64 {
        self.timestamp_ns
    }

    // Returns the quantized value
    pub fn get_quantized_value(self) -> I128 {
        self.quantized_value
    }
}

#[test]
fn test_get_timestamp_ns() {
    let temporal_numeric_value = TemporalNumericValue::new(1000u64, I128::new());
    assert(temporal_numeric_value.get_timestamp_ns() == 1000u64);
}

#[test]
fn test_get_quantized_value() {
    let temporal_numeric_value = TemporalNumericValue::new(1000u64, I128::new());
    assert(temporal_numeric_value.get_quantized_value() == I128::new());
}
