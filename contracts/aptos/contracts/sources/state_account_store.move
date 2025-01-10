module stork::state_account_store {

    // === Imports ===

    use aptos_framework::account::{Self, SignerCapability};
    use aptos_std::signer;
    // === Constants ===

    const STATE_ACCOUNT_SEED: vector<u8> = b"state-account";

    // === Structs ===

    struct StateAccountStore has key {
        state_account_signer_cap: SignerCapability
    }

    // === Functions ===

    /// Runs on publish, sets up the state account store
    fun init_module(package: &signer){
        let (state_account_signer, state_account_signer_cap) = account::create_resource_account(package, STATE_ACCOUNT_SEED);
        let state_account_store = new(state_account_signer_cap);
        move_to(
            package,
            state_account_store
        );
    }

    fun new(state_account_signer_cap: SignerCapability): StateAccountStore {
        StateAccountStore {
            state_account_signer_cap
        }
    }

    #[view]
    /// Returns the address of the state object
    public fun get_state_account_address(): address acquires StateAccountStore {
        let state_account_store = borrow_global<StateAccountStore>(@stork);
        account::get_signer_capability_address(&state_account_store.state_account_signer_cap)
    }

    /// Returns the signer for the state object
    package fun get_state_account_signer(): signer acquires StateAccountStore {
        let state_account_store = borrow_global<StateAccountStore>(@stork);
        account::create_signer_with_capability(&state_account_store.state_account_signer_cap)
    }

    // === Tests Constants ===

    const STORK: address = @stork;

    // === Test Helpers ===

    #[test_only]
    package fun init_module_for_test(package: &signer) {
        init_module(package);
    }

    // === Tests ===

    #[test]
    fun test_state_account_store() {
        let package = account::create_account_for_test(STORK);
        init_module_for_test(&package);
        assert!(exists<StateAccountStore>(STORK), 0);
    }

    #[test]
    fun test_get_state_account_address() acquires StateAccountStore {
        let package = account::create_account_for_test(STORK);
        init_module_for_test(&package);
        assert!(get_state_account_address() == account::create_resource_address(&STORK, STATE_ACCOUNT_SEED), 0);
    }

    #[test]
    fun test_get_state_account_signer() acquires StateAccountStore {
        let package = account::create_account_for_test(STORK);
        init_module_for_test(&package);
        let signer = get_state_account_signer();
        assert!(signer::address_of(&signer) == account::create_resource_address(&STORK, STATE_ACCOUNT_SEED), 0);
    }
}
