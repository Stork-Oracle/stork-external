module stork::stork {

    // === Imports ===

    use stork::admin::AdminCap;
    use stork::state::{Self, StorkState};
    use stork::temporal_numeric_value_feed::{Self, TemporalNumericValueFeed};
    use stork::temporal_numeric_value::TemporalNumericValue;
    use stork::encoded_asset_id::{Self, EncodedAssetId};
    use stork::verify::verify_stork_evm_signature;
    use stork::event::Self;
    use stork::update_temporal_numeric_value_evm_input::UpdateTemporalNumericValueEvmInput;
    use stork::update_temporal_numeric_value_evm_input_vec::UpdateTemporalNumericValueEvmInputVec;
    use sui::sui::SUI;
    use sui::coin::Coin;
    use sui::object_table::ObjectTable;

    // === Errors ===

    const EInvalidSignature: u64 = 0;
    const EInsufficientFee: u64 = 1;
    const EFeedNotFound: u64 = 2;

    // === Entry Functions ===

    entry fun init_stork(
        // admin capability
        _: &AdminCap,
        // the address of the Stork program
        stork_sui_address: address,
        // Storks EVM public key
        stork_evm_public_key: vector<u8>,
        // the fee to update a value
        single_update_fee: u64,
        // version of the Stork state
        stork_state_version: u64,
        // context
        ctx: &mut TxContext,
    ) {
        let state = state::new(stork_sui_address, stork_evm_public_key, single_update_fee, stork_state_version, ctx);
        let state_id = object::id(&state);
        state::share(state);
        event::emit_stork_initialization_event(
            stork_sui_address,
            stork_evm_public_key,
            single_update_fee,
            state_id,
            stork_state_version
        );
    }

    // === Public Functions ===

    // updates the price feed for an asset
    // input data is an assets update data signed with Storks EVM public key
    public fun update_single_temporal_numeric_value_evm(
        stork_state: &mut StorkState,
        // the input data
        update_data: UpdateTemporalNumericValueEvmInput,
        // fee
        fee: Coin<SUI>,
        // context
        ctx: &mut TxContext,
    ) {
        let feed_id = update_data.get_id();
        let evm_pubkey = stork_state.get_stork_evm_public_key();
        let fee_in_mist = stork_state.get_single_update_fee_in_mist();

        assert!(fee.value() >= fee_in_mist, EInsufficientFee);
        stork_state.deposit_fee(fee);

        let feed_registry = stork_state.borrow_tnv_feeds_registry_mut();

        if (feed_registry.contains(feed_id)) {
            let feed = feed_registry.borrow_mut(feed_id);
            if (feed.get_latest_canonical_temporal_numeric_value_unchecked().get_timestamp_ns() >= update_data.get_temporal_numeric_value().get_timestamp_ns()) {
                return
            }
        };

        assert!(verify_stork_evm_signature(
            &evm_pubkey,
            update_data.get_id().get_bytes(),
            update_data.get_temporal_numeric_value().get_timestamp_ns(),
            update_data.get_temporal_numeric_value().get_quantized_value(),
            update_data.get_publisher_merkle_root(),
            update_data.get_value_compute_alg_hash(),
            update_data.get_r(),
            update_data.get_s(),
            update_data.get_v(),
        ), EInvalidSignature);

        create_or_update_temporal_numeric_value_feed(feed_registry, feed_id, update_data, ctx);
    }

    // updates multiple price feeds for assets
    public fun update_multiple_temporal_numeric_values_evm(
        stork_state: &mut StorkState,
        // the input data
        update_data: UpdateTemporalNumericValueEvmInputVec,
        // fee
        fee: Coin<SUI>,
        // context
        ctx: &mut TxContext,
    ) {

        let evm_pubkey = stork_state.get_stork_evm_public_key();
        let feed_registry = stork_state.borrow_tnv_feeds_registry_mut();
        let mut num_updates = 0;

        let mut i = 0;
        while (i < update_data.length()) {
            let update = update_data.get_data()[i];
            let feed_id = update.get_id();
            if (feed_registry.contains(feed_id)) {
                let feed = feed_registry.borrow_mut(feed_id);
                if (feed.get_latest_canonical_temporal_numeric_value_unchecked().get_timestamp_ns() >= update.get_temporal_numeric_value().get_timestamp_ns()) {
                    i = i + 1;
                    continue
                };
            };

            assert!(verify_stork_evm_signature(
                &evm_pubkey,
                update.get_id().get_bytes(),
                update.get_temporal_numeric_value().get_timestamp_ns(),
                update.get_temporal_numeric_value().get_quantized_value(),
                update.get_publisher_merkle_root(),
                update.get_value_compute_alg_hash(),
                update.get_r(),
                update.get_s(),
                update.get_v(),
            ), EInvalidSignature);

            create_or_update_temporal_numeric_value_feed(feed_registry, feed_id, update, ctx);
            num_updates = num_updates + 1;
        };
        let required_fee = stork_state.get_total_fees_in_mist(num_updates);
        assert!(fee.value() >= required_fee, EInsufficientFee);
        stork_state.deposit_fee(fee);
    }

    // allows caller to get the latest canonical temporal numeric value from a feed id (encoded asset id) as a byte vec
    // unchecked because it does not impose staleness check
    public fun get_temporal_numeric_value_unchecked(stork_state: &StorkState, feed_id: vector<u8>): TemporalNumericValue {
        let feed_id = encoded_asset_id::from_bytes(feed_id);
        let feed_registry = stork_state.borrow_tnv_feeds_registry();
        assert!(feed_registry.contains(feed_id), EFeedNotFound);
        let feed = feed_registry.borrow(feed_id);
        feed.get_latest_canonical_temporal_numeric_value_unchecked()
    }

    // === Private Functions ===

    // updates feed if it exists, or creates it with update if it doesn't
    fun create_or_update_temporal_numeric_value_feed(
        feed_registry: &mut ObjectTable<EncodedAssetId, TemporalNumericValueFeed>,
        feed_id: EncodedAssetId,
        update_data: UpdateTemporalNumericValueEvmInput,
        ctx: &mut TxContext,
    ) {
         if (!feed_registry.contains(feed_id)) {
            let feed = temporal_numeric_value_feed::new(feed_id, update_data.get_temporal_numeric_value(), ctx);
            feed_registry.add(feed_id, feed);
        }
        else {
            let feed = feed_registry.borrow_mut(feed_id);
            feed.set_latest_value(update_data.get_temporal_numeric_value());
        };
        event::emit_temporal_numeric_value_feed_update_event(feed_id, update_data.get_temporal_numeric_value());
    }

    // === Test Helpers ===

    #[test_only]
    public fun init_stork_for_testing(
        // admin capability
        admin_cap: &AdminCap,
        // the address of the Stork program
        stork_sui_address: address,
        // Storks EVM public key
        stork_evm_public_key: vector<u8>,
        // the fee to update a value
        single_update_fee: u64,
        // version of the Stork state
        stork_state_version: u64,
        // context
        ctx: &mut TxContext,
    ) {
        init_stork(
            admin_cap,
            stork_sui_address,
            stork_evm_public_key,
            single_update_fee,
            stork_state_version,
            ctx
        );
    }
}
