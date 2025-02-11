module stork::stork {

    // === Imports ===

    use stork::state;
    use stork::event::{emit_stork_initialization_event};
    use stork::encoded_asset_id::{Self, EncodedAssetId};
    use stork::temporal_numeric_value_feed_registry;
    use stork::temporal_numeric_value::{Self, TemporalNumericValue};
    use stork::evm_pubkey;
    use stork::verify;
    use stork::i128;
    use stork::state_account_store;
    use aptos_std::signer;
    use aptos_framework::aptos_coin::AptosCoin;
    use aptos_framework::coin;

    // === Errors ===

    const E_ALREADY_INITIALIZED: u64 = 0;
    const E_INVALID_SIGNATURE: u64 = 1;
    const E_INVALID_LENGTHS: u64 = 2;
    const E_NO_UPDATES: u64 = 3;

    // === Functions ===



    entry fun init_stork(
        owner: &signer,
        stork_evm_public_key: vector<u8>,
        single_update_fee: u64,
    ) {
        assert!(
            !state::state_exists(),
            E_ALREADY_INITIALIZED
        );


        let state_account_signer = state_account_store::get_state_account_signer();
        coin::register<AptosCoin>(&state_account_signer);

        // Stork State resource
        let evm_pubkey = evm_pubkey::from_bytes(stork_evm_public_key);
        let state = state::new(evm_pubkey, single_update_fee, signer::address_of(owner));        
        state.move_state(&state_account_signer);

        // TNV feed table resource
        let feed_registry = temporal_numeric_value_feed_registry::new();
        feed_registry.move_tnv_feed_registry(&state_account_signer);

        emit_stork_initialization_event(
            @stork,
            evm_pubkey,
            single_update_fee,
            signer::address_of(owner),
            signer::address_of(&state_account_signer),
        );
    }

    // === Public Functions ===

    /// Updates a single temporal numeric value using EVM signature for verification
    public entry fun update_single_temporal_numeric_value_evm(
        // signer of the transaction to pay the fee
        signer: &signer,
        // asset id
        asset_id: vector<u8>,
        // temporal numeric value timestamp ns
        temporal_numeric_value_timestamp_ns: u64,
        // temporal numeric value magnitude
        temporal_numeric_value_magnitude: u128,
        // temporal numeric value negative
        temporal_numeric_value_negative: bool,
        // publisher's merkle root
        publisher_merkle_root: vector<u8>,
        // value compute algorithm hash
        value_compute_alg_hash: vector<u8>,
        // signature r
        r: vector<u8>,
        // signature s
        s: vector<u8>,
        // signature v
        v: u8,
    ) {
        let evm_pubkey = state::get_stork_evm_public_key();
        let fee = state::get_single_update_fee_in_octas();
        let encoded_asset_id = encoded_asset_id::from_bytes(asset_id);
        // recency
        if (!is_recent(encoded_asset_id, temporal_numeric_value_timestamp_ns)) {
            return;
        };

        // verify signature
        assert!(
            verify::verify_evm_signature(
                &evm_pubkey,
                asset_id,
                temporal_numeric_value_timestamp_ns,
                i128::new(temporal_numeric_value_magnitude, temporal_numeric_value_negative),
                publisher_merkle_root,
                value_compute_alg_hash,
                r,
                s,
                v,
            ),
            E_INVALID_SIGNATURE
        );

        transfer_fee(signer, fee);

        let temporal_numeric_value = temporal_numeric_value::new(temporal_numeric_value_timestamp_ns, i128::new(temporal_numeric_value_magnitude, temporal_numeric_value_negative));
        temporal_numeric_value_feed_registry::update_latest_temporal_numeric_value(encoded_asset_id, temporal_numeric_value);
    }

    /// Updates multiple temporal numeric values using EVM signature for verification
    /// For each update, the position in the vectors corresponds to the position in the updates vector
    /// i.e ids[0] corresponds to timestamps_ns[0], quantized_values[0], etc.
    public entry fun update_multiple_temporal_numeric_values_evm(
        // signer of the transaction to pay the fee
        signer: &signer,
        // asset ids
        ids: vector<vector<u8>>,
        // temporal numeric value timestamp ns
        temporal_numeric_value_timestamp_ns: vector<u64>,
        // temporal numeric value magnitude
        temporal_numeric_value_magnitude: vector<u128>,
        // temporal numeric value negative
        temporal_numeric_value_negative: vector<bool>,
        // publisher's merkle roots
        publisher_merkle_roots: vector<vector<u8>>,
        // value compute algorithm hashes
        value_compute_alg_hashes: vector<vector<u8>>,
        // signatures r
        rs: vector<vector<u8>>,
        // signatures s
        ss: vector<vector<u8>>,
        // signatures v
        vs: vector<u8>,
    ) {
        let evm_pubkey = state::get_stork_evm_public_key();
        let fee = state::get_single_update_fee_in_octas();

        assert!(ids.length() > 0, E_NO_UPDATES);
        assert!(temporal_numeric_value_timestamp_ns.length() == ids.length(), E_INVALID_LENGTHS);
        assert!(temporal_numeric_value_magnitude.length() == ids.length(), E_INVALID_LENGTHS);
        assert!(temporal_numeric_value_negative.length() == ids.length(), E_INVALID_LENGTHS);
        assert!(publisher_merkle_roots.length() == ids.length(), E_INVALID_LENGTHS);
        assert!(value_compute_alg_hashes.length() == ids.length(), E_INVALID_LENGTHS);
        assert!(rs.length() == ids.length(), E_INVALID_LENGTHS);
        assert!(ss.length() == ids.length(), E_INVALID_LENGTHS);
        assert!(vs.length() == ids.length(), E_INVALID_LENGTHS);


        let num_updates = 0;
        while (ids.length() > 0) {
            let id = ids.pop_back();
            let encoded_asset_id = encoded_asset_id::from_bytes(id);
            let timestamp_ns = temporal_numeric_value_timestamp_ns.pop_back();
            let magnitude = temporal_numeric_value_magnitude.pop_back();
            let negative = temporal_numeric_value_negative.pop_back();
            let publisher_merkle_root = publisher_merkle_roots.pop_back();
            let value_compute_alg_hash = value_compute_alg_hashes.pop_back();
            let r = rs.pop_back();
            let s = ss.pop_back();
            let v = vs.pop_back();

            // recency
            if (!is_recent(encoded_asset_id, timestamp_ns)) {
                continue;
            };

            // verify signature
            assert!(
                verify::verify_evm_signature(
                    &evm_pubkey,
                    id,
                    timestamp_ns,
                    i128::new(magnitude, negative),
                    publisher_merkle_root,
                    value_compute_alg_hash,
                    r,
                    s,
                    v,
                ),
                E_INVALID_SIGNATURE
            );

            num_updates = num_updates + 1;

            temporal_numeric_value_feed_registry::update_latest_temporal_numeric_value(encoded_asset_id, temporal_numeric_value::new(timestamp_ns, i128::new(magnitude, negative)));
        };

        transfer_fee(signer, fee * num_updates);
    }

    #[view]
    /// Returns the latest temporal numeric value for an asset id
    public fun get_temporal_numeric_value_unchecked(
        // The asset id
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
        coin::deposit<AptosCoin>(state_account_store::get_state_account_address(), coin);
    }

    fun is_recent(
        asset_id: EncodedAssetId,
        temporal_numeric_value_timestamp_ns: u64,
    ): bool {
        if (temporal_numeric_value_feed_registry::contains(asset_id)) {
            let existing_temporal_numeric_value = temporal_numeric_value_feed_registry::get_latest_canonical_temporal_numeric_value_unchecked(asset_id);
            temporal_numeric_value_timestamp_ns > existing_temporal_numeric_value.get_timestamp_ns()
        } else {
            true
        }
    }

    // === Test Imports ===

    #[test_only]
    use aptos_framework::account::{Self, create_account_for_test};
    #[test_only]
    use aptos_framework::aptos_coin;
    #[test_only]
    use std::vector;

    // === Test Constants ===

    #[test_only]
    const STORK: address = @stork;
    #[test_only]
    const DEPLOYER: address = @0xFACE;
    #[test_only]
    const USER: address = @0xCAFE;
    
    
    // === Test Helpers ===

    #[test_only]
    fun setup_test(): (signer, signer) {
        let stork_signer = create_account_for_test(STORK);
        state_account_store::init_module_for_test(&stork_signer);
        let deployer_signer = create_account_for_test(DEPLOYER);
        let pubkey = evm_pubkey::from_bytes(x"0a803F9b1CCe32e2773e0d2e98b37E0775cA5d44");
        let fee = 1;

        let user_signer = create_account_for_test(USER);
        // coin stores
        let framework_signer = account::create_account_for_test(@aptos_framework);
        let (burn_cap, mint_cap) = aptos_framework::aptos_coin::initialize_for_test(&framework_signer);
        coin::register<AptosCoin>(&stork_signer);
        coin::register<AptosCoin>(&deployer_signer);
        coin::register<AptosCoin>(&user_signer);

        aptos_coin::mint(&framework_signer, USER, 50);
        aptos_coin::mint(&framework_signer, DEPLOYER, 50);

        // clean up capabilities
        coin::destroy_burn_cap(burn_cap);
        coin::destroy_mint_cap(mint_cap);
        init_stork(&deployer_signer, evm_pubkey::get_bytes(&pubkey), fee);
        (deployer_signer, user_signer)
    }
    
    // === Tests ===

    #[test]
    fun test_init_stork() {
        setup_test();
        assert!(state::state_exists(), 0);
    }

    #[test]
    #[expected_failure(abort_code = E_ALREADY_INITIALIZED)]
    fun test_init_stork_already_initialized() {
        let (deployer_signer, _) = setup_test();
        init_stork(&deployer_signer, x"0a803F9b1CCe32e2773e0d2e98b37E0775cA5d44", 1);
    }

    #[test]
    fun test_update_single_temporal_numeric_value_evm() {
        // Setup accounts
        let (_, user_signer) = setup_test();
        
        let asset_id = x"7404e3d104ea7841c3d9e6fd20adfe99b4ad586bc08d8f3bd3afef894cf184de";
        let timestamp = 1722632569208762117;
        let magnitude = 62507457175499998000000;
        let negative = false;
        let merkle_root = x"e5ff773b0316059c04aa157898766731017610dcbeede7d7f169bfeaab7cc318";
        let alg_hash = x"9be7e9f9ed459417d96112a7467bd0b27575a2c7847195c68f805b70ce1795ba";
        let r = x"b9b3c9f80a355bd0cd6f609fff4a4b15fa4e3b4632adabb74c020f5bcd240741";
        let s = x"16fab526529ac795108d201832cff8c2d2b1c710da6711fe9f7ab288a7149758";
        let v = 28;

        update_single_temporal_numeric_value_evm(
            &user_signer,
            asset_id,
            timestamp,
            magnitude,
            negative,
            merkle_root,
            alg_hash,
            r,
            s,
            v
        );

        let stored_value = get_temporal_numeric_value_unchecked(asset_id);
        assert!(temporal_numeric_value::get_timestamp_ns(&stored_value) == timestamp, 0);
    }

    #[test]
    fun test_update_multiple_temporal_numeric_values_evm() {
        let (_, user_signer) = setup_test();
        
        let asset_id = x"7404e3d104ea7841c3d9e6fd20adfe99b4ad586bc08d8f3bd3afef894cf184de";
        let ids = vector::singleton(asset_id);
        vector::push_back(&mut ids, asset_id);
        
        let timestamps = vector::singleton(1722632569208762117u64);
        vector::push_back(&mut timestamps, 1722632569208762117u64);
        
        let magnitudes = vector::singleton(62507457175499998000000u128);
        vector::push_back(&mut magnitudes, 62507457175499998000000u128);
        
        let negatives = vector::singleton(false);
        vector::push_back(&mut negatives, false);
        
        let merkle_root = x"e5ff773b0316059c04aa157898766731017610dcbeede7d7f169bfeaab7cc318";
        let merkle_roots = vector::singleton(merkle_root);
        vector::push_back(&mut merkle_roots, merkle_root);
        
        let alg_hash = x"9be7e9f9ed459417d96112a7467bd0b27575a2c7847195c68f805b70ce1795ba";
        let alg_hashes = vector::singleton(alg_hash);
        vector::push_back(&mut alg_hashes, alg_hash);
        
        let r = x"b9b3c9f80a355bd0cd6f609fff4a4b15fa4e3b4632adabb74c020f5bcd240741";
        let rs = vector::singleton(r);
        vector::push_back(&mut rs, r);
        
        let s = x"16fab526529ac795108d201832cff8c2d2b1c710da6711fe9f7ab288a7149758";
        let ss = vector::singleton(s);
        vector::push_back(&mut ss, s);
        
        let vs = vector::singleton(28u8);
        vector::push_back(&mut vs, 28u8);

        update_multiple_temporal_numeric_values_evm(
            &user_signer,
            ids,
            timestamps,
            magnitudes,
            negatives,
            merkle_roots,
            alg_hashes,
            rs,
            ss,
            vs
        );

        let stored_value = get_temporal_numeric_value_unchecked(asset_id);
        assert!(temporal_numeric_value::get_timestamp_ns(&stored_value) == 1722632569208762117, 0);
    }

    #[test]
    fun test_fee_transfer() {
        let (_, user_signer) = setup_test();
        
        let fee = state::get_single_update_fee_in_octas();

        let initial_user_balance = coin::balance<AptosCoin>(USER);
        let initial_state_account_balance = coin::balance<AptosCoin>(state_account_store::get_state_account_address());

        transfer_fee(&user_signer, fee);

        assert!(coin::balance<AptosCoin>(USER) == initial_user_balance - fee, 0);
        assert!(coin::balance<AptosCoin>(state_account_store::get_state_account_address()) == initial_state_account_balance + fee, 1);
    }

    #[test]
    fun test_is_recent_no_existing_value() {
        setup_test();

        let asset_id = x"7404e3d104ea7841c3d9e6fd20adfe99b4ad586bc08d8f3bd3afef894cf184de";
        let encoded_asset_id = encoded_asset_id::from_bytes(asset_id);
        let timestamp = 1722632569208762117;

        assert!(is_recent(encoded_asset_id, timestamp), 0);
    }

    #[test]
    fun test_is_recent_with_existing_value() {
        let (_, user_signer) = setup_test();
        
        let asset_id = x"7404e3d104ea7841c3d9e6fd20adfe99b4ad586bc08d8f3bd3afef894cf184de";
        let timestamp = 1722632569208762117;
        let magnitude = 62507457175499998000000;
        let negative = false;
        let merkle_root = x"e5ff773b0316059c04aa157898766731017610dcbeede7d7f169bfeaab7cc318";
        let alg_hash = x"9be7e9f9ed459417d96112a7467bd0b27575a2c7847195c68f805b70ce1795ba";
        let r = x"b9b3c9f80a355bd0cd6f609fff4a4b15fa4e3b4632adabb74c020f5bcd240741";
        let s = x"16fab526529ac795108d201832cff8c2d2b1c710da6711fe9f7ab288a7149758";
        let v = 28;

        update_single_temporal_numeric_value_evm(
            &user_signer,
            asset_id,
            timestamp,
            magnitude,
            negative,
            merkle_root,
            alg_hash,
            r,
            s,
            v
        );

        let encoded_asset_id = encoded_asset_id::from_bytes(asset_id);

        assert!(is_recent(encoded_asset_id, timestamp + 1), 0);
        assert!(!is_recent(encoded_asset_id, timestamp), 1);
        assert!(!is_recent(encoded_asset_id, timestamp - 1), 2);
    }

    #[test]
    #[expected_failure(abort_code = E_INVALID_SIGNATURE)]
    fun test_update_single_temporal_numeric_value_evm_invalid_signature() {
        let (_, user_signer) = setup_test();
        
        let asset_id = x"7404e3d104ea7841c3d9e6fd20adfe99b4ad586bc08d8f3bd3afef894cf184de";
        let timestamp = 1722632569208762117;
        // Changed magnitude to trigger invalid signature
        let magnitude = 62507457175499998000001;
        let negative = false;
        let merkle_root = x"e5ff773b0316059c04aa157898766731017610dcbeede7d7f169bfeaab7cc318";
        let alg_hash = x"9be7e9f9ed459417d96112a7467bd0b27575a2c7847195c68f805b70ce1795ba";
        let r = x"b9b3c9f80a355bd0cd6f609fff4a4b15fa4e3b4632adabb74c020f5bcd240741";
        let s = x"16fab526529ac795108d201832cff8c2d2b1c710da6711fe9f7ab288a7149758";
        let v = 28;

        update_single_temporal_numeric_value_evm(
            &user_signer,
            asset_id,
            timestamp,
            magnitude,
            negative,
            merkle_root,
            alg_hash,
            r,
            s,
            v
        );
    }

    #[test]
    #[expected_failure(abort_code = E_INVALID_SIGNATURE)]
    fun test_update_multiple_temporal_numeric_values_evm_invalid_signature() {
        let (_, user_signer) = setup_test();
        
        let asset_id = x"7404e3d104ea7841c3d9e6fd20adfe99b4ad586bc08d8f3bd3afef894cf184de";
        let ids = vector::singleton(asset_id);
        vector::push_back(&mut ids, asset_id);
        
        let timestamps = vector::singleton(1722632569208762118u64);
        vector::push_back(&mut timestamps, 1722632569208762117u64);
        
        // Changed one magnitude to trigger invalid signature
        let magnitudes = vector::singleton(62507457175499998000001u128);
        vector::push_back(&mut magnitudes, 62507457175499998000000u128);
        
        let negatives = vector::singleton(false);
        vector::push_back(&mut negatives, false);
        
        let merkle_root = x"e5ff773b0316059c04aa157898766731017610dcbeede7d7f169bfeaab7cc318";
        let merkle_roots = vector::singleton(merkle_root);
        vector::push_back(&mut merkle_roots, merkle_root);
        
        let alg_hash = x"9be7e9f9ed459417d96112a7467bd0b27575a2c7847195c68f805b70ce1795ba";
        let alg_hashes = vector::singleton(alg_hash);
        vector::push_back(&mut alg_hashes, alg_hash);
        
        let r = x"b9b3c9f80a355bd0cd6f609fff4a4b15fa4e3b4632adabb74c020f5bcd240741";
        let rs = vector::singleton(r);
        vector::push_back(&mut rs, r);
        
        let s = x"16fab526529ac795108d201832cff8c2d2b1c710da6711fe9f7ab288a7149758";
        let ss = vector::singleton(s);
        vector::push_back(&mut ss, s);
        
        let vs = vector::singleton(28u8);
        vector::push_back(&mut vs, 28u8);

        update_multiple_temporal_numeric_values_evm(
            &user_signer,
            ids,
            timestamps,
            magnitudes,
            negatives,
            merkle_roots,
            alg_hashes,
            rs,
            ss,
            vs
        );
    }

    #[test]
    #[expected_failure(abort_code = E_NO_UPDATES)]
    fun test_update_multiple_temporal_numeric_values_evm_no_updates() {
        let (_, user_signer) = setup_test();
        
        let ids = vector::empty<vector<u8>>();
        let timestamps = vector::empty<u64>();
        let magnitudes = vector::empty<u128>();
        let negatives = vector::empty<bool>();
        let merkle_roots = vector::empty<vector<u8>>();
        let alg_hashes = vector::empty<vector<u8>>();
        let rs = vector::empty<vector<u8>>();
        let ss = vector::empty<vector<u8>>();
        let vs = vector::empty<u8>();

        update_multiple_temporal_numeric_values_evm(
            &user_signer,
            ids,
            timestamps,
            magnitudes,
            negatives,
            merkle_roots,
            alg_hashes,
            rs,
            ss,
            vs
        );
    }

    #[test]
    #[expected_failure(abort_code = E_INVALID_LENGTHS)]
    fun test_update_multiple_temporal_numeric_values_evm_invalid_lengths() {
        let (_, user_signer) = setup_test();
        
        let asset_id = x"7404e3d104ea7841c3d9e6fd20adfe99b4ad586bc08d8f3bd3afef894cf184de";
        let ids = vector::singleton(asset_id);
        vector::push_back(&mut ids, asset_id);
        
        // Only add one timestamp while ids has two elements
        let timestamps = vector::singleton(1722632569208762117u64);
        
        let magnitudes = vector::singleton(62507457175499998000000u128);
        vector::push_back(&mut magnitudes, 62507457175499998000000u128);
        
        let negatives = vector::singleton(false);
        vector::push_back(&mut negatives, false);
        
        let merkle_root = x"e5ff773b0316059c04aa157898766731017610dcbeede7d7f169bfeaab7cc318";
        let merkle_roots = vector::singleton(merkle_root);
        vector::push_back(&mut merkle_roots, merkle_root);
        
        let alg_hash = x"9be7e9f9ed459417d96112a7467bd0b27575a2c7847195c68f805b70ce1795ba";
        let alg_hashes = vector::singleton(alg_hash);
        vector::push_back(&mut alg_hashes, alg_hash);
        
        let r = x"b9b3c9f80a355bd0cd6f609fff4a4b15fa4e3b4632adabb74c020f5bcd240741";
        let rs = vector::singleton(r);
        vector::push_back(&mut rs, r);
        
        let s = x"16fab526529ac795108d201832cff8c2d2b1c710da6711fe9f7ab288a7149758";
        let ss = vector::singleton(s);
        vector::push_back(&mut ss, s);
        
        let vs = vector::singleton(28u8);
        vector::push_back(&mut vs, 28u8);

        update_multiple_temporal_numeric_values_evm(
            &user_signer,
            ids,
            timestamps,
            magnitudes,
            negatives,
            merkle_roots,
            alg_hashes,
            rs,
            ss,
            vs
        );
    }
}