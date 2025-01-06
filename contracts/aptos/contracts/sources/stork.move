module stork::stork {

    // === Imports ===

    use stork::state::{StorkState, new};
    use stork::event::StorkInitializationEvent;
    use stork::temporal_numeric_value_feed::TemporalNumericValueFeed;
    use stork::encoded_asset_id::EncodedAssetId;
    use aptos_std::table;
    use stork::temporal_numeric_value_feed_registry::TemporalNumericValueFeedRegistry;
    use stork::temporal_numeric_value_evm_update::TemporalNumericValueEVMUpdate;
    use stork::verify::verify_evm_signature;

    // === Errors ===

    const E_NOT_OWNER: u64 = 0;
    const E_ALREADY_INITIALIZED: u64 = 1;
    const E_INVALID_SIGNATURE: u64 = 2;
    const E_FEED_NOT_FOUND: u64 = 3;

    // === Functions ===

    entry fun init_stork(
        stork_evm_public_key: EvmPubKey,
        single_update_fee: u64,
        owner: address,
    ) {
        assert!(
            !exists<StorkState>(@stork),
            E_ALREADY_INITIALIZED
        );
        let state = state::new(stork_evm_public_key, single_update_fee, owner);
        move_to(@stork, state);

        // TNV feed table
        let feed_table = temporal_numeric_value_feed_registry::new();
        move_to(@stork, feed_table);

        let stork_initialized_event = StorkInitializationEvent {
            stork_address: @stork,
            stork_evm_public_key,
            single_update_fee,
            owner,
        };
        event::emit(stork_initialized_event);
    }

    // === Public Functions ===

    /// Updates a single temporal numeric value
    public entry fun update_single_temporal_numeric_value(
        /// The signer of the transaction to pay the fee
        signer: &signer,
        /// The update to the temporal numeric value
        update: TemporalNumericValueEVMUpdate,
    ) {
        let state = borrow_global<StorkState>(@stork);
        let fee = state::get_single_update_fee(&state);
        let evm_pubkey = state::get_stork_evm_public_key(&state);
        let feed_registry = borrow_global_mut<TemporalNumericValueFeedRegistry>(@stork);
        let temporal_numeric_value = temporal_numeric_value_evm_update::get_temporal_numeric_value(&update);

        // recency
        if (table::contains(&feed_registry.feed_table, temporal_numeric_value_evm_update::get_id(&update))) {
            let existing_temporal_numeric_value = table::borrow(&feed_registry.feed_table, temporal_numeric_value_evm_update::get_id(&update));
            if (temporal_numeric_value.timestamp <= existing_temporal_numeric_value.timestamp) {
                return;
            };
        };

        // verify signature
        assert!(
            verify::verify_evm_signature(
                &evm_pubkey,
                encoded_asset_id::get_bytes(&temporal_numeric_value_evm_update::get_id(&update)),
                temporal_numeric_value::get_timestamp_ns(&temporal_numeric_value),
                temporal_numeric_value::get_quantized_value(&temporal_numeric_value),
                temporal_numeric_value_evm_update::get_publisher_merkle_root(&update),
                temporal_numeric_value_evm_update::get_value_compute_alg_hash(&update),
                temporal_numeric_value_evm_update::get_r(&update),
                temporal_numeric_value_evm_update::get_s(&update),
                temporal_numeric_value_evm_update::get_v(&update),
            ),
            E_INVALID_SIGNATURE
        );
        
        primary_fungible_store::transfer(signer, AptosCoin, @stork, fee);
        temporal_numeric_value_feed_registry::update_latest_temporal_numeric_value(&mut feed_registry, temporal_numeric_value_evm_update::get_id(&update), temporal_numeric_value);

    }

    public entry fun update_multiple_temporal_numeric_values_EVM(
        /// The signer of the transaction to pay the fee
        signer: &signer,
        /// The updates to the temporal numeric values
        updates: vector<TemporalNumericValueEVMUpdate>, 
    ) {
        let state = borrow_global<StorkState>(@stork);
        let evm_pubkey = state::get_stork_evm_public_key(&state);
        let fee = state::get_single_update_fee(&state);
        let feed_registry = borrow_global_mut<TemporalNumericValueFeedRegistry>(@stork);

        let num_updates = 0;

        while (updates.len() > 0) {
            let update = vector::pop_back(&mut updates);
            let temporal_numeric_value = temporal_numeric_value_evm_update::get_temporal_numeric_value(&update);

            // recency
            if (table::contains(&feed_registry.feed_table, temporal_numeric_value_evm_update::get_id(&update))) {
                let existing_temporal_numeric_value = table::borrow(&feed_registry.feed_table, temporal_numeric_value_evm_update::get_id(&update));
                if (temporal_numeric_value.timestamp <= existing_temporal_numeric_value.timestamp) {
                    continue;
                };
            };

            // verify signature
            assert!(
                verify::verify_evm_signature(
                    &evm_pubkey,
                    encoded_asset_id::get_bytes(&temporal_numeric_value_evm_update::get_id(&update)),
                    temporal_numeric_value::get_timestamp_ns(&temporal_numeric_value),
                    temporal_numeric_value::get_quantized_value(&temporal_numeric_value),
                    temporal_numeric_value_evm_update::get_publisher_merkle_root(&update),
                    temporal_numeric_value_evm_update::get_value_compute_alg_hash(&update),
                    temporal_numeric_value_evm_update::get_r(&update),
                    temporal_numeric_value_evm_update::get_s(&update),
                    temporal_numeric_value_evm_update::get_v(&update),
                ),
                E_INVALID_SIGNATURE
            );

            temporal_numeric_value_feed_registry::update_latest_temporal_numeric_value(&mut feed_registry, temporal_numeric_value_evm_update::get_id(&update), temporal_numeric_value);
            num_updates = num_updates + 1;
        };

        primary_fungible_store::transfer(signer, AptosCoin, @stork, fee * num_updates);
    }

    #[view]
    /// Returns the latest temporal numeric value for an asset id
    public entry fun get_temporal_numeric_value_unchecked(
        /// The asset id
        asset_id: EncodedAssetId,
    ): TemporalNumericValue {
        let feed_registry = borrow_global<TemporalNumericValueFeedRegistry>(@stork);
        assert!(
            table::contains(&feed_registry.feed_table, asset_id),
            E_FEED_NOT_FOUND
        );
        temporal_numeric_value_feed_registry::get_latest_temporal_numeric_value_unchecked(&feed_registry, asset_id)
    }

    
}