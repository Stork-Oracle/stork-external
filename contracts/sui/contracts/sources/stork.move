module stork::stork {

    // === Imports ===

    use stork::admin::AdminCap;
    use stork::state::{Self, StorkState};
    use stork::temporal_numeric_value_feed;
    use stork::temporal_numeric_value::TemporalNumericValue;
    use stork::encoded_asset_id;
    use stork::verify::verify_stork_evm_signature;
    use sui::sui::SUI;
    use sui::coin::Coin;
    use stork::event::Self;

    // === Errors ===

    const EInvalidSignature: u64 = 0;
    const EInsufficientFee: u64 = 1;
    const EFeedNotFound: u64 = 2;

    // === Structs ===

    public struct UpdateTemporalNumericValueEvmInput has copy, drop, store {
        // the id of the asset to update
        id: vector<u8>,
        // the temporal numeric value to update
        temporal_numeric_value: TemporalNumericValue,
        // the publisher merkle root
        publisher_merkle_root: vector<u8>,
        // the value compute alg hash
        value_compute_alg_hash: vector<u8>,
        // the r value
        r: vector<u8>,
        // the s value
        s: vector<u8>,
        // the v value
        v: u8,
    }

    // === Functions ===

    entry fun init_stork(
        // admin capability
        _: &AdminCap,
        // the address of the Stork program
        stork_sui_public_key: address,
        // Storks EVM public key
        stork_evm_public_key: vector<u8>,
        // the fee to update a value
        single_update_fee: u64,
        // version of the Stork state 
        stork_state_version: u64,
        // context
        ctx: &mut TxContext,
    ) {
        let state = state::new(stork_sui_public_key, stork_evm_public_key, single_update_fee, stork_state_version, ctx);
        let state_id = object::id(&state);
        state::share(state);
        event::emit_stork_initialization_event(
            stork_sui_public_key,
            stork_evm_public_key,
            single_update_fee,
            state_id,
            stork_state_version
        );
    }

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
        let feed_id = encoded_asset_id::from_bytes(update_data.id);
        let evm_pubkey = stork_state.get_stork_evm_public_key();
        let fee_in_mist = stork_state.get_single_update_fee_in_mist();

        assert!(fee.value() >= fee_in_mist, EInsufficientFee);
        stork_state.deposit_fee(fee);

        let feed_registry = stork_state.borrow_tnv_feeds_registry_mut();

        if (feed_registry.contains(feed_id)) {
            let feed = feed_registry.borrow_mut(feed_id);
            if (feed.get_latest_canonical_temporal_numeric_value_unchecked().get_timestamp_ns() >= update_data.temporal_numeric_value.get_timestamp_ns()) {
                return
            }
        };

        assert!(verify_stork_evm_signature(
            &evm_pubkey,
            update_data.id,
            update_data.temporal_numeric_value.get_timestamp_ns(),
            update_data.temporal_numeric_value.get_quantized_value(),
            update_data.publisher_merkle_root,
            update_data.value_compute_alg_hash,
            update_data.r,
            update_data.s,
            update_data.v,
        ), EInvalidSignature);

        if (!feed_registry.contains(feed_id)) {
            let feed = temporal_numeric_value_feed::new(feed_id, update_data.temporal_numeric_value, ctx);
            // temporal_numeric_value_feed::share(feed);
            feed_registry.add(feed_id, feed);
        };

        let feed = feed_registry.borrow_mut(feed_id);
        feed.set_latest_value(update_data.temporal_numeric_value);
    }

    public fun update_multiple_temporal_numeric_values_evm(
        stork_state: &mut StorkState,
        // the input data
        update_data: vector<UpdateTemporalNumericValueEvmInput>,
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
            let update = update_data[i];
            let feed_id = encoded_asset_id::from_bytes(update.id);
            if (feed_registry.contains(feed_id)) {
                let feed = feed_registry.borrow_mut(feed_id);
                if (feed.get_latest_canonical_temporal_numeric_value_unchecked().get_timestamp_ns() >= update.temporal_numeric_value.get_timestamp_ns()) {
                    i = i + 1;
                    continue
                };
            };

            assert!(verify_stork_evm_signature(
                &evm_pubkey,
                update.id,
                update.temporal_numeric_value.get_timestamp_ns(),
                update.temporal_numeric_value.get_quantized_value(),
                update.publisher_merkle_root,
                update.value_compute_alg_hash,
                update.r,
                update.s,
                update.v,
            ), EInvalidSignature);

            // create tvn feed account if it doesn't exist
            if (!feed_registry.contains(feed_id)) {
                feed_registry.add(feed_id, temporal_numeric_value_feed::new(feed_id, update.temporal_numeric_value, ctx));
            };

            let feed = feed_registry.borrow_mut(feed_id);
            feed.set_latest_value(update.temporal_numeric_value);

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

}
