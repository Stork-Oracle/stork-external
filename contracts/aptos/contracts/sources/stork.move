module stork::stork {

    // === Imports ===

    use stork::state::{Self, StorkState};
    use stork::event::{emit_stork_initialization_event};
    use stork::encoded_asset_id::{Self, EncodedAssetId};
    use stork::temporal_numeric_value_feed_registry::{Self, TemporalNumericValueFeedRegistry};
    use stork::temporal_numeric_value_evm_update::{Self, TemporalNumericValueEVMUpdate};
    use stork::temporal_numeric_value::{Self, TemporalNumericValue};
    use stork::evm_pubkey::{Self, EvmPubKey};
    use stork::verify;
    use stork::i128;
    use aptos_std::table;
    use aptos_std::primary_fungible_store;
    use aptos_std::vector;
    use aptos_std::event;
    use aptos_std::signer;
    use aptos_framework::aptos_coin::AptosCoin;
    use aptos_framework::coin::{Self, Coin};

    // === Errors ===

    const E_NOT_OWNER: u64 = 0;
    const E_ALREADY_INITIALIZED: u64 = 1;
    const E_INVALID_SIGNATURE: u64 = 2;

    // === Functions ===

    entry fun init_stork(
        stork_evm_public_key: vector<u8>,
        single_update_fee: u64,
        owner: &signer,
    ) {
        assert!(
            !state::state_exists(),
            E_ALREADY_INITIALIZED
        );
        let evm_pubkey = evm_pubkey::from_bytes(stork_evm_public_key);
        let state = state::new(evm_pubkey, single_update_fee, signer::address_of(owner));        
        state.move_state(owner);

        // TNV feed table
        let feed_registry = temporal_numeric_value_feed_registry::new();
        feed_registry.move_tnv_feed_registry(owner);

        emit_stork_initialization_event(
            @stork,
            evm_pubkey,
            single_update_fee,
            signer::address_of(owner),
        );
    }

    // === Public Functions ===

    /// Updates a single temporal numeric value using EVM signature for verification
    public entry fun update_single_temporal_numeric_value_evm(
        /// The signer of the transaction to pay the fee
        signer: &signer,
        /// The asset id
        asset_id: vector<u8>,
        /// The temporal numeric value
        temporal_numeric_value_timestamp_ns: u64,
        /// The temporal numeric value
        temporal_numeric_value_quantized_value: u128,
        /// The publisher's merkle root
        publisher_merkle_root: vector<u8>,
        /// The value compute algorithm hash
        value_compute_alg_hash: vector<u8>,
        /// The signature r
        r: vector<u8>,
        /// The signature s
        s: vector<u8>,
        /// The signature v
        v: u8,
    ) {
        let evm_pubkey = state::get_stork_evm_public_key();
        let fee = state::get_single_update_fee();
        let encoded_asset_id = encoded_asset_id::from_bytes(asset_id);
        // recency
        if (temporal_numeric_value_feed_registry::contains(encoded_asset_id)) {
            let existing_temporal_numeric_value = temporal_numeric_value_feed_registry::get_latest_canonical_temporal_numeric_value_unchecked(encoded_asset_id);
            if (temporal_numeric_value_timestamp_ns <= existing_temporal_numeric_value.get_timestamp_ns()) {
                return;
            };
        };

        // verify signature
        assert!(
            verify::verify_evm_signature(
                &evm_pubkey,
                asset_id,
                temporal_numeric_value_timestamp_ns,
                i128::from_u128(temporal_numeric_value_quantized_value),
                publisher_merkle_root,
                value_compute_alg_hash,
                r,
                s,
                v,
            ),
            E_INVALID_SIGNATURE
        );

        transfer_fee(signer, fee);

        let temporal_numeric_value = temporal_numeric_value::new(temporal_numeric_value_timestamp_ns, i128::from_u128(temporal_numeric_value_quantized_value));
        temporal_numeric_value_feed_registry::update_latest_temporal_numeric_value(encoded_asset_id, temporal_numeric_value);
    }

    /// Updates multiple temporal numeric values using EVM signature for verification
    /// For each update, the position in the vectors corresponds to the position in the updates vector
    /// i.e ids[0] corresponds to timestamps_ns[0], quantized_values[0], etc.
    public entry fun update_temporal_numeric_values_evm(
        /// The signer of the transaction to pay the fee
        signer: &signer,
        /// The asset ids
        ids: vector<vector<u8>>,
        /// The timestamps in nanoseconds
        timestamps_ns: vector<u64>,
        /// The quantized values
        quantized_values: vector<u128>,
        /// The publisher's merkle roots
        publisher_merkle_roots: vector<vector<u8>>,
        /// The value compute algorithm hashes
        value_compute_alg_hashes: vector<vector<u8>>,
        /// The signatures r
        rs: vector<vector<u8>>,
        /// The signatures s
        ss: vector<vector<u8>>,
        /// The signatures v
        vs: vector<u8>,
    ) {
        let evm_pubkey = state::get_stork_evm_public_key();
        let fee = state::get_single_update_fee();
        let updates = temporal_numeric_value_evm_update::from_vectors(ids, timestamps_ns, quantized_values, publisher_merkle_roots, value_compute_alg_hashes, rs, ss, vs);

        let num_updates = 0;
        while (updates.length() > 0) {
            let update = updates.pop_back();
            // recency
            if (temporal_numeric_value_feed_registry::contains(update.get_id())) {
                let existing_temporal_numeric_value = temporal_numeric_value_feed_registry::get_latest_canonical_temporal_numeric_value_unchecked(update.get_id());
                if (update.get_temporal_numeric_value().get_timestamp_ns() <= existing_temporal_numeric_value.get_timestamp_ns()) {
                    continue;
                };
            };

            // verify signature
            assert!(
                verify::verify_evm_signature(
                    &evm_pubkey,
                    update.get_id().get_bytes(),
                    update.get_temporal_numeric_value().get_timestamp_ns(),
                    update.get_temporal_numeric_value().get_quantized_value(),
                    update.get_publisher_merkle_root(),
                    update.get_value_compute_alg_hash(),
                    update.get_r(),
                    update.get_s(),
                    update.get_v(),
                ),
                E_INVALID_SIGNATURE
            );

            num_updates = num_updates + 1;

            temporal_numeric_value_feed_registry::update_latest_temporal_numeric_value(update.get_id(), update.get_temporal_numeric_value());
        };

        transfer_fee(signer, fee * num_updates);
    }

    #[view]
    /// Returns the latest temporal numeric value for an asset id
    public fun get_temporal_numeric_value_unchecked(
        /// The asset id
        asset_id: vector<u8>,
    ): TemporalNumericValue {
        let encoded_asset_id = encoded_asset_id::from_bytes(asset_id);
        temporal_numeric_value_feed_registry::get_latest_canonical_temporal_numeric_value_unchecked(encoded_asset_id)
    }

    // === Private Functions ===

    fun transfer_fee(
        signer: &signer,
        fee: u64,
    ) {
        let coin = coin::withdraw<AptosCoin>(signer, fee);
        coin::deposit<AptosCoin>(@stork, coin);
    }

    
}