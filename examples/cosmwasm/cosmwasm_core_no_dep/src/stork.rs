// These are the types that can be used to consume feeds fromthe Stork Cosmwasm Contract
// Normally, they would be pulled in from the stork-cw crate, but in this example, we simply define them ourselves
// Ensure they match the types in the stork-cw crate
// Copy and pasting this file into your contract is a good way to ensure you are using the correct types

use cosmwasm_std::{Int128, Uint64};
use cosmwasm_schema::cw_serde;

// Structs
// A struct representing a timestamped value
#[cw_serde]
pub struct TemporalNumericValue {
    // The unix timestamp of the value in nanoseconds
    pub timestamp_ns: Uint64,
    // The quantized value
    pub quantized_value: Int128,
}

// The type for encoded asset ids. This is an alias for a 32 byte array.
pub type EncodedAssetId = [u8; 32];

// Query Messages
// Request 
#[cw_serde]
pub struct GetTemporalNumericValueRequest {
    pub id: EncodedAssetId,
}

// Response
#[cw_serde]
pub struct GetTemporalNumericValueResponse {
    pub temporal_numeric_value: TemporalNumericValue,
}


