module stork::temporal_numeric_value_feed {
     
     // === Imports ===

    use stork::temporal_numeric_value::TemporalNumericValue;
    use stork::encoded_asset_id::EncodedAssetId;

    // === Structs ===

    public struct TemporalNumericValueFeed has key, store {
        id: UID,
        asset_id: EncodedAssetId,
        latest_value: TemporalNumericValue,
    }

    // === Functions ===

    public(package) fun new(asset_id: EncodedAssetId, latest_value: TemporalNumericValue, ctx: &mut TxContext): TemporalNumericValueFeed {
        TemporalNumericValueFeed {
            id: object::new(ctx),
            asset_id,
            latest_value,
        }
    }
    
    // public(package) fun share(feed: TemporalNumericValueFeed) {
    //     transfer::share_object(feed)
    // }

    public fun get_latest_canonical_temporal_numeric_value_unchecked(feed: &TemporalNumericValueFeed): TemporalNumericValue {
        feed.latest_value
    }

    public fun get_asset_id(feed: &TemporalNumericValueFeed): EncodedAssetId {
        feed.asset_id
    }

    public(package) fun set_latest_value(feed: &mut TemporalNumericValueFeed, latest_value: TemporalNumericValue) {
        feed.latest_value = latest_value;
    }

    // === Test Imports ===

    #[test_only] use stork::encoded_asset_id;
    #[test_only] use stork::temporal_numeric_value;
    #[test_only] use stork::i128;

    // === Tests ===

    #[test]
    fun test_get_latest_canonical_temporal_numeric_value_unchecked() {
        let feed = create_zeroed_feed();
        let value = get_latest_canonical_temporal_numeric_value_unchecked(&feed);
        assert!(value.get_timestamp_ns() == 0);
        assert!(value.get_quantized_value() == i128::from_u128(0));
        destroy(feed)
    }

    #[test]
    fun test_get_asset_id() {
        let feed = create_zeroed_feed();
        let asset_id = get_asset_id(&feed);
        std::debug::print(&asset_id.get_bytes());
        assert!(asset_id.get_bytes() == b"00000000000000000000000000000000");
        destroy(feed)
    }

    // === Test Helpers ===

    #[test_only]
    fun create_zeroed_feed(): TemporalNumericValueFeed {
        let mut ctx = tx_context::dummy();
        new(encoded_asset_id::create_zeroed_asset_id(), temporal_numeric_value::create_zeroed_temporal_numeric_value(), &mut ctx)
    }

    #[test_only]
    fun destroy(feed: TemporalNumericValueFeed) {
        let TemporalNumericValueFeed { 
            id,
            asset_id: _,
            latest_value: _,
        } = feed;

        object::delete(id);
    }
}
