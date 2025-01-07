module stork::temporal_numeric_value_feed_registry {

    // === Imports ===

    use aptos_std::table;
    use aptos_std::event;
    use stork::temporal_numeric_value::TemporalNumericValue;
    use stork::encoded_asset_id::EncodedAssetId;
    use stork::event::{emit_temporal_numeric_value_update_event};

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
        let feed_registry = borrow_global<TemporalNumericValueFeedRegistry>(@stork);
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
        let feed_registry = borrow_global_mut<TemporalNumericValueFeedRegistry>(@stork);
        feed_registry.feed_table.upsert(asset_id, temporal_numeric_value);
        emit_temporal_numeric_value_update_event(asset_id, temporal_numeric_value);
    }

    package fun contains(
        asset_id: EncodedAssetId,
    ): bool acquires TemporalNumericValueFeedRegistry {
        let feed_registry = borrow_global<TemporalNumericValueFeedRegistry>(@stork);
        feed_registry.feed_table.contains(asset_id)
    }

}