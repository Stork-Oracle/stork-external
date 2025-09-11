module stork::event {

    // === Imports ===

    use aptos_std::event;
    use stork::temporal_numeric_value::TemporalNumericValue;
    use stork::encoded_asset_id::EncodedAssetId;
    use stork::evm_pubkey::EvmPubKey;

    // === Events ===

    #[event]
    /// Event emitted when the Stork contract is initialized
    struct StorkInitializationEvent has drop, store{
        stork_address: address,
        stork_evm_public_key: EvmPubKey,
        single_update_fee: u64,
        owner: address,
        state_account_address: address,
    }

    #[event]
    /// Event emitted when a temporal numeric value is updated
    struct TemporalNumericValueUpdateEvent has drop, store {
        asset_id: EncodedAssetId,
        temporal_numeric_value: TemporalNumericValue,
    }

    package fun emit_stork_initialization_event(
        stork_address: address,
        stork_evm_public_key: EvmPubKey,
        single_update_fee: u64,
        owner: address,
        state_account_address: address,
    ) {
        event::emit(StorkInitializationEvent { stork_address, stork_evm_public_key, single_update_fee, owner, state_account_address });
    }

    package fun emit_temporal_numeric_value_update_event(
        asset_id: EncodedAssetId,
        temporal_numeric_value: TemporalNumericValue,
    ) {
        event::emit(TemporalNumericValueUpdateEvent { asset_id, temporal_numeric_value });
    }

}