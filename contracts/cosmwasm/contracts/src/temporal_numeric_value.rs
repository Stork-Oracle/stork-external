//! Provides the [`TemporalNumericValue`] struct.

use sylvia::{
    cw_schema::cw_serde,
    cw_std::{Int128, Uint64},
};

/// The type for encoded asset ids. This is an alias for a 32 byte array.
pub type EncodedAssetId = [u8; 32];

/// A struct representing a timestamped value.
#[cw_serde(crate = "sylvia")]
pub struct TemporalNumericValue {
    /// The unix timestamp of the value in nanoseconds.
    pub timestamp_ns: Uint64,
    /// The quantized value.
    pub quantized_value: Int128,
}

impl TemporalNumericValue {
    /// Creates a new [`TemporalNumericValue`] struct.
    pub fn new(timestamp_ns: u64, quantized_value: i128) -> Self {
        TemporalNumericValue {
            timestamp_ns: timestamp_ns.into(),
            quantized_value: quantized_value.into(),
        }
    }
}
