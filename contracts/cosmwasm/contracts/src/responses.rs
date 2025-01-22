use crate::temporal_numeric_value::TemporalNumericValue;
use crate::verify::EvmPubkey;
use sylvia::cw_schema::cw_serde;
use sylvia::cw_std::Addr;
use sylvia::cw_std::Coin;

#[cw_serde(crate = "sylvia")]
pub struct GetSingleUpdateFeeResponse {
    pub fee: Coin,
}

#[cw_serde(crate = "sylvia")]
pub struct GetTemporalNumericValueResponse {
    pub temporal_numeric_value: TemporalNumericValue,
}

#[cw_serde(crate = "sylvia")]
pub struct GetStorkEvmPublicKeyResponse {
    pub stork_evm_public_key: EvmPubkey,
}

#[cw_serde(crate = "sylvia")]
pub struct GetOwnerResponse {
    pub owner: Addr,
}
