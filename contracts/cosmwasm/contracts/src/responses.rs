//! Responses for the Stork Cosmwasm Contract.
//! These are the responses returned from queries to the Stork contract.
use crate::temporal_numeric_value::TemporalNumericValue;
use crate::verify::EvmPubkey;
use sylvia::cw_schema::cw_serde;
use sylvia::cw_std::Addr;
use sylvia::cw_std::Coin;

/// Response for the `get_single_update_fee` query containing the fee for a single update.
#[cw_serde(crate = "sylvia")]
pub struct GetSingleUpdateFeeResponse {
    pub fee: Coin,
}

/// Response for the `get_temporal_numeric_value` query containing the [`TemporalNumericValue`](./temporal_numeric_value.rs) for a given asset id.
#[cw_serde(crate = "sylvia")]
pub struct GetTemporalNumericValueResponse {
    pub temporal_numeric_value: TemporalNumericValue,
}

/// Response for the `get_stork_evm_public_key` query containing the EVM public key set in the Stork contract.
/// This is typically the Stork Aggregator's public key
#[cw_serde(crate = "sylvia")]
pub struct GetStorkEvmPublicKeyResponse {
    pub stork_evm_public_key: EvmPubkey,
}

/// Response for the `get_owner` query containing the address owner of the Stork contract.
#[cw_serde(crate = "sylvia")]
pub struct GetOwnerResponse {
    pub owner: Addr,
}
