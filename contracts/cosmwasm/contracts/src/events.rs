//! Event creation functions for the Stork CosmWasm Contract. The event names are typically automatically prefixed with `wasm-` by the comsos runtime.
use crate::{
    temporal_numeric_value::{EncodedAssetId, TemporalNumericValue},
    verify::EvmPubkey,
};
use sylvia::{cw_std::Coin, cw_std::Addr, cw_std::Event};

/// Creates an event for the Stork contract initialization with the name `stork_init`.
/// This event contains important information about the state of the Stork contract.
pub(crate) fn new_stork_init_event(
    stork_evm_public_key: EvmPubkey,
    single_update_fee: Coin,
    owner: Addr,
) -> Event {
    Event::new("stork_init")
        .add_attribute("stork_evm_public_key", byte_array_to_hex_string(&stork_evm_public_key))
        .add_attribute("single_update_fee", single_update_fee.amount)
        .add_attribute("single_update_fee_denom", single_update_fee.denom)
        .add_attribute("owner", owner)
}

/// Creates an event for the Stork contract initialization with the name `temporal_numeric_value_update`.
/// This event is emitted whenever a new price update is successfully submitted to the Stork contract.
pub(crate) fn new_temporal_numeric_value_update_event(
    id: EncodedAssetId,
    value: TemporalNumericValue,
) -> Event {
    Event::new("temporal_numeric_value_update")
        .add_attribute("id", byte_array_to_hex_string(&id))
        .add_attribute("value", value.timestamp_ns.to_string())
        .add_attribute("value", value.quantized_value)
}

fn byte_array_to_hex_string(bytes: &[u8]) -> String {
    bytes.iter().map(|b| format!("{:02x}", b)).collect()
}
