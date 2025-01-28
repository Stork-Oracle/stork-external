use cosmwasm_schema::{cw_serde, QueryResponses};
use stork_cw::temporal_numeric_value::EncodedAssetId;

#[cw_serde]
pub struct InstantiateMsg {
    pub stork_contract_address: String
}

#[cw_serde]
pub enum ExecuteMsg {
    UseStorkPrice { feed_id: EncodedAssetId }
}

#[cw_serde]
#[derive(QueryResponses)]
pub enum QueryMsg {}
