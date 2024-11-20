module stork::event {

    // === Imports ===

    use sui::event;

    // === Structs ===

    public struct StorkInitializationEvent has copy, drop, store {
        stork_sui_public_key: address,
        stork_evm_public_key: vector<u8>,
        single_update_fee: u64,
        stork_state_id: ID,
        stork_state_version: u64,
    }

    public struct FeeWithdrawalEvent has copy, drop{
        asset_id: vector<u8>,
        amount: u64,
    }

    public struct PriceFeedUpdateEvent {}

    // === Functions ===

    public(package) fun emit_stork_initialization_event(
        stork_sui_public_key: address,
        stork_evm_public_key: vector<u8>,
        single_update_fee: u64,
        stork_state_id: ID,
        stork_state_version: u64,
    ) {
        event::emit(
            StorkInitializationEvent {
                stork_sui_public_key,
                stork_evm_public_key,
                single_update_fee,
                stork_state_id,
                stork_state_version,
            }
        );
    }

    public(package) fun emit_fee_withdrawal_event(
        asset_id: vector<u8>,
        amount: u64,
    ) {
        event::emit(FeeWithdrawalEvent { asset_id, amount });
    }
}