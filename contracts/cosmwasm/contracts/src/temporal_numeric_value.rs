use sylvia::{cw_schema::cw_serde, cw_std::{Int128, Uint64}};

pub type EncodedAssetId = [u8; 32];

#[cw_serde(crate = "sylvia")]
pub struct TemporalNumericValue {
    pub timestamp_ns: Uint64,
    pub quantized_value: Int128,
}

impl TemporalNumericValue {
    pub fn new(timestamp_ns: u64, quantized_value: i128) -> Self {
        TemporalNumericValue {
            timestamp_ns: timestamp_ns.into(),
            quantized_value: quantized_value.into(),
        }
    }
}
