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
    use sui::test_utils::Self;

    // === Constants ===

    const DEPLOYER: address = @0x26;
    const STORK_SUI_PUBLIC_KEY: address = @0x42;
    const SINGLE_UPDATE_FEE: u64 = 1000;
    const VERSION: u64 = 1;
    const STORK_EVM_PUBLIC_KEY: vector<u8> = x"0a803F9b1CCe32e2773e0d2e98b37E0775cA5d44";

    // Constants from verify.move test cases
    const VALID_ID: vector<u8> = x"7404e3d104ea7841c3d9e6fd20adfe99b4ad586bc08d8f3bd3afef894cf184de";
    const VALID_RECV_TIME: u64 = 1722632569208762117;
    const VALID_MERKLE_ROOT: vector<u8> = x"e5ff773b0316059c04aa157898766731017610dcbeede7d7f169bfeaab7cc318";
    const VALID_ALG_HASH: vector<u8> = x"9be7e9f9ed459417d96112a7467bd0b27575a2c7847195c68f805b70ce1795ba";
    const VALID_R: vector<u8> = x"b9b3c9f80a355bd0cd6f609fff4a4b15fa4e3b4632adabb74c020f5bcd240741";
    const VALID_S: vector<u8> = x"16fab526529ac795108d201832cff8c2d2b1c710da6711fe9f7ab288a7149758";
    const VALID_V: u8 = 28;

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
            let mut ids = vector::empty();
            vector::push_back(&mut ids, VALID_ID);
            
            let mut timestamps = vector::empty();
            vector::push_back(&mut timestamps, VALID_RECV_TIME);
            
            let mut magnitudes = vector::empty();
            vector::push_back(&mut magnitudes, 62507457175499998000000);
            
            let mut negatives = vector::empty();
            vector::push_back(&mut negatives, false);
            
            let mut merkle_roots = vector::empty();
            vector::push_back(&mut merkle_roots, VALID_MERKLE_ROOT);
            
            let mut alg_hashes = vector::empty();
            vector::push_back(&mut alg_hashes, VALID_ALG_HASH);
            
            let mut rs = vector::empty();
            vector::push_back(&mut rs, VALID_R);
            
            let mut ss = vector::empty();
            vector::push_back(&mut ss, VALID_S);
            
            let mut vs = vector::empty();
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
                test_utils::destroy(withdrawn_coins);
                
                test_scenario::return_shared(state);
                test_scenario::return_to_sender(&scenario, admin_cap);
            };
        };

        test_scenario::end(scenario);
    }
}   