module stork::state_object_store {

    // === Imports ===

    use initia_std::object::{Self, ExtendRef};
    
    // === Constants ===

    const STATE_OBJECT_SEED: vector<u8> = b"state-object";

    // === Structs ===

    struct StateObjectStore has key {
        state_object_extend_ref: ExtendRef 
    }

    // === Functions ===

    /// Runs on publish, sets up the state account store
    fun init_module(package: &signer){
        let constructor_ref = object::create_named_object(package, STATE_OBJECT_SEED);
        let extend_ref = object::generate_extend_ref(&constructor_ref);
        let transfer_ref = object::generate_transfer_ref(&constructor_ref);

        object::disable_ungated_transfer(&transfer_ref);

        let state_object_store = new(extend_ref);

        move_to(
            package,
            state_object_store
        );
    }

    fun new(state_object_extend_ref: ExtendRef): StateObjectStore {
        StateObjectStore {
            state_object_extend_ref
        }
    }

    #[view]
    /// Returns the address of the state object
    public fun get_state_object_address(): address acquires StateObjectStore {
        let state_object_store = borrow_global<StateObjectStore>(@stork);
        object::address_from_extend_ref(&state_object_store.state_object_extend_ref)
    }

    /// Returns the signer for the state object
    package fun get_state_object_signer(): signer acquires StateObjectStore {
        let state_object_store = borrow_global<StateObjectStore>(@stork);
        object::generate_signer_for_extending(&state_object_store.state_object_extend_ref)
    }

    // === Test Imports ===

    #[test_only]
    use initia_std::signer;
    #[test_only]
    use initia_std::account;
    // === Tests Constants ===

    const STORK: address = @stork;

    // === Test Helpers ===

    #[test_only]
    package fun init_module_for_test(package: &signer) {
        init_module(package);
    }

    // === Tests ===

    #[test]
    fun test_state_object_store() {
        let package = account::create_account_for_test(STORK);
        init_module_for_test(&package);
        assert!(exists<StateObjectStore>(STORK), 0);
    }

    #[test]
    fun test_get_state_object_address() acquires StateObjectStore {
        let package = account::create_account_for_test(STORK);
        init_module_for_test(&package);
        assert!(get_state_object_address() == object::create_object_address(&STORK, STATE_OBJECT_SEED), 0);
    }

    #[test]
    fun test_get_state_object_signer() acquires StateObjectStore {
        let package = account::create_account_for_test(STORK);
        init_module_for_test(&package);
        let signer = get_state_object_signer();
        assert!(signer::address_of(&signer) == object::create_object_address(&STORK, STATE_OBJECT_SEED), 0);
    }
}
