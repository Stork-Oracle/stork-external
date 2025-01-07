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
}