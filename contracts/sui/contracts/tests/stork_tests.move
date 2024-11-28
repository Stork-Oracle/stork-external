#[test_only]
module stork::stork_tests {

    // === Imports ===

    use stork::stork;
    use sui::test_scenario::{Self, Scenario};
    use stork::admin::{Self, AdminCap};
    use stork::state::{Self, StorkState};

    // === Constants ===

    const DEPLOYER: address = @0x26;
    const STORK_SUI_PUBLIC_KEY: address = @0x42;
    const SINGLE_UPDATE_FEE: u64 = 1000;
    const VERSION: u64 = 1;
    const STORK_EVM_PUBLIC_KEY: vector<u8> = x"0a803F9b1CCe32e2773e0d2e98b37E0775cA5d44";

    // === Errors ===

    const ENotImplemented: u64 = 0;

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
            assert!(state.get_stork_sui_public_key() == STORK_SUI_PUBLIC_KEY);
            assert!(state.get_stork_evm_public_key().get_bytes() == STORK_EVM_PUBLIC_KEY);
            assert!(state.get_single_update_fee_in_mist() == SINGLE_UPDATE_FEE);
            assert!(state.get_version() == VERSION);
            test_scenario::return_shared(state);
        };
        test_scenario::end(scenario);
    }


}   