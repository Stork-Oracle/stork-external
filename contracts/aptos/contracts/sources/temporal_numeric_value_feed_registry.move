module stork::temporal_numeric_value_feed_registry {

    // === Imports ===

    use aptos_std::table;
    use stork::temporal_numeric_value::TemporalNumericValue;
    use stork::encoded_asset_id::EncodedAssetId;
    use stork::event::{emit_temporal_numeric_value_update_event};
    use stork::state_account_store;

    // === Errors ===

    const E_FEED_NOT_FOUND: u64 = 0;

    // === Structs ===

    struct TemporalNumericValueFeedRegistry has key {
        feed_table: table::Table<EncodedAssetId, TemporalNumericValue>,
    }

    // === Functions ===

    package fun new(): TemporalNumericValueFeedRegistry {
        TemporalNumericValueFeedRegistry {
            feed_table: table::new(),
        }
    }

    package fun move_tnv_feed_registry(self: TemporalNumericValueFeedRegistry, owner: &signer) {
        move_to(owner, self);
    }

    package fun get_latest_canonical_temporal_numeric_value_unchecked(
        asset_id: EncodedAssetId,
    ): TemporalNumericValue acquires TemporalNumericValueFeedRegistry {
        let feed_registry = borrow_global<TemporalNumericValueFeedRegistry>(state_account_store::get_state_account_address());
        assert!(
            feed_registry.feed_table.contains(asset_id),
            E_FEED_NOT_FOUND
        );
        *table::borrow(&feed_registry.feed_table, asset_id)
    }

    package fun update_latest_temporal_numeric_value(
        asset_id: EncodedAssetId,
        temporal_numeric_value: TemporalNumericValue,
    ) acquires TemporalNumericValueFeedRegistry {
        let feed_registry = borrow_global_mut<TemporalNumericValueFeedRegistry>(state_account_store::get_state_account_address());
        feed_registry.feed_table.upsert(asset_id, temporal_numeric_value);
        emit_temporal_numeric_value_update_event(asset_id, temporal_numeric_value);
    }

    package fun contains(
        asset_id: EncodedAssetId,
    ): bool acquires TemporalNumericValueFeedRegistry {
        let feed_registry = borrow_global<TemporalNumericValueFeedRegistry>(state_account_store::get_state_account_address());
        feed_registry.feed_table.contains(asset_id)
    }

    // === Test Imports ===

    #[test_only]
    use aptos_framework::account::create_account_for_test;
    #[test_only]
    use stork::temporal_numeric_value;
    #[test_only]
    use stork::encoded_asset_id;

    // === Test Constants ===

    #[test_only]
    const DEPLOYER: address = @0xFACE;
    #[test_only]
    const STORK: address = @stork;

    // === Test Helpers ===

    #[test_only]
    fun setup_test(): signer {
        let stork_signer = create_account_for_test(STORK);
        state_account_store::init_module_for_test(&stork_signer);
        let deployer_signer = create_account_for_test(DEPLOYER);
        let stork_state_account_signer = state_account_store::get_state_account_signer();
        let registry = new();
        
        registry.move_tnv_feed_registry(&stork_state_account_signer);
        deployer_signer
    }

    // === Tests ===

    #[test]
    fun test_update_and_get_value() acquires TemporalNumericValueFeedRegistry {
        setup_test();

        let asset_id = encoded_asset_id::create_zeroed_asset_id();
        let value = temporal_numeric_value::create_zeroed_temporal_numeric_value();

        assert!(!contains(asset_id), 0);

        update_latest_temporal_numeric_value(asset_id, value);

        assert!(contains(asset_id), 1);
        let stored_value = get_latest_canonical_temporal_numeric_value_unchecked(asset_id);
        assert!(stored_value == value, 2);
    }

    #[test]
    fun test_multiple_updates() acquires TemporalNumericValueFeedRegistry {
        setup_test();

        let asset_id = encoded_asset_id::create_zeroed_asset_id();
        let value1 = temporal_numeric_value::create_zeroed_temporal_numeric_value();
        let value2 = temporal_numeric_value::create_zeroed_temporal_numeric_value();

        update_latest_temporal_numeric_value(asset_id, value1);
        update_latest_temporal_numeric_value(asset_id, value2);

        let stored_value = get_latest_canonical_temporal_numeric_value_unchecked(asset_id);
        assert!(stored_value == value2, 0);
    }

    #[test]
    #[expected_failure(abort_code = E_FEED_NOT_FOUND)]
    fun test_get_nonexistent_feed() acquires TemporalNumericValueFeedRegistry {
        setup_test();

        let asset_id = encoded_asset_id::create_zeroed_asset_id();
        get_latest_canonical_temporal_numeric_value_unchecked(asset_id);
    }

    #[test]
    fun test_multiple_assets() acquires TemporalNumericValueFeedRegistry {
        setup_test();

        let asset_id1 = encoded_asset_id::create_zeroed_asset_id();
        let asset_id2 = encoded_asset_id::create_zeroed_asset_id();
        let value1 = temporal_numeric_value::create_zeroed_temporal_numeric_value();
        let value2 = temporal_numeric_value::create_zeroed_temporal_numeric_value();

        update_latest_temporal_numeric_value(asset_id1, value1);
        update_latest_temporal_numeric_value(asset_id2, value2);

        assert!(get_latest_canonical_temporal_numeric_value_unchecked(asset_id1) == value1, 0);
        assert!(get_latest_canonical_temporal_numeric_value_unchecked(asset_id2) == value2, 1);
    }

}