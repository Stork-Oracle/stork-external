module stork::state {

    // === Imports ===

    use stork::admin::AdminCap;
    use stork::evm_pubkey::{EvmPubkey, Self};
    use stork::encoded_asset_id::EncodedAssetId;
    use sui::dynamic_object_field;
    use sui::object_table::{Self, ObjectTable};
    use stork::temporal_numeric_value_feed::TemporalNumericValueFeed;
    use sui::coin::{Self, Coin};
    use sui::sui::SUI;

    // === Errors ===

    const EIncorrectVersion: u64 = 0;
    const ENoFeesToWithdraw: u64 = 1;
    // === Constants ===

    const VERSION: u64 = 1;
    const TNV_FEEDS_REGISTRY_NAME: vector<u8> = b"temporal_numeric_value_feed_registry";
    const TREASURY_NAME: vector<u8> = b"treasury";

    // === Structs ===

    public struct StorkState has key {
        id: UID,
        // the address of the Stork program
        stork_sui_public_key: address,
        // Storks EVM public key
        stork_evm_public_key: EvmPubkey,
        // the fee to update a value
        single_update_fee_in_mist: u64,
        // version of the Stork state
        version: u64,
    }

    // === Functions ===
    
    public(package) fun new(
        stork_sui_public_key: address,
        stork_evm_public_key: vector<u8>,
        single_update_fee_in_mist: u64,
        version: u64,
        ctx: &mut TxContext,
    ): StorkState {
        assert!(version == VERSION, EIncorrectVersion);

        let evm_pubkey = evm_pubkey::from_bytes(stork_evm_public_key);

        let mut uid: UID = object::new(ctx);

        // Create a table that associates asset ids with tnv feeds
        // attach to the stork state object as a dynamic field
        dynamic_object_field::add(
            &mut uid,
            TNV_FEEDS_REGISTRY_NAME,
            object_table::new<EncodedAssetId, TemporalNumericValueFeed>(ctx),
        );

        // Create treasury object
        dynamic_object_field::add(
            &mut uid,
            TREASURY_NAME,
            coin::zero<SUI>(ctx),
        );

        StorkState {
            id: uid,
            stork_sui_public_key,
            stork_evm_public_key: evm_pubkey,
            single_update_fee_in_mist,
            version,
        }
    }

    public(package) fun share(stork_state: StorkState) {
        transfer::share_object(stork_state)
    }

    public(package) fun borrow_tnv_feeds_registry_mut(stork_state: &mut StorkState): &mut ObjectTable<EncodedAssetId, TemporalNumericValueFeed> {
        assert!(stork_state.version == VERSION, EIncorrectVersion);
        dynamic_object_field::borrow_mut(&mut stork_state.id, TNV_FEEDS_REGISTRY_NAME)
    }

    public(package) fun borrow_tnv_feeds_registry(stork_state: &StorkState): &ObjectTable<EncodedAssetId, TemporalNumericValueFeed> {
        assert!(stork_state.version == VERSION, EIncorrectVersion);
        dynamic_object_field::borrow(&stork_state.id, TNV_FEEDS_REGISTRY_NAME)
    }

    public fun get_stork_evm_public_key(stork_state: &StorkState): EvmPubkey {
        assert!(stork_state.version == VERSION, EIncorrectVersion);
        stork_state.stork_evm_public_key
    }

    public fun get_single_update_fee_in_mist(stork_state: &StorkState): u64 {
        assert!(stork_state.version == VERSION, EIncorrectVersion);
        stork_state.single_update_fee_in_mist
    }

    public fun get_stork_sui_public_key(stork_state: &StorkState): address {
        assert!(stork_state.version == VERSION, EIncorrectVersion);
        stork_state.stork_sui_public_key
    }

    public fun get_version(stork_state: &StorkState): u64 {
        assert!(stork_state.version == VERSION, EIncorrectVersion);
        stork_state.version
    }

    public fun deposit_fee(stork_state: &mut StorkState, fee: Coin<SUI>) {
        assert!(stork_state.version == VERSION, EIncorrectVersion);
        let treasury = dynamic_object_field::borrow_mut(
            &mut stork_state.id,
            TREASURY_NAME,
        );
        coin::join(treasury, fee)
    }

    public fun get_total_fees_in_mist(stork_state: &StorkState, num_updates: u64): u64 {
        assert!(stork_state.version == VERSION, EIncorrectVersion);
        stork_state.get_single_update_fee_in_mist() * num_updates
    }

    // === Admin Functions ===
    entry 
    public fun update_single_update_fee_in_mist(
        _: &AdminCap,
        state: &mut StorkState,
        new_single_update_fee_in_mist: u64,
    ) {
        assert!(state.version == VERSION, EIncorrectVersion);
        state.single_update_fee_in_mist = new_single_update_fee_in_mist;
    }

    public fun update_stork_sui_public_key(
        _: &AdminCap,
        state: &mut StorkState,
        new_stork_sui_public_key: address,
    ) {
        assert!(state.version == VERSION, EIncorrectVersion);
        state.stork_sui_public_key = new_stork_sui_public_key;
    }

    public fun update_stork_evm_public_key(
        _: &AdminCap,
        state: &mut StorkState,
        new_stork_evm_public_key: vector<u8>,
    ) {
        assert!(state.version == VERSION, EIncorrectVersion);
        state.stork_evm_public_key = evm_pubkey::from_bytes(new_stork_evm_public_key);  
    }

    public fun withdraw_fees(
        _: &AdminCap,
        state: &mut StorkState,
        ctx: &mut TxContext,
    ): Coin<SUI> {
        assert!(state.version == VERSION, EIncorrectVersion);
        let treasury = dynamic_object_field::borrow_mut<vector<u8>, Coin<SUI>>(
            &mut state.id,
            TREASURY_NAME,
        );
        assert!(treasury.value() > 0, ENoFeesToWithdraw);
        let treasury_value = treasury.value();
        coin::split(treasury, treasury_value, ctx)
    }

    entry fun migrate(
        _: &AdminCap,
        state: &mut StorkState,
        version: u64,
        stork_sui_public_key: address,
    ) {
        assert!(state.version == VERSION - 1, EIncorrectVersion);
        state.stork_sui_public_key = stork_sui_public_key;
        state.version = version;
    }

}
