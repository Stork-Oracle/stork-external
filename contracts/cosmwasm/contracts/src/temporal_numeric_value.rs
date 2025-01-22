use sylvia::{cw_schema::cw_serde, cw_std::Int128};

pub type EncodedAssetId = [u8; 32];

#[cw_serde(crate = "sylvia")]
pub struct TemporalNumericValue {
    pub timestamp_ns: u64,
    pub quantized_value: Int128,
}

impl TemporalNumericValue {
    pub fn new(timestamp_ns: u64, quantized_value: i128) -> Self {
        TemporalNumericValue {
            timestamp_ns,
            quantized_value: quantized_value.into(),
        }
    }
}
