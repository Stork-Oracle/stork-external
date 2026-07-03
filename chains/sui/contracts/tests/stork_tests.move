#[test_only]
module stork::stork_tests {

    // === Imports ===

    use stork::stork;
    use sui::test_scenario::Self;
    use stork::admin::{Self, AdminCap};
    use stork::state::{Self, StorkState};
    use sui::coin::Self;
    use sui::sui::SUI;
    use stork::update_temporal_numeric_value_evm_input;
    use stork::update_temporal_numeric_value_evm_input_vec;
    use std::unit_test::Self;
    use stork::i128;
    use stork::evm_pubkey;
    
    // === Constants ===

    const DEPLOYER: address = @0x26;
    const STORK_SUI_PUBLIC_KEY: address = @0x42;
    const SINGLE_UPDATE_FEE: u64 = 1000;
    const VERSION: u64 = 3;
    const STORK_EVM_PUBLIC_KEY: vector<u8> = x"0a803F9b1CCe32e2773e0d2e98b37E0775cA5d44";

    // Constants for verify.move test cases
    const VALID_ID: vector<u8> = x"7404e3d104ea7841c3d9e6fd20adfe99b4ad586bc08d8f3bd3afef894cf184de";
    const VALID_RECV_TIME: u64 = 1722632569208762117;
    const VALID_MERKLE_ROOT: vector<u8> = x"e5ff773b0316059c04aa157898766731017610dcbeede7d7f169bfeaab7cc318";
    const VALID_ALG_HASH: vector<u8> = x"9be7e9f9ed459417d96112a7467bd0b27575a2c7847195c68f805b70ce1795ba";
    const VALID_R: vector<u8> = x"b9b3c9f80a355bd0cd6f609fff4a4b15fa4e3b4632adabb74c020f5bcd240741";
    const VALID_S: vector<u8> = x"16fab526529ac795108d201832cff8c2d2b1c710da6711fe9f7ab288a7149758";
    const VALID_V: u8 = 28;

    // Second valid update signed with STORK_EVM_PUBLIC_KEY — distinct feed id from VALID_*.
    // Used to exercise batches with multiple distinct fresh updates.
    const SECOND_ID: vector<u8> = x"59102b37de83bdda9f38ac8254e596f0d9ac61d2035c07936675e87342817160";
    const SECOND_RECV_TIME: u64 = 1782760998091044836;
    const SECOND_VALUE: u128 = 1622121730843749000000;
    const SECOND_MERKLE_ROOT: vector<u8> = x"d01cd1e9214bf9cec434e82be8f995b9221864fcaab1012841407260049963da";
    const SECOND_R: vector<u8> = x"58245f6c5536f47c7c95b29f88ff5ceecc871378aeb778cd2c53f69bdf70e383";
    const SECOND_S: vector<u8> = x"14ae74028a55509bf8aa6631a8e1b171ecafdddddbcd5d3915c0d2bc1ac9287f";
    const SECOND_V: u8 = 27;

    // Constants for verify.mov negative value test case
    const NEGATIVE_ID: vector<u8> = x"281a649a11eb25eca04f0025c15e99264a056229e722735c7d6c55fef649dfbf";
    const NEGATIVE_RECV_TIME: u64 = 1750794968021348308;
    const NEGATIVE_MERKLE_ROOT: vector<u8> = x"5ea4136e8064520a3311961f3f7030dfbc0b96652f46a473e79f2a019b3cd878";
    const NEGATIVE_ALG_HASH: vector<u8> = x"9be7e9f9ed459417d96112a7467bd0b27575a2c7847195c68f805b70ce1795ba";
    const NEGATIVE_R: vector<u8> = x"14c36cf7272689cec0335efdc5f82dc2d4b1aceb8d2320d3245e4593df32e696";
    const NEGATIVE_S: vector<u8> = x"79ab437ecd56dc9fcf850f192328840f7f47d5df57cb939d99146b33014c39f0";
    const NEGATIVE_V: u8 = 27;
    // === Tests ===

    #[test]
    fun test_admin_init() {
        let deployer = @0x26;

        // Start the test scenario
        let mut scenario = test_scenario::begin(deployer);
        {
            // Initialize admin in the first transaction
            admin::test_init(test_scenario::ctx(&mut scenario));
        };

        // Move to next transaction to check if AdminCap was properly transferred
        test_scenario::next_tx(&mut scenario, deployer);
        {
            // Try to retrieve AdminCap from sender's inventory
            let admin_cap = test_scenario::take_from_sender<AdminCap>(&scenario);
            // Return the AdminCap back to sender's inventory
            test_scenario::return_to_sender(&scenario, admin_cap);
        };

        test_scenario::end(scenario);
    }

    #[test]
    fun test_transfer_admin_cap() {
        let deployer = @0x26;
        let new_admin = @0x27;

        let mut scenario = test_scenario::begin(deployer);
        {
            admin::test_init(test_scenario::ctx(&mut scenario));
        };

        test_scenario::next_tx(&mut scenario, deployer);
        {
            let admin_cap = test_scenario::take_from_sender<AdminCap>(&scenario);
            admin::transfer_admin_cap(admin_cap, new_admin);
        };

        // Original deployer must no longer hold the cap.
        test_scenario::next_tx(&mut scenario, deployer);
        {
            assert!(!test_scenario::has_most_recent_for_sender<AdminCap>(&scenario), 0);
        };

        // New admin must now hold the cap.
        test_scenario::next_tx(&mut scenario, new_admin);
        {
            let admin_cap = test_scenario::take_from_sender<AdminCap>(&scenario);
            test_scenario::return_to_sender(&scenario, admin_cap);
        };

        test_scenario::end(scenario);
    }

    #[test]
    fun test_stork_init() {
        let mut scenario = test_scenario::begin(DEPLOYER);

        {
            admin::test_init(test_scenario::ctx(&mut scenario));
        };  

        test_scenario::next_tx(&mut scenario, DEPLOYER);
        {
            let admin_cap = test_scenario::take_from_sender<AdminCap>(&scenario);
            stork::init_stork(
                &admin_cap,
                STORK_SUI_PUBLIC_KEY,
                STORK_EVM_PUBLIC_KEY,
                SINGLE_UPDATE_FEE,
                VERSION,
                test_scenario::ctx(&mut scenario),
            );
            test_scenario::return_to_sender(&scenario, admin_cap);
        };

        test_scenario::next_tx(&mut scenario, DEPLOYER);
        {
            let state = test_scenario::take_shared<state::StorkState>(&scenario);
            assert!(state.get_stork_sui_address() == STORK_SUI_PUBLIC_KEY);
            assert!(state.get_stork_evm_public_key().get_bytes() == STORK_EVM_PUBLIC_KEY);
            assert!(state.get_single_update_fee_in_mist() == SINGLE_UPDATE_FEE);
            assert!(state.get_version() == VERSION);
            test_scenario::return_shared(state);
        };
        test_scenario::end(scenario);
    }

    #[test]
    fun test_all_admin_functions() {
        let mut scenario = test_scenario::begin(DEPLOYER);
        
        // Init admin and stork
        {
            admin::test_init(test_scenario::ctx(&mut scenario));
        };
        test_scenario::next_tx(&mut scenario, DEPLOYER);
        {
            let admin_cap = test_scenario::take_from_sender<AdminCap>(&scenario);
            stork::init_stork(
                &admin_cap,
                STORK_SUI_PUBLIC_KEY,
                STORK_EVM_PUBLIC_KEY,
                SINGLE_UPDATE_FEE,
                VERSION,
                test_scenario::ctx(&mut scenario)
            );
            test_scenario::return_to_sender(&scenario, admin_cap);
        };

        // Test all admin functions
        test_scenario::next_tx(&mut scenario, DEPLOYER);
        {
            let admin_cap = test_scenario::take_from_sender<AdminCap>(&scenario);
            let mut state = test_scenario::take_shared<StorkState>(&scenario);

            // Test update fee
            state::update_single_update_fee_in_mist(&admin_cap, &mut state, 2000);
            assert!(state.get_single_update_fee_in_mist() == 2000, 0);

            // Test update SUI public key
            let new_sui_key = @0x99;
            state::update_stork_sui_address(&admin_cap, &mut state, new_sui_key);
            assert!(state.get_stork_sui_address() == new_sui_key, 0);

            // Test update EVM public key
            let new_evm_key = x"1111111111111111111111111111111111111111";
            state::update_stork_evm_public_key(&admin_cap, &mut state, new_evm_key);
            assert!(state.get_stork_evm_public_key().get_bytes() == new_evm_key, 0);

            test_scenario::return_shared(state);
            test_scenario::return_to_sender(&scenario, admin_cap);
        };

        test_scenario::end(scenario);
    }

    #[test]
    fun test_feed_operations() {
        let mut scenario = test_scenario::begin(DEPLOYER);
        
        // Setup initial state
        {
            admin::test_init(test_scenario::ctx(&mut scenario));
        };
        test_scenario::next_tx(&mut scenario, DEPLOYER);
        {
            let admin_cap = test_scenario::take_from_sender<AdminCap>(&scenario);
            stork::init_stork(
                &admin_cap,
                STORK_SUI_PUBLIC_KEY,
                STORK_EVM_PUBLIC_KEY,
                SINGLE_UPDATE_FEE,
                VERSION,
                test_scenario::ctx(&mut scenario)
            );
            test_scenario::return_to_sender(&scenario, admin_cap);
        };

        // Test valid feed update
        test_scenario::next_tx(&mut scenario, DEPLOYER);
        {
            let mut state = test_scenario::take_shared<StorkState>(&scenario);
            let update = update_temporal_numeric_value_evm_input::new(
                VALID_ID,
                VALID_RECV_TIME,
                62507457175499998000000,
                false,
                VALID_MERKLE_ROOT,
                VALID_ALG_HASH,
                VALID_R,
                VALID_S,
                VALID_V
            );
            
            let fee = coin::mint_for_testing<SUI>(2000, test_scenario::ctx(&mut scenario));
            
            stork::update_single_temporal_numeric_value_evm(
                &mut state,
                update,
                fee,
                test_scenario::ctx(&mut scenario)
            );

            // Verify feed was created and updated
            let feed_value = stork::get_temporal_numeric_value_unchecked(&state, VALID_ID);
            assert!(feed_value.get_timestamp_ns() == VALID_RECV_TIME, 0);
            assert!(feed_value.get_quantized_value() == i128::from_u128(62507457175499998000000), 0);

            test_scenario::return_shared(state);
        };

        test_scenario::end(scenario);
    }

    #[test]
    fun test_feed_operations_negative_value() {
        let mut scenario = test_scenario::begin(DEPLOYER);
        
        // Setup initial state
        {
            admin::test_init(test_scenario::ctx(&mut scenario));
        };
        test_scenario::next_tx(&mut scenario, DEPLOYER);
        {
            let admin_cap = test_scenario::take_from_sender<AdminCap>(&scenario);
            stork::init_stork(
                &admin_cap,
                STORK_SUI_PUBLIC_KEY,
                x"3db9E960ECfCcb11969509FAB000c0c96DC51830",
                SINGLE_UPDATE_FEE,
                VERSION,
                test_scenario::ctx(&mut scenario)
            );
            test_scenario::return_to_sender(&scenario, admin_cap);
        };

        // Test valid feed update
        test_scenario::next_tx(&mut scenario, DEPLOYER);
        {
            let mut state = test_scenario::take_shared<StorkState>(&scenario);
            let update = update_temporal_numeric_value_evm_input::new(
                NEGATIVE_ID,
                NEGATIVE_RECV_TIME,
                3020199000000,
                true,
                NEGATIVE_MERKLE_ROOT,
                NEGATIVE_ALG_HASH,
                NEGATIVE_R,
                NEGATIVE_S,
                NEGATIVE_V
            );

            let fee = coin::mint_for_testing<SUI>(2000, test_scenario::ctx(&mut scenario));
            
            stork::update_single_temporal_numeric_value_evm(
                &mut state,
                update,
                fee,
                test_scenario::ctx(&mut scenario)  
            );

            // Verify feed was created and updated
            let feed_value = stork::get_temporal_numeric_value_unchecked(&state, NEGATIVE_ID);
            assert!(feed_value.get_timestamp_ns() == NEGATIVE_RECV_TIME, 0);
            assert!(feed_value.get_quantized_value() == i128::new(3020199000000, true), 0);
            assert!(feed_value.get_quantized_value().is_negative(), 0);

            test_scenario::return_shared(state);
        };

        test_scenario::end(scenario);
    }

    #[test]
    #[expected_failure]
    fun test_feed_operations_invalid_signature() {
        let mut scenario = test_scenario::begin(DEPLOYER);
        
        // Setup initial state
        {
            admin::test_init(test_scenario::ctx(&mut scenario));
        };
        test_scenario::next_tx(&mut scenario, DEPLOYER);
        {
            let admin_cap = test_scenario::take_from_sender<AdminCap>(&scenario);
            stork::init_stork(
                &admin_cap,
                STORK_SUI_PUBLIC_KEY,
                STORK_EVM_PUBLIC_KEY,
                SINGLE_UPDATE_FEE,
                VERSION,
                test_scenario::ctx(&mut scenario)
            );
            test_scenario::return_to_sender(&scenario, admin_cap);
        };

        // Try update with invalid signature
        test_scenario::next_tx(&mut scenario, DEPLOYER);
        {
            let mut state = test_scenario::take_shared<StorkState>(&scenario);
            let update = update_temporal_numeric_value_evm_input::new(
                VALID_ID,
                VALID_RECV_TIME,
                62507457175499998000000,
                false,
                VALID_MERKLE_ROOT,
                VALID_ALG_HASH,
                // Invalid signature components
                x"0000000000000000000000000000000000000000000000000000000000000000",
                x"0000000000000000000000000000000000000000000000000000000000000000",
                0
            );
            
            let fee = coin::mint_for_testing<SUI>(2000, test_scenario::ctx(&mut scenario));
            
            stork::update_single_temporal_numeric_value_evm(
                &mut state,
                update,
                fee,
                test_scenario::ctx(&mut scenario)
            );
            
            test_scenario::return_shared(state);
        };

        test_scenario::end(scenario);
    }

    #[test]
    fun test_multiple_feed_updates() {
        let mut scenario = test_scenario::begin(DEPLOYER);
        
        // Setup initial state
        {
            admin::test_init(test_scenario::ctx(&mut scenario));
        };
        test_scenario::next_tx(&mut scenario, DEPLOYER);
        {
            let admin_cap = test_scenario::take_from_sender<AdminCap>(&scenario);
            stork::init_stork(
                &admin_cap,
                STORK_SUI_PUBLIC_KEY,
                STORK_EVM_PUBLIC_KEY,
                SINGLE_UPDATE_FEE,
                VERSION,
                test_scenario::ctx(&mut scenario)
            );
            test_scenario::return_to_sender(&scenario, admin_cap);
        };

        // Create vector input with single valid update
        test_scenario::next_tx(&mut scenario, DEPLOYER);
        {
            let mut state = test_scenario::take_shared<StorkState>(&scenario);
            
            // Create vectors with single element each
            let mut ids = vector[];
            vector::push_back(&mut ids, VALID_ID);
            
            let mut timestamps = vector[];
            vector::push_back(&mut timestamps, VALID_RECV_TIME);
            
            let mut magnitudes = vector[];
            vector::push_back(&mut magnitudes, 62507457175499998000000);
            
            let mut negatives = vector[];
            vector::push_back(&mut negatives, false);
            
            let mut merkle_roots = vector[];
            vector::push_back(&mut merkle_roots, VALID_MERKLE_ROOT);
            
            let mut alg_hashes = vector[];
            vector::push_back(&mut alg_hashes, VALID_ALG_HASH);
            
            let mut rs = vector[];
            vector::push_back(&mut rs, VALID_R);
            
            let mut ss = vector[];
            vector::push_back(&mut ss, VALID_S);
            
            let mut vs = vector[];
            vector::push_back(&mut vs, VALID_V);

            let updates = update_temporal_numeric_value_evm_input_vec::new(
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
            
            let fee = coin::mint_for_testing<SUI>(2000, test_scenario::ctx(&mut scenario));
            
            stork::update_multiple_temporal_numeric_values_evm(
                &mut state,
                updates,
                fee,
                test_scenario::ctx(&mut scenario)
            );

            // Verify feed was created and updated
            let feed_value = stork::get_temporal_numeric_value_unchecked(&state, VALID_ID);
            assert!(feed_value.get_timestamp_ns() == VALID_RECV_TIME, 0);
            
            test_scenario::return_shared(state);
        };

        test_scenario::end(scenario);
    }

    // Exercises a batch with two distinct fresh updates. Both must end up applied
    // and num_updates must equal 2 (verified indirectly via the exact fee coin size:
    // an off-by-one would cause an EInsufficientFee abort).
    //
    // Note: this passes on the pre-fix loop too (the self-stale-skip on the next
    // iteration still advances `i`), so it does not demonstrate a correctness bug —
    // it's a regression guard for the multi-distinct-fresh-update path, which had
    // no coverage before, and confirms the `i = i + 1` change preserves semantics.
    #[test]
    fun test_batch_multiple_distinct_fresh_updates() {
        let mut scenario = test_scenario::begin(DEPLOYER);

        {
            admin::test_init(test_scenario::ctx(&mut scenario));
        };
        test_scenario::next_tx(&mut scenario, DEPLOYER);
        {
            let admin_cap = test_scenario::take_from_sender<AdminCap>(&scenario);
            stork::init_stork(
                &admin_cap,
                STORK_SUI_PUBLIC_KEY,
                STORK_EVM_PUBLIC_KEY,
                SINGLE_UPDATE_FEE,
                VERSION,
                test_scenario::ctx(&mut scenario)
            );
            test_scenario::return_to_sender(&scenario, admin_cap);
        };

        test_scenario::next_tx(&mut scenario, DEPLOYER);
        {
            let mut state = test_scenario::take_shared<StorkState>(&scenario);

            let mut ids = vector[];
            vector::push_back(&mut ids, VALID_ID);
            vector::push_back(&mut ids, SECOND_ID);

            let mut timestamps = vector[];
            vector::push_back(&mut timestamps, VALID_RECV_TIME);
            vector::push_back(&mut timestamps, SECOND_RECV_TIME);

            let mut magnitudes = vector[];
            vector::push_back(&mut magnitudes, 62507457175499998000000);
            vector::push_back(&mut magnitudes, SECOND_VALUE);

            let mut negatives = vector[];
            vector::push_back(&mut negatives, false);
            vector::push_back(&mut negatives, false);

            let mut merkle_roots = vector[];
            vector::push_back(&mut merkle_roots, VALID_MERKLE_ROOT);
            vector::push_back(&mut merkle_roots, SECOND_MERKLE_ROOT);

            let mut alg_hashes = vector[];
            vector::push_back(&mut alg_hashes, VALID_ALG_HASH);
            vector::push_back(&mut alg_hashes, VALID_ALG_HASH);

            let mut rs = vector[];
            vector::push_back(&mut rs, VALID_R);
            vector::push_back(&mut rs, SECOND_R);

            let mut ss = vector[];
            vector::push_back(&mut ss, VALID_S);
            vector::push_back(&mut ss, SECOND_S);

            let mut vs = vector[];
            vector::push_back(&mut vs, VALID_V);
            vector::push_back(&mut vs, SECOND_V);

            let updates = update_temporal_numeric_value_evm_input_vec::new(
                ids, timestamps, magnitudes, negatives,
                merkle_roots, alg_hashes, rs, ss, vs
            );

            // Exact required fee for 2 updates. If the loop miscounted num_updates,
            // this call would abort with EInsufficientFee.
            let fee = coin::mint_for_testing<SUI>(2 * SINGLE_UPDATE_FEE, test_scenario::ctx(&mut scenario));

            stork::update_multiple_temporal_numeric_values_evm(
                &mut state,
                updates,
                fee,
                test_scenario::ctx(&mut scenario)
            );

            let first = stork::get_temporal_numeric_value_unchecked(&state, VALID_ID);
            assert!(first.get_timestamp_ns() == VALID_RECV_TIME, 0);
            assert!(first.get_quantized_value() == i128::from_u128(62507457175499998000000), 0);

            let second = stork::get_temporal_numeric_value_unchecked(&state, SECOND_ID);
            assert!(second.get_timestamp_ns() == SECOND_RECV_TIME, 0);
            assert!(second.get_quantized_value() == i128::from_u128(SECOND_VALUE), 0);

            test_scenario::return_shared(state);
        };

        test_scenario::end(scenario);
    }

    #[test]
    #[expected_failure(abort_code = stork::ENoFreshUpdate)]
    fun test_single_update_stale_aborts() {
        let mut scenario = test_scenario::begin(DEPLOYER);

        {
            admin::test_init(test_scenario::ctx(&mut scenario));
        };
        test_scenario::next_tx(&mut scenario, DEPLOYER);
        {
            let admin_cap = test_scenario::take_from_sender<AdminCap>(&scenario);
            stork::init_stork(
                &admin_cap,
                STORK_SUI_PUBLIC_KEY,
                STORK_EVM_PUBLIC_KEY,
                SINGLE_UPDATE_FEE,
                VERSION,
                test_scenario::ctx(&mut scenario)
            );
            test_scenario::return_to_sender(&scenario, admin_cap);
        };

        // Seed the feed with a fresh update.
        test_scenario::next_tx(&mut scenario, DEPLOYER);
        {
            let mut state = test_scenario::take_shared<StorkState>(&scenario);
            let update = update_temporal_numeric_value_evm_input::new(
                VALID_ID,
                VALID_RECV_TIME,
                62507457175499998000000,
                false,
                VALID_MERKLE_ROOT,
                VALID_ALG_HASH,
                VALID_R,
                VALID_S,
                VALID_V
            );
            let fee = coin::mint_for_testing<SUI>(SINGLE_UPDATE_FEE, test_scenario::ctx(&mut scenario));
            stork::update_single_temporal_numeric_value_evm(
                &mut state,
                update,
                fee,
                test_scenario::ctx(&mut scenario)
            );
            test_scenario::return_shared(state);
        };

        // Resubmit the same (now-stale) update — must abort, fee coin stays with caller.
        test_scenario::next_tx(&mut scenario, DEPLOYER);
        {
            let mut state = test_scenario::take_shared<StorkState>(&scenario);
            let dup = update_temporal_numeric_value_evm_input::new(
                VALID_ID,
                VALID_RECV_TIME,
                62507457175499998000000,
                false,
                VALID_MERKLE_ROOT,
                VALID_ALG_HASH,
                VALID_R,
                VALID_S,
                VALID_V
            );
            let fee = coin::mint_for_testing<SUI>(SINGLE_UPDATE_FEE, test_scenario::ctx(&mut scenario));
            stork::update_single_temporal_numeric_value_evm(
                &mut state,
                dup,
                fee,
                test_scenario::ctx(&mut scenario)
            );
            test_scenario::return_shared(state);
        };

        test_scenario::end(scenario);
    }

    #[test]
    #[expected_failure(abort_code = stork::ENoFreshUpdate)]
    fun test_batch_all_stale_aborts() {
        let mut scenario = test_scenario::begin(DEPLOYER);

        {
            admin::test_init(test_scenario::ctx(&mut scenario));
        };
        test_scenario::next_tx(&mut scenario, DEPLOYER);
        {
            let admin_cap = test_scenario::take_from_sender<AdminCap>(&scenario);
            stork::init_stork(
                &admin_cap,
                STORK_SUI_PUBLIC_KEY,
                STORK_EVM_PUBLIC_KEY,
                SINGLE_UPDATE_FEE,
                VERSION,
                test_scenario::ctx(&mut scenario)
            );
            test_scenario::return_to_sender(&scenario, admin_cap);
        };

        // Seed the feed.
        test_scenario::next_tx(&mut scenario, DEPLOYER);
        {
            let mut state = test_scenario::take_shared<StorkState>(&scenario);
            let update = update_temporal_numeric_value_evm_input::new(
                VALID_ID,
                VALID_RECV_TIME,
                62507457175499998000000,
                false,
                VALID_MERKLE_ROOT,
                VALID_ALG_HASH,
                VALID_R,
                VALID_S,
                VALID_V
            );
            let fee = coin::mint_for_testing<SUI>(SINGLE_UPDATE_FEE, test_scenario::ctx(&mut scenario));
            stork::update_single_temporal_numeric_value_evm(
                &mut state,
                update,
                fee,
                test_scenario::ctx(&mut scenario)
            );
            test_scenario::return_shared(state);
        };

        // Submit a batch containing only the same (now-stale) update — must abort.
        test_scenario::next_tx(&mut scenario, DEPLOYER);
        {
            let mut state = test_scenario::take_shared<StorkState>(&scenario);

            let mut ids = vector[];
            vector::push_back(&mut ids, VALID_ID);
            let mut timestamps = vector[];
            vector::push_back(&mut timestamps, VALID_RECV_TIME);
            let mut magnitudes = vector[];
            vector::push_back(&mut magnitudes, 62507457175499998000000);
            let mut negatives = vector[];
            vector::push_back(&mut negatives, false);
            let mut merkle_roots = vector[];
            vector::push_back(&mut merkle_roots, VALID_MERKLE_ROOT);
            let mut alg_hashes = vector[];
            vector::push_back(&mut alg_hashes, VALID_ALG_HASH);
            let mut rs = vector[];
            vector::push_back(&mut rs, VALID_R);
            let mut ss = vector[];
            vector::push_back(&mut ss, VALID_S);
            let mut vs = vector[];
            vector::push_back(&mut vs, VALID_V);

            let updates = update_temporal_numeric_value_evm_input_vec::new(
                ids, timestamps, magnitudes, negatives,
                merkle_roots, alg_hashes, rs, ss, vs
            );
            let fee = coin::mint_for_testing<SUI>(SINGLE_UPDATE_FEE, test_scenario::ctx(&mut scenario));
            stork::update_multiple_temporal_numeric_values_evm(
                &mut state,
                updates,
                fee,
                test_scenario::ctx(&mut scenario)
            );

            test_scenario::return_shared(state);
        };

        test_scenario::end(scenario);
    }

    #[test]
    fun test_withdraw_fees() {
        let mut scenario = test_scenario::begin(DEPLOYER);
        
        // Setup initial state
        {
            admin::test_init(test_scenario::ctx(&mut scenario));
        };
        test_scenario::next_tx(&mut scenario, DEPLOYER);
        {
            let admin_cap = test_scenario::take_from_sender<AdminCap>(&scenario);
            stork::init_stork(
                &admin_cap,
                STORK_SUI_PUBLIC_KEY,
                STORK_EVM_PUBLIC_KEY,
                SINGLE_UPDATE_FEE,
                VERSION,
                test_scenario::ctx(&mut scenario)
            );
            test_scenario::return_to_sender(&scenario, admin_cap);
        };

        // Deposit some fees via an update
        test_scenario::next_tx(&mut scenario, DEPLOYER);
        {
            let mut state = test_scenario::take_shared<StorkState>(&scenario);
            let update = update_temporal_numeric_value_evm_input::new(
                VALID_ID,
                VALID_RECV_TIME,
                62507457175499998000000,
                false,
                VALID_MERKLE_ROOT,
                VALID_ALG_HASH,
                VALID_R,
                VALID_S,
                VALID_V
            );
            
            let fee = coin::mint_for_testing<SUI>(2000, test_scenario::ctx(&mut scenario));
            let expected_amount = coin::value(&fee);
            
            stork::update_single_temporal_numeric_value_evm(
                &mut state,
                update,
                fee,
                test_scenario::ctx(&mut scenario)
            );
            
            test_scenario::return_shared(state);

            // Withdraw and verify fees
            test_scenario::next_tx(&mut scenario, DEPLOYER);
            {
                let admin_cap = test_scenario::take_from_sender<AdminCap>(&scenario);
                let mut state = test_scenario::take_shared<StorkState>(&scenario);
                
                let withdrawn_coins = state::withdraw_fees(&admin_cap, &mut state, test_scenario::ctx(&mut scenario));
                assert!(coin::value(&withdrawn_coins) == expected_amount, 0);
                
                // Clean up the coin
                unit_test::destroy(withdrawn_coins);
                
                test_scenario::return_shared(state);
                test_scenario::return_to_sender(&scenario, admin_cap);
            };
        };

        test_scenario::end(scenario);
    }

    // === Fee & freshness edge cases ===

    // A batch mixing one stale entry and one fresh entry: only the fresh one is
    // applied, num_updates == 1, and the required fee is exactly one update's worth.
    // The exact fee coin (1 * SINGLE_UPDATE_FEE) also guards the per-entry fee math:
    // if the stale entry were counted, this would abort with EInsufficientFee.
    #[test]
    fun test_batch_mixed_fresh_and_stale() {
        let mut scenario = test_scenario::begin(DEPLOYER);

        {
            admin::test_init(test_scenario::ctx(&mut scenario));
        };
        test_scenario::next_tx(&mut scenario, DEPLOYER);
        {
            let admin_cap = test_scenario::take_from_sender<AdminCap>(&scenario);
            stork::init_stork(
                &admin_cap, STORK_SUI_PUBLIC_KEY, STORK_EVM_PUBLIC_KEY,
                SINGLE_UPDATE_FEE, VERSION, test_scenario::ctx(&mut scenario)
            );
            test_scenario::return_to_sender(&scenario, admin_cap);
        };

        // Seed the VALID feed so it becomes stale for the batch below.
        test_scenario::next_tx(&mut scenario, DEPLOYER);
        {
            let mut state = test_scenario::take_shared<StorkState>(&scenario);
            let update = update_temporal_numeric_value_evm_input::new(
                VALID_ID, VALID_RECV_TIME, 62507457175499998000000, false,
                VALID_MERKLE_ROOT, VALID_ALG_HASH, VALID_R, VALID_S, VALID_V
            );
            let fee = coin::mint_for_testing<SUI>(SINGLE_UPDATE_FEE, test_scenario::ctx(&mut scenario));
            stork::update_single_temporal_numeric_value_evm(&mut state, update, fee, test_scenario::ctx(&mut scenario));
            test_scenario::return_shared(state);
        };

        // Batch: entry 0 is the now-stale VALID update, entry 1 is a fresh SECOND update.
        test_scenario::next_tx(&mut scenario, DEPLOYER);
        {
            let mut state = test_scenario::take_shared<StorkState>(&scenario);

            let mut ids = vector[];
            vector::push_back(&mut ids, VALID_ID);
            vector::push_back(&mut ids, SECOND_ID);
            let mut timestamps = vector[];
            vector::push_back(&mut timestamps, VALID_RECV_TIME);
            vector::push_back(&mut timestamps, SECOND_RECV_TIME);
            let mut magnitudes = vector[];
            vector::push_back(&mut magnitudes, 62507457175499998000000);
            vector::push_back(&mut magnitudes, SECOND_VALUE);
            let mut negatives = vector[];
            vector::push_back(&mut negatives, false);
            vector::push_back(&mut negatives, false);
            let mut merkle_roots = vector[];
            vector::push_back(&mut merkle_roots, VALID_MERKLE_ROOT);
            vector::push_back(&mut merkle_roots, SECOND_MERKLE_ROOT);
            let mut alg_hashes = vector[];
            vector::push_back(&mut alg_hashes, VALID_ALG_HASH);
            vector::push_back(&mut alg_hashes, VALID_ALG_HASH);
            let mut rs = vector[];
            vector::push_back(&mut rs, VALID_R);
            vector::push_back(&mut rs, SECOND_R);
            let mut ss = vector[];
            vector::push_back(&mut ss, VALID_S);
            vector::push_back(&mut ss, SECOND_S);
            let mut vs = vector[];
            vector::push_back(&mut vs, VALID_V);
            vector::push_back(&mut vs, SECOND_V);

            let updates = update_temporal_numeric_value_evm_input_vec::new(
                ids, timestamps, magnitudes, negatives, merkle_roots, alg_hashes, rs, ss, vs
            );

            // Exactly one update's fee — proves only the fresh entry was counted.
            let fee = coin::mint_for_testing<SUI>(SINGLE_UPDATE_FEE, test_scenario::ctx(&mut scenario));
            stork::update_multiple_temporal_numeric_values_evm(&mut state, updates, fee, test_scenario::ctx(&mut scenario));

            // Stale entry left the VALID feed untouched at its original timestamp.
            let first = stork::get_temporal_numeric_value_unchecked(&state, VALID_ID);
            assert!(first.get_timestamp_ns() == VALID_RECV_TIME, 0);
            // Fresh entry was applied.
            let second = stork::get_temporal_numeric_value_unchecked(&state, SECOND_ID);
            assert!(second.get_timestamp_ns() == SECOND_RECV_TIME, 0);

            test_scenario::return_shared(state);
        };

        test_scenario::end(scenario);
    }

    // A single update whose fee coin is smaller than the required fee must abort.
    #[test]
    #[expected_failure(abort_code = stork::EInsufficientFee)]
    fun test_single_update_insufficient_fee_aborts() {
        let mut scenario = test_scenario::begin(DEPLOYER);

        {
            admin::test_init(test_scenario::ctx(&mut scenario));
        };
        test_scenario::next_tx(&mut scenario, DEPLOYER);
        {
            let admin_cap = test_scenario::take_from_sender<AdminCap>(&scenario);
            stork::init_stork(
                &admin_cap, STORK_SUI_PUBLIC_KEY, STORK_EVM_PUBLIC_KEY,
                SINGLE_UPDATE_FEE, VERSION, test_scenario::ctx(&mut scenario)
            );
            test_scenario::return_to_sender(&scenario, admin_cap);
        };

        test_scenario::next_tx(&mut scenario, DEPLOYER);
        {
            let mut state = test_scenario::take_shared<StorkState>(&scenario);
            let update = update_temporal_numeric_value_evm_input::new(
                VALID_ID, VALID_RECV_TIME, 62507457175499998000000, false,
                VALID_MERKLE_ROOT, VALID_ALG_HASH, VALID_R, VALID_S, VALID_V
            );
            // One mist short of the required fee.
            let fee = coin::mint_for_testing<SUI>(SINGLE_UPDATE_FEE - 1, test_scenario::ctx(&mut scenario));
            stork::update_single_temporal_numeric_value_evm(&mut state, update, fee, test_scenario::ctx(&mut scenario));
            test_scenario::return_shared(state);
        };

        test_scenario::end(scenario);
    }

    // A batch whose fee coin is smaller than the required fee must abort.
    #[test]
    #[expected_failure(abort_code = stork::EInsufficientFee)]
    fun test_batch_insufficient_fee_aborts() {
        let mut scenario = test_scenario::begin(DEPLOYER);

        {
            admin::test_init(test_scenario::ctx(&mut scenario));
        };
        test_scenario::next_tx(&mut scenario, DEPLOYER);
        {
            let admin_cap = test_scenario::take_from_sender<AdminCap>(&scenario);
            stork::init_stork(
                &admin_cap, STORK_SUI_PUBLIC_KEY, STORK_EVM_PUBLIC_KEY,
                SINGLE_UPDATE_FEE, VERSION, test_scenario::ctx(&mut scenario)
            );
            test_scenario::return_to_sender(&scenario, admin_cap);
        };

        test_scenario::next_tx(&mut scenario, DEPLOYER);
        {
            let mut state = test_scenario::take_shared<StorkState>(&scenario);

            let mut ids = vector[];
            vector::push_back(&mut ids, VALID_ID);
            let mut timestamps = vector[];
            vector::push_back(&mut timestamps, VALID_RECV_TIME);
            let mut magnitudes = vector[];
            vector::push_back(&mut magnitudes, 62507457175499998000000);
            let mut negatives = vector[];
            vector::push_back(&mut negatives, false);
            let mut merkle_roots = vector[];
            vector::push_back(&mut merkle_roots, VALID_MERKLE_ROOT);
            let mut alg_hashes = vector[];
            vector::push_back(&mut alg_hashes, VALID_ALG_HASH);
            let mut rs = vector[];
            vector::push_back(&mut rs, VALID_R);
            let mut ss = vector[];
            vector::push_back(&mut ss, VALID_S);
            let mut vs = vector[];
            vector::push_back(&mut vs, VALID_V);

            let updates = update_temporal_numeric_value_evm_input_vec::new(
                ids, timestamps, magnitudes, negatives, merkle_roots, alg_hashes, rs, ss, vs
            );
            // One fresh update requires SINGLE_UPDATE_FEE; provide one mist less.
            let fee = coin::mint_for_testing<SUI>(SINGLE_UPDATE_FEE - 1, test_scenario::ctx(&mut scenario));
            stork::update_multiple_temporal_numeric_values_evm(&mut state, updates, fee, test_scenario::ctx(&mut scenario));
            test_scenario::return_shared(state);
        };

        test_scenario::end(scenario);
    }

    // === Setup, admin, and upgrade paths ===

    // init_stork has no reinitialization guard: calling it twice creates two
    // independent shared StorkState objects.
    #[test]
    fun test_init_stork_twice_creates_two_states() {
        let mut scenario = test_scenario::begin(DEPLOYER);

        {
            admin::test_init(test_scenario::ctx(&mut scenario));
        };
        test_scenario::next_tx(&mut scenario, DEPLOYER);
        {
            let admin_cap = test_scenario::take_from_sender<AdminCap>(&scenario);
            stork::init_stork(
                &admin_cap, STORK_SUI_PUBLIC_KEY, STORK_EVM_PUBLIC_KEY,
                SINGLE_UPDATE_FEE, VERSION, test_scenario::ctx(&mut scenario)
            );
            test_scenario::return_to_sender(&scenario, admin_cap);
        };
        // Second initialization with the same AdminCap.
        test_scenario::next_tx(&mut scenario, DEPLOYER);
        {
            let admin_cap = test_scenario::take_from_sender<AdminCap>(&scenario);
            stork::init_stork(
                &admin_cap, STORK_SUI_PUBLIC_KEY, STORK_EVM_PUBLIC_KEY,
                SINGLE_UPDATE_FEE, VERSION, test_scenario::ctx(&mut scenario)
            );
            test_scenario::return_to_sender(&scenario, admin_cap);
        };

        // Exactly two shared StorkState objects now exist.
        test_scenario::next_tx(&mut scenario, DEPLOYER);
        {
            let first = test_scenario::take_shared<StorkState>(&scenario);
            assert!(test_scenario::has_most_recent_shared<StorkState>(), 0);
            let second = test_scenario::take_shared<StorkState>(&scenario);
            assert!(!test_scenario::has_most_recent_shared<StorkState>(), 0);
            test_scenario::return_shared(first);
            test_scenario::return_shared(second);
        };

        test_scenario::end(scenario);
    }

    // withdraw_fees on a treasury that has never collected a fee must abort.
    #[test]
    #[expected_failure(abort_code = state::ENoFeesToWithdraw)]
    fun test_withdraw_fees_empty_treasury_aborts() {
        let mut scenario = test_scenario::begin(DEPLOYER);

        {
            admin::test_init(test_scenario::ctx(&mut scenario));
        };
        test_scenario::next_tx(&mut scenario, DEPLOYER);
        {
            let admin_cap = test_scenario::take_from_sender<AdminCap>(&scenario);
            stork::init_stork(
                &admin_cap, STORK_SUI_PUBLIC_KEY, STORK_EVM_PUBLIC_KEY,
                SINGLE_UPDATE_FEE, VERSION, test_scenario::ctx(&mut scenario)
            );
            test_scenario::return_to_sender(&scenario, admin_cap);
        };

        test_scenario::next_tx(&mut scenario, DEPLOYER);
        {
            let admin_cap = test_scenario::take_from_sender<AdminCap>(&scenario);
            let mut state = test_scenario::take_shared<StorkState>(&scenario);
            // No update has deposited a fee — treasury is empty.
            let withdrawn = state::withdraw_fees(&admin_cap, &mut state, test_scenario::ctx(&mut scenario));
            unit_test::destroy(withdrawn);
            test_scenario::return_shared(state);
            test_scenario::return_to_sender(&scenario, admin_cap);
        };

        test_scenario::end(scenario);
    }

    // migrate requires the current state version to be VERSION - 1. A freshly
    // initialized state is at VERSION, so migrate must abort.
    #[test]
    #[expected_failure(abort_code = state::EIncorrectVersion)]
    fun test_migrate_wrong_version_aborts() {
        let mut scenario = test_scenario::begin(DEPLOYER);

        {
            admin::test_init(test_scenario::ctx(&mut scenario));
        };
        test_scenario::next_tx(&mut scenario, DEPLOYER);
        {
            let admin_cap = test_scenario::take_from_sender<AdminCap>(&scenario);
            stork::init_stork(
                &admin_cap, STORK_SUI_PUBLIC_KEY, STORK_EVM_PUBLIC_KEY,
                SINGLE_UPDATE_FEE, VERSION, test_scenario::ctx(&mut scenario)
            );
            test_scenario::return_to_sender(&scenario, admin_cap);
        };

        test_scenario::next_tx(&mut scenario, DEPLOYER);
        {
            let admin_cap = test_scenario::take_from_sender<AdminCap>(&scenario);
            let mut state = test_scenario::take_shared<StorkState>(&scenario);
            // state.version == VERSION (3), but migrate expects VERSION - 1 (2).
            state::migrate(&admin_cap, &mut state, VERSION, @0x123);
            test_scenario::return_shared(state);
            test_scenario::return_to_sender(&scenario, admin_cap);
        };

        test_scenario::end(scenario);
    }

    // migrate success path: from a state pinned at VERSION - 1, migrate bumps the
    // version to VERSION and updates the stork sui address.
    #[test]
    fun test_migrate_success() {
        let mut scenario = test_scenario::begin(DEPLOYER);

        {
            admin::test_init(test_scenario::ctx(&mut scenario));
        };
        test_scenario::next_tx(&mut scenario, DEPLOYER);
        {
            let admin_cap = test_scenario::take_from_sender<AdminCap>(&scenario);
            stork::init_stork(
                &admin_cap, STORK_SUI_PUBLIC_KEY, STORK_EVM_PUBLIC_KEY,
                SINGLE_UPDATE_FEE, VERSION, test_scenario::ctx(&mut scenario)
            );
            test_scenario::return_to_sender(&scenario, admin_cap);
        };

        test_scenario::next_tx(&mut scenario, DEPLOYER);
        {
            let admin_cap = test_scenario::take_from_sender<AdminCap>(&scenario);
            let mut state = test_scenario::take_shared<StorkState>(&scenario);

            // Simulate a pre-upgrade state one version behind.
            state::set_version_for_testing(&mut state, VERSION - 1);

            let new_address = @0x123;
            state::migrate(&admin_cap, &mut state, VERSION, new_address);

            assert!(state.get_version() == VERSION, 0);
            assert!(state.get_stork_sui_address() == new_address, 0);

            test_scenario::return_shared(state);
            test_scenario::return_to_sender(&scenario, admin_cap);
        };

        test_scenario::end(scenario);
    }

    // set_version directly overrides the stored version with no sequential-upgrade
    // guard, recovering a state stuck at an arbitrary version (unlike migrate,
    // which requires exactly VERSION - 1).
    #[test]
    fun test_set_version_success() {
        let mut scenario = test_scenario::begin(DEPLOYER);

        {
            admin::test_init(test_scenario::ctx(&mut scenario));
        };
        test_scenario::next_tx(&mut scenario, DEPLOYER);
        {
            let admin_cap = test_scenario::take_from_sender<AdminCap>(&scenario);
            stork::init_stork(
                &admin_cap, STORK_SUI_PUBLIC_KEY, STORK_EVM_PUBLIC_KEY,
                SINGLE_UPDATE_FEE, VERSION, test_scenario::ctx(&mut scenario)
            );
            test_scenario::return_to_sender(&scenario, admin_cap);
        };

        test_scenario::next_tx(&mut scenario, DEPLOYER);
        {
            let admin_cap = test_scenario::take_from_sender<AdminCap>(&scenario);
            let mut state = test_scenario::take_shared<StorkState>(&scenario);

            // Simulate a state stuck several versions behind — migrate would
            // abort here since it requires exactly VERSION - 1.
            state::set_version_for_testing(&mut state, VERSION - 2);

            state::set_version(&admin_cap, &mut state, VERSION);

            assert!(state.get_version() == VERSION, 0);

            test_scenario::return_shared(state);
            test_scenario::return_to_sender(&scenario, admin_cap);
        };

        test_scenario::end(scenario);
    }

    // Functions guarded by the version check must abort when the stored version
    // does not match the package VERSION (e.g. after an upgrade, before migrate).
    #[test]
    #[expected_failure(abort_code = state::EIncorrectVersion)]
    fun test_version_check_aborts_on_getter() {
        let mut scenario = test_scenario::begin(DEPLOYER);

        {
            admin::test_init(test_scenario::ctx(&mut scenario));
        };
        test_scenario::next_tx(&mut scenario, DEPLOYER);
        {
            let admin_cap = test_scenario::take_from_sender<AdminCap>(&scenario);
            stork::init_stork(
                &admin_cap, STORK_SUI_PUBLIC_KEY, STORK_EVM_PUBLIC_KEY,
                SINGLE_UPDATE_FEE, VERSION, test_scenario::ctx(&mut scenario)
            );
            test_scenario::return_to_sender(&scenario, admin_cap);
        };

        test_scenario::next_tx(&mut scenario, DEPLOYER);
        {
            let mut state = test_scenario::take_shared<StorkState>(&scenario);
            // Put the state at a version that no longer matches VERSION.
            state::set_version_for_testing(&mut state, VERSION - 1);
            // Any version-guarded accessor must now abort.
            let _ = state.get_single_update_fee_in_mist();
            test_scenario::return_shared(state);
        };

        test_scenario::end(scenario);
    }

    // update_stork_evm_public_key must reject an EVM key that is not 20 bytes.
    #[test]
    #[expected_failure(abort_code = evm_pubkey::EInvalidLength)]
    fun test_update_stork_evm_public_key_wrong_length_aborts() {
        let mut scenario = test_scenario::begin(DEPLOYER);

        {
            admin::test_init(test_scenario::ctx(&mut scenario));
        };
        test_scenario::next_tx(&mut scenario, DEPLOYER);
        {
            let admin_cap = test_scenario::take_from_sender<AdminCap>(&scenario);
            stork::init_stork(
                &admin_cap, STORK_SUI_PUBLIC_KEY, STORK_EVM_PUBLIC_KEY,
                SINGLE_UPDATE_FEE, VERSION, test_scenario::ctx(&mut scenario)
            );
            test_scenario::return_to_sender(&scenario, admin_cap);
        };

        test_scenario::next_tx(&mut scenario, DEPLOYER);
        {
            let admin_cap = test_scenario::take_from_sender<AdminCap>(&scenario);
            let mut state = test_scenario::take_shared<StorkState>(&scenario);
            // 19 bytes instead of 20.
            let short_key = x"11111111111111111111111111111111111111";
            state::update_stork_evm_public_key(&admin_cap, &mut state, short_key);
            test_scenario::return_shared(state);
            test_scenario::return_to_sender(&scenario, admin_cap);
        };

        test_scenario::end(scenario);
    }
}
