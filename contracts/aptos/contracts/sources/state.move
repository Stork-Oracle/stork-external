module stork::state {
    // === Imports ===  
    
    use stork::evm_pubkey::EvmPubKey;
    use aptos_std::signer;

    // === Errors ===

    const E_NOT_OWNER: u64 = 0;

    // == Structs ==

    /// State object for the Stork contract
    struct StorkState has key {
        /// The address of the Stork contract
        stork_address: address,
        /// Stork's EVM public key
        stork_evm_public_key: EvmPubKey,
        /// The fee for a single update
        single_update_fee: u64,
        /// The owner of the Stork contract
        owner: address,
    }

    // === Functions ===

    /// Creates a new StorkState
    public fun new(
        stork_evm_public_key: EvmPubKey,
        single_update_fee: u64,
        owner: address,
    ): StorkState {
        StorkState {
            stork_address: @stork,
            stork_evm_public_key,
            single_update_fee,
            owner,
        }
    }

    /// Moves a StorkState to the given signer
    public fun move_state(self: StorkState, signer: &signer) {
        move_to(signer, self);
    }

    /// Returns the Stork's EVM public key
    public fun get_stork_evm_public_key(): EvmPubKey acquires StorkState {
        borrow_global<StorkState>(@stork).stork_evm_public_key
    }

    /// Returns the fee for a single update
    public fun get_single_update_fee(): u64 acquires StorkState {
        borrow_global<StorkState>(@stork).single_update_fee
    }

    /// Returns the address of the Stork contract
    public fun get_stork_address(): address acquires StorkState {
        borrow_global<StorkState>(@stork).stork_address
    }

    public fun state_exists(): bool {
        exists<StorkState>(@stork)
    }

    /// === Admin Functions ===

    /// Sets the owner of the Stork contract
    public fun set_owner(signer: &signer, new_owner: address) acquires StorkState {
        let state = borrow_global_mut<StorkState>(@stork);
        assert!(
            signer::address_of(signer) == state.owner,
            E_NOT_OWNER
        );
        state.owner = new_owner;
    }

    /// Sets the fee for a single update
    public fun set_single_update_fee(signer: &signer, new_fee: u64) acquires StorkState {
        let state = borrow_global_mut<StorkState>(@stork);
        assert!(
            signer::address_of(signer) == state.owner,
            E_NOT_OWNER
        );
        state.single_update_fee = new_fee;
    }

    // === Test Imports ===

    #[test_only]
    use aptos_framework::account::create_account_for_test;
    #[test_only]
    use stork::evm_pubkey;
    
    // === Test Constants ===

    #[test_only]
    const OWNER: address = @stork;
    #[test_only]
    const NEW_OWNER: address = @0xFACE;

    // === Tests ===

    #[test]
    fun test_state_initialization() acquires StorkState {
        let owner = create_account_for_test(OWNER);
        let pubkey = evm_pubkey::create_zeroed_evm_pubkey();
        let fee = 100;
        
        let state = new(pubkey, fee, OWNER);
        move_state(state, &owner);
        
        assert!(state_exists(), 0);
        assert!(get_single_update_fee() == fee, 1);
        assert!(get_stork_address() == @stork, 2);
        assert!(get_stork_evm_public_key() == pubkey, 3);
    }

    #[test]
    fun test_set_owner() acquires StorkState {
        let owner = create_account_for_test(OWNER);
        let pubkey = evm_pubkey::create_zeroed_evm_pubkey();
        let state = new(pubkey, 100, OWNER);
        move_state(state, &owner);
        
        set_owner(&owner, NEW_OWNER);
    }

    #[test]
    fun test_set_single_update_fee() acquires StorkState {
        let owner = create_account_for_test(OWNER);
        let pubkey = evm_pubkey::create_zeroed_evm_pubkey();
        let initial_fee = 100;
        let state = new(pubkey, initial_fee, OWNER);
        move_state(state, &owner);
        
        let new_fee = 200;
        set_single_update_fee(&owner, new_fee);
        
        assert!(get_single_update_fee() == new_fee, 0);
    }

    #[test]
    #[expected_failure(abort_code = E_NOT_OWNER)]
    fun test_set_owner_unauthorized() acquires StorkState {
        let owner = create_account_for_test(OWNER);
        let non_owner = create_account_for_test(NEW_OWNER);
        let pubkey = evm_pubkey::create_zeroed_evm_pubkey();
        let state = new(pubkey, 100, OWNER);
        move_state(state, &owner);
        
        set_owner(&non_owner, NEW_OWNER);
    }

    #[test]
    #[expected_failure(abort_code = E_NOT_OWNER)]
    fun test_set_fee_unauthorized() acquires StorkState {
        let owner = create_account_for_test(OWNER);
        let non_owner = create_account_for_test(NEW_OWNER);
        let pubkey = evm_pubkey::create_zeroed_evm_pubkey();
        let state = new(pubkey, 100, OWNER);
        move_state(state, &owner);
        
        set_single_update_fee(&non_owner, 200);
    }
}