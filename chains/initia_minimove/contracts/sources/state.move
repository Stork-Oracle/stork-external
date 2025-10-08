module stork::state {
    // === Imports ===  
    
    use stork::evm_pubkey::{Self, EvmPubKey};
    use stork::state_object_store;
    use stork::fee::{Self, Fee};
    use initia_std::signer;
    use initia_std::string::String;

    // === Errors ===

    const E_NOT_OWNER: u64 = 0;

    // == Structs ==

    /// State object for the Stork contract
    struct StorkState has key {
        // Stork's EVM public key
        stork_evm_public_key: EvmPubKey,
        // fee for a single update
        single_update_fee: Fee,
        // owner of the Stork contract
        owner: address,
    }

    // === Functions ===

    /// Creates a new StorkState
    package fun new(
        stork_evm_public_key: EvmPubKey,
        single_update_fee: Fee,
        owner: address,
    ): StorkState {
        StorkState {
            stork_evm_public_key,
            single_update_fee,
            owner,
        }
    }

    /// Moves a StorkState to the given signer
    package fun move_state(self: StorkState, signer: &signer) {
        move_to(signer, self);
    }

    #[view]
    /// Returns the owner of the Stork contract
    public fun get_owner(): address acquires StorkState {
        borrow_global<StorkState>(state_object_store::get_state_object_address()).owner
    }

    #[view]
    /// Returns the Stork's EVM public key
    public fun get_stork_evm_public_key(): EvmPubKey acquires StorkState {
        borrow_global<StorkState>(state_object_store::get_state_object_address()).stork_evm_public_key
    }

    #[view]
    /// Returns the fee for a single update
    public fun get_single_update_fee(): Fee acquires StorkState {
        borrow_global<StorkState>(state_object_store::get_state_object_address()).single_update_fee
    }

    #[view]
    /// Returns true if the state exists
    public fun state_exists(): bool {
        exists<StorkState>(state_object_store::get_state_object_address())
    }

    /// === Admin Functions ===

    /// Sets the owner of the Stork contract
    public entry fun set_owner(owner: &signer, new_owner: address) acquires StorkState {
        let state = borrow_global_mut<StorkState>(state_object_store::get_state_object_address());
        assert!(
            signer::address_of(owner) == state.owner,
            E_NOT_OWNER
        );
        state.owner = new_owner;
    }

    /// Sets the fee for a single update
    public entry fun set_single_update_fee(owner: &signer, new_fee_amount: u64, new_fee_denom: String) acquires StorkState {
        let new_fee = fee::new(new_fee_amount, new_fee_denom);
        let state = borrow_global_mut<StorkState>(state_object_store::get_state_object_address());
        assert!(
            signer::address_of(owner) == state.owner,
            E_NOT_OWNER
        );
        state.single_update_fee = new_fee;
    }

    public entry fun set_stork_evm_public_key(owner: &signer, new_stork_evm_public_key: vector<u8>) acquires StorkState {
        let state = borrow_global_mut<StorkState>(state_object_store::get_state_object_address());
        assert!(
            signer::address_of(owner) == state.owner,
            E_NOT_OWNER
        );
        state.stork_evm_public_key = evm_pubkey::from_bytes(new_stork_evm_public_key);
    }

    // === Test Imports ===

    #[test_only]
    use initia_std::account::create_account_for_test;
    #[test_only]
    use initia_std::string;
    
    // === Test Constants ===

    #[test_only]
    const STORK: address = @stork;
    #[test_only]
    const DEPLOYER: address = @0xFACE;
    #[test_only]
    const USER: address = @0xCAFE;

    // === Test Helpers ===

    #[test_only]
    fun setup_test(): signer {
        let stork_signer = create_account_for_test(STORK);
        state_object_store::init_module_for_test(&stork_signer);
        let deployer_signer = create_account_for_test(DEPLOYER);
        let pubkey = evm_pubkey::create_zeroed_evm_pubkey();
        let fee = fee::new(1, string::utf8(b"unit"));
        let stork_state_object_signer = state_object_store::get_state_object_signer();

        let state = new(pubkey, fee, DEPLOYER);
        state.move_state(&stork_state_object_signer);
        deployer_signer
    }

    // === Tests ===

    #[test]
    fun test_state_initialization() acquires StorkState {
        setup_test();
        
        assert!(state_exists(), 0);
        assert!(get_single_update_fee() == fee::new(1, string::utf8(b"unit")), 1);
        assert!(get_stork_evm_public_key() == evm_pubkey::create_zeroed_evm_pubkey(), 3);
        assert!(get_owner() == DEPLOYER, 5);
    }

    #[test]
    fun test_set_owner() acquires StorkState {
        let deployer_signer = setup_test();
        
        set_owner(&deployer_signer, USER);
        assert!(get_owner() == USER, 0);
    }

    #[test]
    fun test_set_single_update_fee_amount() acquires StorkState {
        let deployer_signer = setup_test();

        let new_fee = fee::new(200, string::utf8(b"unit"));
        set_single_update_fee(&deployer_signer, new_fee.get_amount(), new_fee.get_denom());
        
        assert!(get_single_update_fee() == new_fee, 0);
        assert!(get_single_update_fee().get_amount() == 200, 1);
        assert!(get_single_update_fee().get_denom() == string::utf8(b"unit"), 2);
    }

    #[test]
    fun test_set_single_update_fee_denom() acquires StorkState {
        let deployer_signer = setup_test();

        let new_fee = fee::new(1, string::utf8(b"ustork"));
        set_single_update_fee(&deployer_signer, new_fee.get_amount(), new_fee.get_denom());

        assert!(get_single_update_fee() == new_fee, 0);
        assert!(get_single_update_fee().get_amount() == 1, 1);
        assert!(get_single_update_fee().get_denom() == string::utf8(b"ustork"), 2);
    }

    #[test]
    fun test_set_stork_evm_public_key() acquires StorkState {
        let deployer_signer = setup_test();

        let new_stork_evm_public_key = x"0a803F9b1CCe32e2773e0d2e98b37E0775cA5d44";
        set_stork_evm_public_key(&deployer_signer, new_stork_evm_public_key);
        
        assert!(get_stork_evm_public_key() == evm_pubkey::from_bytes(new_stork_evm_public_key), 0);
    }

    #[test]
    #[expected_failure(abort_code = E_NOT_OWNER)]
    fun test_set_owner_unauthorized() acquires StorkState {
        setup_test();        
        
        let user_signer = create_account_for_test(USER);
        set_owner(&user_signer, USER);
    }

    #[test]
    #[expected_failure(abort_code = E_NOT_OWNER)]
    fun test_set_fee_unauthorized() acquires StorkState {
        setup_test();
        
        let user_signer = create_account_for_test(USER);
        set_single_update_fee(&user_signer, 200, string::utf8(b"ustork"));
    }

    #[test]
    #[expected_failure(abort_code = E_NOT_OWNER)]
    fun test_set_stork_evm_public_key_unauthorized() acquires StorkState {
        setup_test();
        
        let user_signer = create_account_for_test(USER);
        set_stork_evm_public_key(&user_signer, x"0a803F9b1CCe32e2773e0d2e98b37E0775cA5d44");
    }
}
