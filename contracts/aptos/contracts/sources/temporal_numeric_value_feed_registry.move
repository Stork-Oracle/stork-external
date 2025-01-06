module stork::temporal_numeric_value_feed_registry {

    // === Imports ===

    use aptos_std::table;
    use stork::temporal_numeric_value::TemporalNumericValue;
    use stork::encoded_asset_id::EncodedAssetId;

    // === Structs ===

    struct TemporalNumericValueFeedRegistry {
        feed_table: table::Table<EncodedAssetId, TemporalNumericValue>,
    }

    // === Functions ===

    package fun new(): TemporalNumericValueFeedRegistry {
        TemporalNumericValueFeedRegistry {
            feed_table: table::new(),
        }
    }

    // === Public Functions ===

    package fun get_latest_temporal_numeric_value_unchecked(
        feed_registry: &TemporalNumericValueFeedRegistry,
        asset_id: EncodedAssetId,
    ): TemporalNumericValue {
        table::borrow(&feed_registry.feed_table, asset_id)
    }

    package fun update_latest_temporal_numeric_value(
        feed_registry: &mut TemporalNumericValueFeedRegistry,
        asset_id: EncodedAssetId,
        temporal_numeric_value: TemporalNumericValue,
    ) {
        table::upsert(&mut feed_registry.feed_table, asset_id, temporal_numeric_value);
        let event = event::TemporalNumericValueUpdateEvent { 
            asset_id,
            temporal_numeric_value,
        };
        event::emit(&event);
    }
}