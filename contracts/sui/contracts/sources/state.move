module stork::state {
    use stork::stork::AdminCapability;

    public struct StorkState has key, store {
        id: UID,
        // the address of the stork program
        stork_sui_public_key: address,
        // Storks EVM public key
        stork_evm_public_key: vector<u8>,
        // the fee to update a value
        single_update_fee: u64,
        // the capability to perform admin actions
        admin_capability: AdminCapability,
    }
}
