#[cfg(not(feature = "library"))]
use cosmwasm_std::entry_point;
use cosmwasm_std::{DepsMut, Env, MessageInfo, Response, StdError};

// The sv module is the module that contains the sylvia macro expansion of the stork contract
// It is not present when reading the stork contract unless you expand the macro with a tool like rust-analyzer
// It contains useful things like the QueryMsg, ExecuteMsg, and more, and those messages can be imported like this:
use stork_cw::contract::sv::QueryMsg::GetLatestCanonicalTemporalNumericValueUnchecked;

use stork_cw::responses::GetTemporalNumericValueResponse;
use crate::msg::{ExecuteMsg, InstantiateMsg};
use crate::state::{State, STATE};
use cosmwasm_std::Event;

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn instantiate(
    deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    msg: InstantiateMsg,
) -> Result<Response, StdError> {
    let state = State {
        stork: deps.api.addr_validate(&msg.stork_contract_address)?,
    };
    STATE.save(deps.storage, &state)?;
    Ok(Response::new())
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn execute(
    deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    msg: ExecuteMsg,
) -> Result<Response, StdError> {
    match msg {
        ExecuteMsg::UseStorkPrice { feed_id } => {
            let state = STATE.load(deps.storage)?;
            let message = GetLatestCanonicalTemporalNumericValueUnchecked {
                id: feed_id,
            };
            let response: GetTemporalNumericValueResponse = deps.querier.query_wasm_smart(state.stork, &message)?;
            let temporal_numeric_value = response.temporal_numeric_value;
            let feed_id_hex = feed_id.iter().map(|b| format!("{:02x}", b)).collect::<Vec<String>>().join("");
            let event = Event::new("stork_price_used").add_attribute("feed_id", feed_id_hex).add_attribute("value", temporal_numeric_value.quantized_value.to_string());
            Ok(Response::new().add_events(vec![event]))
        }
    }
}
