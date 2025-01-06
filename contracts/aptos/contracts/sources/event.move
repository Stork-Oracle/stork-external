module stork::event {

    // === Events ===

    #[event]
    /// Event emitted when the Stork contract is initialized
    struct StorkInitializationEvent has drop, store{
        stork_address: address,
        stork_evm_public_key: EvmPubKey,
        single_update_fee: u64,
        owner: address,
    }

    #[event]
    /// Event emitted when a temporal numeric value is updated
    struct TemporalNumericValueUpdateEvent has drop, store {
        asset_id: EncodedAssetId,
        temporal_numeric_value: TemporalNumericValue,
    }

}