module stork::event {

    // === Imports ===

    use sui::event;
    use stork::temporal_numeric_value::TemporalNumericValue;
    use stork::encoded_asset_id::EncodedAssetId; 
    // === Structs ===

    public struct StorkInitializationEvent has copy, drop {
        stork_sui_public_key: address,
        stork_evm_public_key: vector<u8>,
        single_update_fee: u64,
        stork_state_id: ID,
        stork_state_version: u64,
    }

    public struct FeeWithdrawalEvent has copy, drop{
        amount: u64,
    }

    public struct TemporalNumericValueFeedUpdateEvent has copy, drop {
        asset_id: EncodedAssetId,
        value: TemporalNumericValue,
    }

    // === Functions ===

    public(package) fun emit_stork_initialization_event(
        stork_sui_public_key: address,
        stork_evm_public_key: vector<u8>,
        single_update_fee: u64,
        stork_state_id: ID,
        stork_state_version: u64,
    ) {
        event::emit(
            StorkInitializationEvent {
                stork_sui_public_key,
                stork_evm_public_key,
                single_update_fee,
                stork_state_id,
                stork_state_version,
            }
        );
    }

    public(package) fun emit_fee_withdrawal_event(
        amount: u64,
    ) {
        event::emit(FeeWithdrawalEvent {amount });
    }

    public(package) fun emit_temporal_numeric_value_feed_update_event(
        asset_id: EncodedAssetId,
        value: TemporalNumericValue,
    ) {
        event::emit(TemporalNumericValueFeedUpdateEvent {asset_id, value});
    }
}